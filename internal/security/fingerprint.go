//go:build windows
package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func GetDeviceFingerprint() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	var machineGUID string
	if err == nil {
		machineGUID, _, err = k.GetStringValue("MachineGuid")
		k.Close()
	}
	if err != nil {
		machineGUID = "GUID-FALLBACK-12345"
	}

	hostname, _ := os.Hostname()

	cpuID := os.Getenv("PROCESSOR_IDENTIFIER")
	cpuArch := os.Getenv("PROCESSOR_ARCHITECTURE")
	cpuCores := os.Getenv("NUMBER_OF_PROCESSORS")
	cpuInfo := fmt.Sprintf("%s-%s-%s", cpuID, cpuArch, cpuCores)

	var volumeSerialNumber uint32
	cDrive, _ := syscall.UTF16PtrFromString("C:\\")
	err = windows.GetVolumeInformation(
		cDrive,
		nil, 0,
		&volumeSerialNumber,
		nil, nil,
		nil, 0,
	)
	diskInfo := fmt.Sprintf("%X", volumeSerialNumber)

	kNT, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	var osVer string
	if err == nil {
		prodName, _, _ := kNT.GetStringValue("ProductName")
		buildNum, _, _ := kNT.GetStringValue("CurrentBuild")
		osVer = fmt.Sprintf("%s (Build %s)", prodName, buildNum)
		kNT.Close()
	} else {
		osVer = "Windows Fallback"
	}

	rawFingerprint := fmt.Sprintf("%s|%s|%s|%s|%s", machineGUID, hostname, cpuInfo, diskInfo, osVer)
	hash := sha256.Sum256([]byte(rawFingerprint))
	return hex.EncodeToString(hash[:]), nil
}
