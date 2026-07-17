package domain

import (
	"context"
)

type TargetCourse struct {
	Nama  string `json:"nama"`
	Kelas string `json:"kelas"`
}

type BlockPolicy struct {
	CSS   bool
	Image bool
	Font  bool
	Media bool
}

// BrowserDriver mendefinisikan plugin abstraksi otomasi browser
type BrowserDriver interface {
	Launch(ctx context.Context, profileMode string, blockPolicy BlockPolicy) error
	Navigate(ctx context.Context, url string) error
	CurrentURL(ctx context.Context) (string, error)
	Refresh(ctx context.Context) error
	SwitchFrame(ctx context.Context, name string) (bool, error)
	SwitchToDefault(ctx context.Context) error
	EvaluateJS(ctx context.Context, script string, args ...interface{}) (interface{}, error)
	TakeScreenshot(ctx context.Context) ([]byte, error)
	Close() error
}

// SecretStore mengabstraksi enkripsi password/token di level OS
type SecretStore interface {
	Save(ctx context.Context, key string, value []byte) error
	Load(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// ScreenshotService mengelola worker pool asinkron penyimpanan file png
type ScreenshotService interface {
	Start(ctx context.Context)
	QueueScreenshot(ctx context.Context, filename string, data []byte)
	Stop()
}
