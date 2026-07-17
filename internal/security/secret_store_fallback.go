//go:build !windows
package security

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/Roti18/siakad-war-bot/internal/domain"
)

type ObfuscatedStore struct {
	mu       sync.Mutex
	filePath string
}

func NewSecretStore(filePath string) domain.SecretStore {
	if filePath == "" {
		filePath = "credentials.dat"
	}
	return &ObfuscatedStore{filePath: filePath}
}

func (s *ObfuscatedStore) loadMap() (map[string]string, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	base64Data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	decryptedBytes, err := base64.StdEncoding.DecodeString(string(base64Data))
	if err != nil {
		return nil, err
	}

	var m map[string]string
	err = json.Unmarshal(decryptedBytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *ObfuscatedStore) saveMap(m map[string]string) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	obfuscated := base64.StdEncoding.EncodeToString(jsonBytes)
	return os.WriteFile(s.filePath, []byte(obfuscated), 0600)
}

func (s *ObfuscatedStore) Save(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, err := s.loadMap()
	if err != nil {
		return err
	}

	m[key] = string(value)
	return s.saveMap(m)
}

func (s *ObfuscatedStore) Load(ctx context.Context, key string) ([]byte, error) {
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

func (s *ObfuscatedStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, err := s.loadMap()
	if err != nil {
		return err
	}

	delete(m, key)
	return s.saveMap(m)
}
