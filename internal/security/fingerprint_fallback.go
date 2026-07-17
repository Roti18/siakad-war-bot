//go:build !windows
package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

func GetDeviceFingerprint() (string, error) {
	hostname, _ := os.Hostname()
	cpuInfo := "POSIX-CPU-DEV"
	diskInfo := "POSIX-DISK-DEV"
	osVer := "POSIX-OS-DEV"
	machineGUID := "POSIX-GUID-DEV"

	rawFingerprint := fmt.Sprintf("%s|%s|%s|%s|%s", machineGUID, hostname, cpuInfo, diskInfo, osVer)
	hash := sha256.Sum256([]byte(rawFingerprint))
	return hex.EncodeToString(hash[:]), nil
}
