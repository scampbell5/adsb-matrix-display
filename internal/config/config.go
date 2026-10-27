package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DisplayName    string
	HTTPAddr       string
	PiAwareURL     string
	ReceiverLat    float64
	ReceiverLon    float64
	PollInterval   time.Duration
	PiAwareTimeout time.Duration
	MaxPositionAge time.Duration
	MaxDistanceNM  float64
	MinAltitudeFT  int
	IncludeGround  bool

	WF2Host      string
	WF2Port      int
	WF2Timeout   time.Duration
	PushEnabled  bool
	PushInterval time.Duration
	ForceRefresh time.Duration
	Brightness   int

	APIToken string
}

func Load() (Config, error) {
	var c Config
	c.DisplayName = env("DISPLAY_NAME", "ADS-B Matrix")
	c.HTTPAddr = env("HTTP_ADDR", ":8080")
	c.PiAwareURL = env("PIAWARE_URL", "http://piaware:8080/data/aircraft.json")

	var err error
	if c.ReceiverLat, err = requiredFloat("RECEIVER_LAT"); err != nil {
		return c, err
	}
	if c.ReceiverLon, err = requiredFloat("RECEIVER_LON"); err != nil {
		return c, err
	}
	if c.ReceiverLat < -90 || c.ReceiverLat > 90 {
		return c, fmt.Errorf("RECEIVER_LAT out of range")
	}
	if c.ReceiverLon < -180 || c.ReceiverLon > 180 {
		return c, fmt.Errorf("RECEIVER_LON out of range")
	}

	c.PollInterval = seconds("POLL_INTERVAL_SECONDS", 2)
	c.PiAwareTimeout = seconds("PIAWARE_TIMEOUT_SECONDS", 4)
	c.MaxPositionAge = seconds("MAX_POSITION_AGE_SECONDS", 15)
	c.MaxDistanceNM = floatEnv("MAX_DISTANCE_NM", 150)
	c.MinAltitudeFT = intEnv("MIN_ALTITUDE_FT", 0)
	c.IncludeGround = boolEnv("INCLUDE_GROUND", false)

	c.WF2Host = strings.TrimSpace(os.Getenv("WF2_HOST"))
	c.WF2Port = intEnv("WF2_PORT", 10001)
	c.WF2Timeout = seconds("WF2_TIMEOUT_SECONDS", 8)
	c.PushEnabled = boolEnv("PUSH_ENABLED", false)
	c.PushInterval = seconds("PUSH_INTERVAL_SECONDS", 10)
	c.ForceRefresh = seconds("FORCE_REFRESH_SECONDS", 300)
	c.Brightness = intEnv("WF2_BRIGHTNESS", 35)
	if c.Brightness < 1 || c.Brightness > 100 {
		return c, fmt.Errorf("WF2_BRIGHTNESS must be 1-100")
	}
	if c.PushEnabled && c.WF2Host == "" {
		return c, fmt.Errorf("PUSH_ENABLED=true requires WF2_HOST")
	}

	c.APIToken = strings.TrimSpace(os.Getenv("API_TOKEN"))
	return c, nil
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func requiredFloat(key string) (float64, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0, fmt.Errorf("%s is required", key)
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return n, nil
}

func floatEnv(key string, def float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
}
func intEnv(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
func boolEnv(key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
func seconds(key string, def int) time.Duration { return time.Duration(intEnv(key, def)) * time.Second }
