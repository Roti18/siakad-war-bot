package browser

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"github.com/Roti18/siakad-war-bot/internal/domain"
)

type screenshotTask struct {
	filename string
	data     []byte
}

type AsyncScreenshotService struct {
	saveDir    string
	queue      chan screenshotTask
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

// NewScreenshotService membuat instance baru dari ScreenshotService
func NewScreenshotService(saveDir string) domain.ScreenshotService {
	if saveDir == "" {
		saveDir = "warResult"
	}
	return &AsyncScreenshotService{
		saveDir: saveDir,
		queue:   make(chan screenshotTask, 100),
	}
}

// Start menjalankan worker pool asinkron untuk menyimpan screenshot
func (s *AsyncScreenshotService) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	// Buat subfolder jika belum ada
	_ = os.MkdirAll(filepath.Join(s.saveDir, "success"), 0755)
	_ = os.MkdirAll(filepath.Join(s.saveDir, "failed"), 0755)

	// Jalankan worker goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case task, ok := <-s.queue:
				if !ok {
					return
				}
				s.saveToFile(task.filename, task.data)
			case <-ctx.Done():
				// Drain queue yang tersisa sebelum shutdown
				for task := range s.queue {
					s.saveToFile(task.filename, task.data)
				}
				return
			}
		}
	}()
}

// QueueScreenshot memasukkan task screenshot ke dalam antrean
func (s *AsyncScreenshotService) QueueScreenshot(ctx context.Context, filename string, data []byte) {
	select {
	case s.queue <- screenshotTask{filename: filename, data: data}:
	default:
		// Queue penuh, drop task untuk menghindari blocker
	}
}

// Stop menghentikan worker secara graceful dan menunggu semua antrean selesai diproses
func (s *AsyncScreenshotService) Stop() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	close(s.queue)
	s.wg.Wait()
}

func (s *AsyncScreenshotService) saveToFile(filename string, data []byte) {
	fullPath := filepath.Join(s.saveDir, filename)
	
	// Pastikan folder induk ada
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
	
	_ = os.WriteFile(fullPath, data, 0644)
}
