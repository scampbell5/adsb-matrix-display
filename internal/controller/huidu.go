package controller

import (
	"fmt"
	"sync"
	"time"

	"example.com/adsb-matrix-display/internal/skywatch"
	huidu "github.com/alparslanahmed/huidu-led"
)

type Status struct {
	Configured   bool   `json:"configured"`
	Connected    bool   `json:"connected"`
	Host         string `json:"host,omitempty"`
	Port         int    `json:"port,omitempty"`
	Model        string `json:"model,omitempty"`
	DeviceID     string `json:"device_id,omitempty"`
	ScreenWidth  int    `json:"screen_width,omitempty"`
	ScreenHeight int    `json:"screen_height,omitempty"`
	Firmware     string `json:"firmware,omitempty"`
}

type Huidu struct {
	mu          sync.Mutex
	host        string
	port        int
	timeout     time.Duration
	displayName string
	device      *huidu.Device
	status      Status
}

func New(host string, port int, timeout time.Duration, displayName string) *Huidu {
	if displayName == "" {
		displayName = "ADS-B Matrix"
	}
	return &Huidu{host: host, port: port, timeout: timeout, displayName: displayName, status: Status{Configured: host != "", Host: host, Port: port}}
}

func (h *Huidu) Status() Status {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.status
	if h.device != nil {
		s.Connected = h.device.IsConnected()
	}
	return s
}

func (h *Huidu) ensureConnectedLocked() error {
	if h.host == "" {
		return fmt.Errorf("WF2_HOST is not configured")
	}
	if h.device != nil && h.device.IsConnected() {
		return nil
	}

	if h.device != nil {
		_ = h.device.Close()
		h.device = nil
	}
	d := huidu.NewDevice(h.host, h.port,
		huidu.WithTimeout(h.timeout),
		huidu.WithHeartbeatInterval(30*time.Second),
		huidu.WithAutoReconnect(true),
	)
	if err := d.Connect(); err != nil {
		h.status.Connected = false
		return err
	}
	info := d.CachedDeviceInfo()
	h.device = d
	h.status.Connected = true
	h.status.Model = info.Model
	h.status.DeviceID = info.DeviceID
	h.status.ScreenWidth = info.ScreenWidth
	h.status.ScreenHeight = info.ScreenHeight
	h.status.Firmware = info.AppVersion
	return nil
}

func (h *Huidu) resetLocked() {
	if h.device != nil {
		_ = h.device.Close()
	}
	h.device = nil
	h.status.Connected = false
}

func (h *Huidu) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.resetLocked()
}

func (h *Huidu) Push(model skywatch.DisplayModel) error {
	frame, err := skywatch.RenderFramePNG(model)
	if err != nil {
		return fmt.Errorf("render frame: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureConnectedLocked(); err != nil {
		return err
	}

	const frameName = "adsb-matrix-frame.png"
	if err := h.device.UploadFileData(frameName, frame, huidu.FileTypeImage); err != nil {
		h.resetLocked()
		return fmt.Errorf("upload frame: %w", err)
	}

	screen := huidu.NewScreen()
	prog := screen.AddProgramWithConfig(huidu.ProgramConfig{Name: h.displayName, Realtime: true})
	area := prog.AddArea(0, 0, skywatch.FrameWidth, skywatch.FrameHeight)
	area.AddImage(frameName, huidu.ImageConfig{Fit: huidu.ImageFitStretch})

	if err := h.device.SendScreen(screen); err != nil {
		h.resetLocked()
		return err
	}
	return nil
}

func addText(area *huidu.Area, text string, fontSize int, color string, bold bool, align huidu.HAlign) {
	if text == "" {
		text = " "
	}
	area.AddText(text, huidu.TextConfig{
		FontSize:        fontSize,
		Bold:            bold,
		Color:           color,
		HAlign:          align,
		VAlign:          huidu.VAlignMiddle,
		BackgroundColor: "#000000",
	})
}

func (h *Huidu) TestText(text string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureConnectedLocked(); err != nil {
		return err
	}
	if err := h.device.SendText(text, huidu.TextConfig{Color: huidu.ColorCyan, FontSize: 10, Bold: true, HAlign: huidu.HAlignCenter, VAlign: huidu.VAlignMiddle, BackgroundColor: "#000000"}); err != nil {
		h.resetLocked()
		return err
	}
	return nil
}

func (h *Huidu) SetBrightness(value int) error {
	if value < 1 || value > 100 {
		return fmt.Errorf("brightness must be 1-100")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureConnectedLocked(); err != nil {
		return err
	}
	if err := h.device.SetBrightness(value); err != nil {
		h.resetLocked()
		return err
	}
	return nil
}

func (h *Huidu) SetScreen(on bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureConnectedLocked(); err != nil {
		return err
	}
	var err error
	if on {
		err = h.device.OpenScreen()
	} else {
		err = h.device.CloseScreen()
	}
	if err != nil {
		h.resetLocked()
	}
	return err
}
