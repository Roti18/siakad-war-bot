package dto

import "time"

type LicenseVerifyRequest struct {
	LicenseKey        string `json:"license_key"`
	DeviceFingerprint string `json:"device_fingerprint"` // SHA-256 Hash
	Hostname          string `json:"hostname"`
	OsVersion         string `json:"os_version"`
	AppVersion        string `json:"app_version"`
}

type LicenseVerifyResponse struct {
	Token      string    `json:"token"`
	LicenseKey string    `json:"license_key"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsLifetime bool      `json:"is_lifetime"`
	IsTrial    bool      `json:"is_trial"`
	Status     string    `json:"status"`
}


