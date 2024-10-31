package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	StateFile                   string  `json:"state_file"`
	InfluxServer                string  `json:"influx_server"`
	InfluxOrg                   string  `json:"influx_org,omitempty"`
	InfluxUser                  string  `json:"influx_user,omitempty"`
	InfluxPass                  string  `json:"influx_password,omitempty"`
	InfluxToken                 string  `json:"influx_token,omitempty"`
	InfluxHealthCheckDisabled   bool    `json:"influx_health_check_disabled"`
	InfluxTimeoutS              int     `json:"influx_timeout_s"`
	PowerMeanRunningThreshold   float64 `json:"power_mean_running_threshold"`
	PriorWindowPowerMeanQuery   string  `json:"prior_window_power_mean_query"`
	CurrentWindowPowerMeanQuery string  `json:"current_window_power_mean_query"`
	NotifyEveryMinutes          int     `json:"notify_every_minutes"`
	APIPort                     int     `json:"api_port"`
	APIRoot                     string  `json:"api_root"`
	NtfyServer                  string  `json:"ntfy_server"`
	NtfyToken                   string  `json:"ntfy_token"`
	NtfyTopic                   string  `json:"ntfy_topic"`
	NtfyTimeoutS                int     `json:"ntfy_timeout_s"`
	NtfyTagsStr                 string  `json:"ntfy_tags"`
	NtfyPriority                int     `json:"ntfy_priority"`
}

func ConfigFromFile(filename string) (*Config, error) {
	config := Config{}
	cfgBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filename, err)
	}
	if err = json.Unmarshal(cfgBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file '%s': %w", filename, err)
	}
	config.SetDefaults()
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) SetDefaults() {
	if c.PowerMeanRunningThreshold == 0 {
		c.PowerMeanRunningThreshold = 5
	}
	if c.InfluxTimeoutS == 0 {
		c.InfluxTimeoutS = 10
	}
	if c.NotifyEveryMinutes == 0 {
		c.NotifyEveryMinutes = 30
	}
	if c.APIPort == 0 {
		c.APIPort = 8080
	}
	if c.APIRoot == "" {
		c.APIRoot = "http://localhost:8080"
	}
	if strings.HasSuffix(c.APIRoot, "/") {
		c.APIRoot = c.APIRoot[:len(c.APIRoot)-1]
	}
	if c.NtfyTimeoutS == 0 {
		c.NtfyTimeoutS = 10
	}
	if c.NtfyPriority == 0 {
		c.NtfyPriority = 3
	}
}

func (c *Config) Validate() error {
	if c.InfluxServer == "" {
		return fmt.Errorf("influx_server is required")
	}
	if c.PriorWindowPowerMeanQuery == "" {
		return fmt.Errorf("power_old_query is required")
	}
	if c.CurrentWindowPowerMeanQuery == "" {
		return fmt.Errorf("power_new_query is required")
	}
	if c.NtfyServer == "" {
		return fmt.Errorf("ntfy_server is required")
	}
	if _, err := url.Parse(c.NtfyServer); err != nil {
		return fmt.Errorf("ntfy_server is not a valid URL")
	}
	if c.NtfyTopic == "" {
		return fmt.Errorf("ntfy_topic is required")
	}
	if c.NtfyPriority < 1 || c.NtfyPriority > 5 {
		return fmt.Errorf("ntfy_priority must be between 1 and 5 (inclusive)")
	}
	if _, err := url.Parse(c.APIRoot); err != nil {
		return fmt.Errorf("api_root is not a valid URL")
	}
	return nil
}

func (c *Config) InfluxTimeout() time.Duration {
	return time.Duration(c.InfluxTimeoutS) * time.Second
}

func (c *Config) NotifyEvery() time.Duration {
	return time.Duration(c.NotifyEveryMinutes) * time.Minute
}

func (c *Config) NtfyTimeout() time.Duration {
	return time.Duration(c.NtfyTimeoutS) * time.Second
}

func (c *Config) NtfyServerURL() *url.URL {
	retv, err := url.Parse(c.NtfyServer)
	if err != nil {
		panic("bad ntfy server URL")
	}
	return retv
}

func (c *Config) NtfyTags() []string {
	return strings.Split(c.NtfyTagsStr, ",")
}
