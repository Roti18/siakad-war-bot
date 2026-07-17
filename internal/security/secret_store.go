//go:build windows
package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/Roti18/siakad-war-bot/internal/domain"
)

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

var (
	dllCrypt32             = syscall.NewLazyDLL("crypt32.dll")
	procCryptProtectData   = dllCrypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = dllCrypt32.NewProc("CryptUnprotectData")
)

type DPAPIStore struct {
	mu       sync.Mutex
	filePath string
}

func NewSecretStore(filePath string) domain.SecretStore {
	if filePath == "" {
		filePath = "credentials.dat"
	}
	return &DPAPIStore{filePath: filePath}
}

func encrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	var in DATA_BLOB
	in.cbData = uint32(len(data))
	in.pbData = &data[0]

	var out DATA_BLOB
	r1, _, err := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(&in)),
		0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(&out)),
	)
	if r1 == 0 {
		return nil, err
	}
	defer syscall.LocalFree(syscall.Handle(unsafe.Pointer(out.pbData)))

	result := make([]byte, out.cbData)
	copy(result, (*[1 << 30]byte)(unsafe.Pointer(out.pbData))[:out.cbData:out.cbData])
	return result, nil
}

func decrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	var in DATA_BLOB
	in.cbData = uint32(len(data))
	in.pbData = &data[0]

	var out DATA_BLOB
	r1, _, err := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(&in)),
		0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(&out)),
	)
	if r1 == 0 {
		return nil, err
	}
	defer syscall.LocalFree(syscall.Handle(unsafe.Pointer(out.pbData)))

	result := make([]byte, out.cbData)
	copy(result, (*[1 << 30]byte)(unsafe.Pointer(out.pbData))[:out.cbData:out.cbData])
	return result, nil
}

func (s *DPAPIStore) loadMap() (map[string]string, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	encryptedData, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	decryptedBytes, err := decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	var m map[string]string
	err = json.Unmarshal(decryptedBytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *DPAPIStore) saveMap(m map[string]string) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	encryptedBytes, err := encrypt(jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	return os.WriteFile(s.filePath, encryptedBytes, 0600)
}

func (s *DPAPIStore) Save(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, err := s.loadMap()
	if err != nil {
		return err
	}

	m[key] = string(value)
	return s.saveMap(m)
}

func (s *DPAPIStore) Load(ctx context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, err := s.loadMap()
	if err != nil {
		return nil, err
	}

	val, ok := m[key]
	if !ok {
		return nil, errors.New("key not found in secret store")
	}
	return []byte(val), nil
}

func (s *DPAPIStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, err := s.loadMap()
	if err != nil {
		return err
	}

	delete(m, key)
	return s.saveMap(m)
}
