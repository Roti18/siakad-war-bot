package config

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/Roti18/siakad-war-bot/internal/domain"
)

type Config struct {
	mu         sync.RWMutex
	filePath   string
	Browser    BrowserConfig         `json:"browser"`
	Schedule   ScheduleConfig        `json:"schedule"`
	Network    NetworkConfig         `json:"network"`
	Screenshot ScreenshotConfig      `json:"screenshot"`
	Courses    []domain.TargetCourse `json:"courses"`
}

type BrowserConfig struct {
	Headless            bool `json:"headless"`
	Width               int  `json:"width"`
	Height              int  `json:"height"`
	BlockImages         bool `json:"block_images"`
	BlockCSS            bool `json:"block_css"`
	BlockFonts          bool `json:"block_fonts"`
	BlockMedia          bool `json:"block_media"`
	DisableGPU          bool `json:"disable_gpu"`
	DisableExtension    bool `json:"disable_extension"`
	DisableNotification bool `json:"disable_notification"`
	EnableIncognito     bool `json:"enable_incognito"`
}

type ScheduleConfig struct {
	Date                 string  `json:"date"`
	Time                 string  `json:"time"`
	StandbyBeforeMinutes int     `json:"standby_before_minutes"`
	RefreshIntervalSec   int     `json:"refresh_interval_seconds"`
	RetryDelaySec        float64 `json:"retry_delay_seconds"`
	MaxRetry             int     `json:"max_retry"`
}

type NetworkConfig struct {
	Proxy             string  `json:"proxy"`
	HTTPTimeoutSec    int     `json:"http_timeout_seconds"`
	ConnTimeoutSec    int     `json:"connection_timeout_seconds"`
	RetryCount        int     `json:"retry_count"`
	RetryDelaySeconds float64 `json:"retry_delay_seconds"`
}

type ScreenshotConfig struct {
	Enable                 bool   `json:"enable"`
	SaveDirectory          string `json:"save_directory"`
	ScreenshotSuccess      bool   `json:"screenshot_success"`
	ScreenshotFailed       bool   `json:"screenshot_failed"`
	ScreenshotBeforeSubmit bool   `json:"screenshot_before_submit"`
	ScreenshotAfterSubmit  bool   `json:"screenshot_after_submit"`
}

func NewConfigManager(filePath string) *Config {
	if filePath == "" {
		filePath = "config.json"
	}
	return &Config{
		filePath: filePath,
	}
}

func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(c.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.filePath, data, 0644)
}

func (c *Config) Reset() error {
	c.mu.Lock()
	c.Browser = BrowserConfig{
		Headless:            false,
		Width:               1920,
		Height:              1080,
		BlockImages:         true,
		BlockCSS:            true,
		BlockFonts:          true,
		BlockMedia:          true,
		DisableGPU:          true,
		DisableExtension:    true,
		DisableNotification: true,
		EnableIncognito:     true,
	}
	c.Schedule = ScheduleConfig{
		Date:                 "",
		Time:                 "09:00:00",
		StandbyBeforeMinutes: 10,
		RefreshIntervalSec:   300,
		RetryDelaySec:        0.5,
		MaxRetry:             100,
	}
	c.Network = NetworkConfig{
		Proxy:             "",
		HTTPTimeoutSec:    30,
		ConnTimeoutSec:    10,
		RetryCount:        3,
		RetryDelaySeconds: 2.0,
	}
	c.Screenshot = ScreenshotConfig{
		Enable:                 true,
		SaveDirectory:          "warResult",
		ScreenshotSuccess:      true,
		ScreenshotFailed:       true,
		ScreenshotBeforeSubmit: false,
		ScreenshotAfterSubmit:  true,
	}
	c.Courses = []domain.TargetCourse{}
	c.mu.Unlock()

	return c.Save()
}

func (c *Config) Export(outPath string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, data, 0644)
}

func (c *Config) Import(inPath string) error {
	file, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	temp := &Config{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}

	// Validasi data impor dasar
	if temp.Schedule.Time == "" {
		return errors.New("invalid config file: schedule time cannot be empty")
	}

	c.mu.Lock()
	c.Browser = temp.Browser
	c.Schedule = temp.Schedule
	c.Network = temp.Network
	c.Screenshot = temp.Screenshot
	c.Courses = temp.Courses
	c.mu.Unlock()

	return c.Save()
}

func (c *Config) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.Schedule.Time == "" {
		return errors.New("schedule time is empty")
	}
	if c.Screenshot.SaveDirectory == "" {
		return errors.New("screenshot directory path is empty")
	}
	return nil
}

// LoadEnv reads a .env file and sets environment variables in the OS.
func LoadEnv(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	envMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
				val = val[1 : len(val)-1]
			}
			envMap[key] = val
			os.Setenv(key, val)
		}
	}
	return envMap, scanner.Err()
}

// SaveEnv writes standard API_SERVER_URL, BASE_URL, NIM, and PASSWORD to a .env file.
func SaveEnv(filename string, apiServerURL, baseURL, nim, password string) error {
	content := fmt.Sprintf("API_SERVER_URL=%s\nBASE_URL=%s\nNIM=%s\nPASSWORD=%s\n", apiServerURL, baseURL, nim, password)
	return os.WriteFile(filename, []byte(content), 0644)
}

// TargetURLHex is injected at compile time using -ldflags.
// If not injected, the client falls back to the example URL.
var TargetURLHex string

// getDefaultURL returns the XOR-obfuscated default URL to prevent static analysis exposure.
func getDefaultURL() string {
	if TargetURLHex == "" {
		return "https://siakad.example.ac.id/"
	}
	bytes, err := hex.DecodeString(TargetURLHex)
	if err != nil {
		return "https://siakad.example.ac.id/"
	}
	key := byte(0x2A)
	for i, b := range bytes {
		bytes[i] = b ^ key
	}
	return string(bytes)
}

// SetupPrompt runs an interactive configuration wizard if credentials are not configured.
func SetupPrompt(envFile string) error {
	// First load existing env to check if they are already configured
	env, _ := LoadEnv(envFile)
	if env != nil && env["NIM"] != "" && env["PASSWORD"] != "" && env["API_SERVER_URL"] != "" {
		return nil
	}

	fmt.Println("\n==================================================")
	fmt.Println("   KRS WAR BOT CLIENT - INITIAL CONFIGURATION")
	fmt.Println("==================================================")

	reader := bufio.NewReader(os.Stdin)

	// 1. API_SERVER_URL
	defaultServerURL := "http://localhost:8080"
	if env != nil && env["API_SERVER_URL"] != "" {
		defaultServerURL = env["API_SERVER_URL"]
	}

	fmt.Printf("Default API Server URL: %s\n", defaultServerURL)
	fmt.Print("Gunakan API Server default ini? (Y/n atau tekan Enter untuk ya): ")
	useDefaultServerInput, _ := reader.ReadString('\n')
	useDefaultServerInput = strings.TrimSpace(strings.ToLower(useDefaultServerInput))

	apiServerURL := defaultServerURL
	if useDefaultServerInput != "" && useDefaultServerInput != "y" && useDefaultServerInput != "yes" {
		fmt.Print("Masukkan API Server URL Baru: ")
		customServerURLInput, _ := reader.ReadString('\n')
		customServerURLInput = strings.TrimSpace(customServerURLInput)
		if customServerURLInput != "" {
			apiServerURL = customServerURLInput
		}
	}

	// 2. BASE_URL
	defaultURL := getDefaultURL()
	if env != nil && env["BASE_URL"] != "" {
		defaultURL = env["BASE_URL"]
	}

	fmt.Printf("Default Base URL: %s\n", defaultURL)
	fmt.Print("Gunakan URL default ini? (Y/n atau tekan Enter untuk ya): ")
	useDefaultInput, _ := reader.ReadString('\n')
	useDefaultInput = strings.TrimSpace(strings.ToLower(useDefaultInput))

	baseURL := defaultURL
	if useDefaultInput != "" && useDefaultInput != "y" && useDefaultInput != "yes" {
		fmt.Print("Masukkan BASE_URL Baru: ")
		customURLInput, _ := reader.ReadString('\n')
		customURLInput = strings.TrimSpace(customURLInput)
		if customURLInput != "" {
			baseURL = customURLInput
		}
	}

	// 3. NIM
	var nim string
	for {
		fmt.Print("Masukkan NIM Anda: ")
		nimInput, _ := reader.ReadString('\n')
		nim = strings.TrimSpace(nimInput)
		if nim != "" {
			break
		}
		fmt.Println("NIM tidak boleh kosong!")
	}

	// 4. PASSWORD
	var password string
	for {
		fmt.Print("Masukkan PASSWORD Anda: ")
		pwdInput, _ := reader.ReadString('\n')
		password = strings.TrimSpace(pwdInput)
		if password != "" {
			break
		}
		fmt.Println("PASSWORD tidak boleh kosong!")
	}

	// Save to .env
	err := SaveEnv(envFile, apiServerURL, baseURL, nim, password)
	if err != nil {
		return fmt.Errorf("gagal menyimpan file .env: %w", err)
	}

	// Set env variables
	os.Setenv("API_SERVER_URL", apiServerURL)
	os.Setenv("BASE_URL", baseURL)
	os.Setenv("NIM", nim)
	os.Setenv("PASSWORD", password)

	fmt.Println("\n[✔] Konfigurasi berhasil disimpan ke", envFile)
	fmt.Println("==================================================\n")
	return nil
}

