package browser

import (
	"context"
	"errors"
	"fmt"

	"github.com/Roti18/siakad-war-bot/internal/domain"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type RodDriver struct {
	browser *rod.Browser
	page    *rod.Page
}

func NewBrowserDriver() domain.BrowserDriver {
	return &RodDriver{}
}

func (d *RodDriver) Launch(ctx context.Context, profileMode string, blockPolicy domain.BlockPolicy) error {
	l := launcher.New()
	
	// Server/Client optimization flags
	l.Set("disable-background-networking").
		Set("disable-extensions").
		Set("disable-translate").
		Set("disable-sync").
		Set("disable-notifications").
		Set("disable-default-apps").
		Set("disable-features", "TranslateUI,AutofillServerCommunication").
		Set("disable-popup-blocking").
		Set("metrics-recording-only").
		Set("no-first-run").
		Set("no-default-browser-check").
		Set("disable-background-timer-throttling").
		Set("disable-renderer-backgrounding").
		Set("disable-gpu").
		Set("disable-software-rasterizer")

	// Profile selection
	if profileMode == "Persistent" {
		l.UserDataDir("./chrome_profile")
	}

	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch Chrome: %w", err)
	}

	d.browser = rod.New().ControlURL(url).MustConnect()
	
	if profileMode == "Incognito" {
		incognitoBrowser, err := d.browser.Incognito()
		if err != nil {
			return fmt.Errorf("failed to create incognito context: %w", err)
		}
		d.page = incognitoBrowser.MustPage()
	} else {
		d.page = d.browser.MustPage()
	}

	// High Performance CDP Request Interception (Asset Blocker)
	if blockPolicy.CSS || blockPolicy.Image || blockPolicy.Font || blockPolicy.Media {
		router := d.page.HijackRequests()
		router.MustAdd("*", func(ctx *rod.Hijack) {
			resType := ctx.Request.Type()
			if (blockPolicy.CSS && resType == proto.NetworkResourceTypeStylesheet) ||
				(blockPolicy.Image && resType == proto.NetworkResourceTypeImage) ||
				(blockPolicy.Font && resType == proto.NetworkResourceTypeFont) ||
				(blockPolicy.Media && resType == proto.NetworkResourceTypeMedia) {
				ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
				return
			}
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
		go router.Run()
	}

	return nil
}

func (d *RodDriver) Navigate(ctx context.Context, url string) error {
	if d.page == nil {
		return errors.New("browser page not initialized")
	}
	return d.page.Context(ctx).Navigate(url)
}

func (d *RodDriver) CurrentURL(ctx context.Context) (string, error) {
	if d.page == nil {
		return "", errors.New("browser page not initialized")
	}
	info, err := d.page.Context(ctx).Info()
	if err != nil {
		return "", err
	}
	return info.URL, nil
}

func (d *RodDriver) Refresh(ctx context.Context) error {
	if d.page == nil {
		return errors.New("browser page not initialized")
	}
	_, err := d.page.Context(ctx).Eval("location.reload()")
	return err
}

func (d *RodDriver) SwitchFrame(ctx context.Context, name string) (bool, error) {
	// Frame-switching logic di Rod dicapai secara langsung saat query DOM.
	// Kita sediakan hook stub agar mematuhi interface BrowserDriver.
	return true, nil
}

func (d *RodDriver) SwitchToDefault(ctx context.Context) error {
	return nil
}

func (d *RodDriver) EvaluateJS(ctx context.Context, script string, args ...interface{}) (interface{}, error) {
	if d.page == nil {
		return nil, errors.New("browser page not initialized")
	}
	res, err := d.page.Context(ctx).Eval(script, args...)
	if err != nil {
		return nil, err
	}
	return res.Value.Val(), nil
}

func (d *RodDriver) TakeScreenshot(ctx context.Context) ([]byte, error) {
	if d.page == nil {
		return nil, errors.New("browser page not initialized")
	}
	return d.page.Context(ctx).Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})
}

func (d *RodDriver) Close() error {
	if d.browser != nil {
		return d.browser.Close()
	}
	return nil
}
