package main

import (
	"fmt"
	
	"github.com/Roti18/siakad-war-bot/internal/config"
	"github.com/Roti18/siakad-war-bot/internal/security"
)

func main() {
	// Run setup wizard if .env is not fully configured, and load env variables
	if err := config.SetupPrompt(".env"); err != nil {
		fmt.Printf("Configuration Error: %v\n", err)
		return
	}

	fmt.Println("KRS War Bot Client Skeleton")
	cfg := config.NewConfigManager("configs/config.json")
	if err := cfg.Load(); err != nil {
		_ = cfg.Reset()
	}
	
	hash, err := security.GetDeviceFingerprint()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Device Fingerprint: %s\n", hash)
}
