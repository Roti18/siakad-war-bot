package validation

import (
	"errors"
	"strings"
	
	"github.com/Roti18/siakad-war-bot/shared/dto"
)

func ValidateLicenseVerifyRequest(req *dto.LicenseVerifyRequest) error {
	if strings.TrimSpace(req.LicenseKey) == "" {
		return errors.New("license key cannot be empty")
	}
	if strings.TrimSpace(req.DeviceFingerprint) == "" {
		return errors.New("device fingerprint cannot be empty")
	}
	if strings.TrimSpace(req.Hostname) == "" {
		return errors.New("hostname cannot be empty")
	}
	if strings.TrimSpace(req.OsVersion) == "" {
		return errors.New("os version cannot be empty")
	}
	if strings.TrimSpace(req.AppVersion) == "" {
		return errors.New("app version cannot be empty")
	}
	return nil
}


