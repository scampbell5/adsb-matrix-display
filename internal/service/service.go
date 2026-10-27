package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"adsb-matrix-display/internal/config"
	"adsb-matrix-display/internal/controller"
	"adsb-matrix-display/internal/skywatch"
)

type State struct {
	LastPollAt      *time.Time            `json:"last_poll_at,omitempty"`
	LastPollError   string                `json:"last_poll_error,omitempty"`
	SourceNow       float64               `json:"source_now,omitempty"`
	SourceMessages  int64                 `json:"source_messages,omitempty"`
	EligibleCount   int                   `json:"eligible_aircraft_count"`
	Nearest         *skywatch.Aircraft    `json:"nearest,omitempty"`
	Display         skywatch.DisplayModel `json:"display"`
	LastPushAt      *time.Time            `json:"last_push_at,omitempty"`
	LastPushError   string                `json:"last_push_error,omitempty"`
	LastFingerprint string                `json:"last_pushed_fingerprint,omitempty"`
}

type Service struct {
	cfg        config.Config
	httpClient *http.Client
	controller *controller.Huidu

	mu       sync.RWMutex
	state    State
	aircraft []skywatch.Aircraft
}

func New(cfg config.Config, ctl *controller.Huidu) *Service {
	return &Service{
		cfg:        cfg,
		controller: ctl,
		httpClient: &http.Client{Timeout: cfg.PiAwareTimeout},
		state:      State{Display: skywatch.BuildDisplayWithName(nil, cfg.DisplayName)},
	}
}

func (s *Service) Run(ctx context.Context) {
	go s.pollLoop(ctx)
	if s.cfg.PushEnabled {
		go s.pushLoop(ctx)
	}
}

func (s *Service) pollLoop(ctx context.Context) {
	s.pollOnce(ctx)
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.pollOnce(ctx)
		}
	}
}

func (s *Service) pollOnce(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.PiAwareURL, nil)
	if err != nil {
		s.recordPollError(err)
		return
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.recordPollError(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		s.recordPollError(fmt.Errorf("piaware returned %s: %s", resp.Status, strings.TrimSpace(string(b))))
		return
	}

	var dump skywatch.Dump1090Response
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&dump); err != nil {
		s.recordPollError(err)
		return
	}

	nearest, eligible := skywatch.Nearest(dump.Aircraft, skywatch.Filter{
		ReceiverLat:    s.cfg.ReceiverLat,
		ReceiverLon:    s.cfg.ReceiverLon,
		MaxPositionAge: s.cfg.MaxPositionAge,
		MaxDistanceNM:  s.cfg.MaxDistanceNM,
		MinAltitudeFT:  s.cfg.MinAltitudeFT,
		IncludeGround:  s.cfg.IncludeGround,
	})
	now := time.Now()
	display := skywatch.BuildDisplayWithName(nearest, s.cfg.DisplayName)

	s.mu.Lock()
	s.aircraft = eligible
	s.state.LastPollAt = &now
	s.state.LastPollError = ""
	s.state.SourceNow = dump.Now
	s.state.SourceMessages = dump.Messages
	s.state.EligibleCount = len(eligible)
	s.state.Nearest = nearest
	s.state.Display = display
	s.mu.Unlock()
}

func (s *Service) recordPollError(err error) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.LastPollAt = &now
	s.state.LastPollError = err.Error()
}

func (s *Service) pushLoop(ctx context.Context) {
	// Wait for the first successful poll, then evaluate on the normal cadence.
	ticker := time.NewTicker(s.cfg.PushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			model, lastFingerprint, lastPush := s.pushSnapshot()
			if model.Fingerprint == "" {
				continue
			}
			force := lastPush.IsZero() || time.Since(lastPush) >= s.cfg.ForceRefresh
			if model.Fingerprint == lastFingerprint && !force {
				continue
			}
			if err := s.push(model); err != nil {
				log.Printf("WF2 push failed: %v", err)
			}
		}
	}
}

func (s *Service) pushSnapshot() (skywatch.DisplayModel, string, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var last time.Time
	if s.state.LastPushAt != nil {
		last = *s.state.LastPushAt
	}
	return s.state.Display, s.state.LastFingerprint, last
}

func (s *Service) push(model skywatch.DisplayModel) error {
	if s.controller == nil {
		return fmt.Errorf("controller is not configured")
	}
	err := s.controller.Push(model)
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.LastPushAt = &now
	if err != nil {
		s.state.LastPushError = err.Error()
		return err
	}
	s.state.LastPushError = ""
	s.state.LastFingerprint = model.Fingerprint
	return nil
}

func (s *Service) Snapshot() State { s.mu.RLock(); defer s.mu.RUnlock(); return s.state }
func (s *Service) Aircraft() []skywatch.Aircraft {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]skywatch.Aircraft, len(s.aircraft))
	copy(out, s.aircraft)
	return out
}
func (s *Service) ControllerStatus() controller.Status {
	if s.controller == nil {
		return controller.Status{}
	}
	return s.controller.Status()
}
func (s *Service) PushNow() error {
	s.mu.RLock()
	model := s.state.Display
	s.mu.RUnlock()
	return s.push(model)
}
func (s *Service) TestText(text string) error {
	if s.controller == nil {
		return fmt.Errorf("controller is not configured")
	}
	return s.controller.TestText(text)
}
func (s *Service) SetBrightness(v int) error {
	if s.controller == nil {
		return fmt.Errorf("controller is not configured")
	}
	return s.controller.SetBrightness(v)
}
func (s *Service) SetScreen(on bool) error {
	if s.controller == nil {
		return fmt.Errorf("controller is not configured")
	}
	return s.controller.SetScreen(on)
}
