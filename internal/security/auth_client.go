package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Roti18/siakad-war-bot/internal/domain"
	"github.com/Roti18/siakad-war-bot/shared/constants"
	"github.com/Roti18/siakad-war-bot/shared/dto"
	"github.com/Roti18/siakad-war-bot/shared/version"
)

type AuthClient interface {
	VerifyLicense(ctx context.Context, apiServerURL, licenseKey string) (*dto.LicenseVerifyResponse, error)
	GetSavedJWT(ctx context.Context) (string, error)
	GetSavedLicenseKey(ctx context.Context) (string, error)
	ClearSession(ctx context.Context) error
}

type authClient struct {
	store domain.SecretStore
}

func NewAuthClient(store domain.SecretStore) AuthClient {
	return &authClient{store: store}
}

func (c *authClient) VerifyLicense(ctx context.Context, apiServerURL, licenseKey string) (*dto.LicenseVerifyResponse, error) {
	fingerprint, err := GetDeviceFingerprint()
	if err != nil {
		return nil, fmt.Errorf("failed to generate device fingerprint: %w", err)
	}

	hostname, _ := os.Hostname()
	osVer := "Windows Desktop"
	
	reqPayload := dto.LicenseVerifyRequest{
		LicenseKey:        licenseKey,
		DeviceFingerprint: fingerprint,
		Hostname:          hostname,
		OsVersion:         osVer,
		AppVersion:        version.AppVersion,
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s%s", apiServerURL, constants.EndpointVerifyLicense)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(constants.HeaderFingerprint, fingerprint)
	req.Header.Set(constants.HeaderLicense, licenseKey)
	req.Header.Set(constants.HeaderAppVersion, version.AppVersion)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var verifyResp dto.LicenseVerifyResponse
	err = json.NewDecoder(resp.Body).Decode(&verifyResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	err = c.store.Save(ctx, "jwt_token", []byte(verifyResp.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	err = c.store.Save(ctx, "license_key", []byte(licenseKey))
	if err != nil {
		return nil, fmt.Errorf("failed to save license key: %w", err)
	}

	return &verifyResp, nil
}

func (c *authClient) GetSavedJWT(ctx context.Context) (string, error) {
	t, err := c.store.Load(ctx, "jwt_token")
	if err != nil {
		return "", err
	}
	return string(t), nil
}

func (c *authClient) GetSavedLicenseKey(ctx context.Context) (string, error) {
	k, err := c.store.Load(ctx, "license_key")
	if err != nil {
		return "", err
	}
	return string(k), nil
}

func (c *authClient) ClearSession(ctx context.Context) error {
	_ = c.store.Delete(ctx, "jwt_token")
	return nil
}
