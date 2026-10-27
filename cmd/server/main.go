package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	assetpkg "adsb-matrix-display/internal/assets"
	"adsb-matrix-display/internal/config"
	"adsb-matrix-display/internal/controller"
	"adsb-matrix-display/internal/service"
	"adsb-matrix-display/internal/skywatch"
	"adsb-matrix-display/internal/webui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	var ctl *controller.Huidu
	if cfg.WF2Host != "" {
		ctl = controller.New(cfg.WF2Host, cfg.WF2Port, cfg.WF2Timeout)
		defer ctl.Close()
	}

	svc := service.New(cfg, ctl)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	svc.Run(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/display", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("GET /display", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(webui.DisplayHTML)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		st := svc.Snapshot()
		status := "ok"
		code := http.StatusOK
		if st.LastPollAt == nil || st.LastPollError != "" {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, map[string]any{"status": status, "piaware_error": st.LastPollError, "controller": svc.ControllerStatus()})
	})
	mux.HandleFunc("GET /api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"service": svc.Snapshot(), "controller": svc.ControllerStatus(), "auto_push_enabled": cfg.PushEnabled, "display_name": cfg.DisplayName})
	})
	mux.HandleFunc("GET /api/v1/aircraft", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"aircraft": svc.Aircraft()})
	})
	mux.HandleFunc("GET /api/v1/aircraft/nearest", func(w http.ResponseWriter, r *http.Request) {
		st := svc.Snapshot()
		if st.Nearest == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "no eligible aircraft with a recent position"})
			return
		}
		writeJSON(w, http.StatusOK, st.Nearest)
	})
	mux.HandleFunc("GET /api/v1/display/preview", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Snapshot().Display)
	})
	mux.HandleFunc("GET /api/v1/display/frame.png", func(w http.ResponseWriter, r *http.Request) {
		frame, err := skywatch.RenderFramePNG(svc.Snapshot().Display)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(frame)
	})
	mux.HandleFunc("GET /api/v1/operators", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"operators": skywatch.Operators()})
	})
	mux.HandleFunc("GET /api/v1/airlines", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"airlines": skywatch.Airlines()})
	})
	mux.HandleFunc("GET /api/v1/aircraft/military", func(w http.ResponseWriter, r *http.Request) {
		all := svc.Aircraft()
		military := make([]skywatch.Aircraft, 0)
		for _, a := range all {
			if a.Military {
				military = append(military, a)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"aircraft": military, "count": len(military)})
	})
	mux.HandleFunc("GET /assets/operators/{file}", func(w http.ResponseWriter, r *http.Request) {
		file := strings.TrimSpace(r.PathValue("file"))
		code := strings.ToUpper(strings.TrimSuffix(file, ".png"))
		if !strings.HasSuffix(strings.ToLower(file), ".png") || code == "" {
			http.NotFound(w, r)
			return
		}
		b, err := assetpkg.OperatorIcon(code)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(b)
	})
	// Legacy asset path retained for clients from the airline-only version.
	mux.HandleFunc("GET /assets/airlines/{file}", func(w http.ResponseWriter, r *http.Request) {
		file := strings.TrimSpace(r.PathValue("file"))
		code := strings.ToUpper(strings.TrimSuffix(file, ".png"))
		if !strings.HasSuffix(strings.ToLower(file), ".png") || code == "" {
			http.NotFound(w, r)
			return
		}
		b, err := assetpkg.OperatorIcon(code)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(b)
	})
	mux.HandleFunc("POST /api/v1/display/push", mutate(cfg.APIToken, func(w http.ResponseWriter, r *http.Request) {
		if err := svc.PushNow(); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "display": svc.Snapshot().Display})
	}))
	mux.HandleFunc("POST /api/v1/display/test", mutate(cfg.APIToken, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Text string `json:"text"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		body.Text = strings.TrimSpace(body.Text)
		if body.Text == "" {
			body.Text = cfg.DisplayName
		}
		if len(body.Text) > 64 {
			body.Text = body.Text[:64]
		}
		if err := svc.TestText(body.Text); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "text": body.Text})
	}))
	mux.HandleFunc("POST /api/v1/controller/brightness", mutate(cfg.APIToken, func(w http.ResponseWriter, r *http.Request) {
		value, err := strconv.Atoi(r.URL.Query().Get("value"))
		if err != nil || value < 1 || value > 100 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "value must be an integer from 1 to 100"})
			return
		}
		if err := svc.SetBrightness(value); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "brightness": value})
	}))
	mux.HandleFunc("POST /api/v1/controller/screen", mutate(cfg.APIToken, func(w http.ResponseWriter, r *http.Request) {
		state := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("state")))
		if state != "on" && state != "off" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "state must be on or off"})
			return
		}
		if err := svc.SetScreen(state == "on"); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "state": state})
	}))

	server := &http.Server{Addr: cfg.HTTPAddr, Handler: logRequests(mux), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("ADS-B Matrix Display API listening on %s", cfg.HTTPAddr)
		log.Printf("PiAware source: %s", cfg.PiAwareURL)
		if cfg.WF2Host == "" {
			log.Printf("WF2 controller not configured; API-only mode")
		} else {
			log.Printf("WF2 controller: %s:%d (auto push=%v)", cfg.WF2Host, cfg.WF2Port, cfg.PushEnabled)
		}
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
}

func mutate(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token != "" {
			got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if got != token {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
				return
			}
		}
		next(w, r)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
