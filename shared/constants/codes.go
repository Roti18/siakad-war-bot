package constants

const (
	HeaderFingerprint = "X-Device-Fingerprint"
	HeaderLicense     = "X-License-Key"
	HeaderAppVersion  = "X-App-Version"
	
	// API Endpoints
	EndpointVerifyLicense = "/api/license/verify"
	EndpointAdminLogin    = "/api/admin/login"
	EndpointAdminLicenses = "/api/admin/licenses"
	EndpointAdminUsers    = "/api/admin/users"
	EndpointHealth        = "/health"
)
