package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version int     `yaml:"version"`
	Server  Server  `yaml:"server"`
	DNS     DNS     `yaml:"dns"`
	Storage Storage `yaml:"storage"`
	Auth    Auth    `yaml:"auth"`
}

type Server struct {
	ListenHTTP string `yaml:"listen_http"`
	ListenDNS  string `yaml:"listen_dns"`
	BaseURL    string `yaml:"base_url"`
}

type DNS struct {
	UpstreamServers []UpstreamServer `yaml:"upstream_servers"`
	CacheTTL        int              `yaml:"cache_ttl"`
}

type UpstreamServer struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type Storage struct {
	DatabaseURL string `yaml:"database_url"`
}

type Auth struct {
	JWTSecret   string `yaml:"jwt_secret"`
	TokenExpiry int    `yaml:"token_expiry_hours"`
}

func Defaults() *Config {
	return &Config{
		Version: 1,
		Server: Server{
			ListenHTTP: ":3000",
			ListenDNS:  ":5353",
			BaseURL:    "http://localhost:3000",
		},
		DNS: DNS{
			UpstreamServers: []UpstreamServer{
				{Name: "Cloudflare", Address: "1.1.1.1:53"},
				{Name: "Quad9", Address: "9.9.9.9:53"},
			},
			CacheTTL: 300,
		},
		Storage: Storage{DatabaseURL: "postgres://openfiltr:openfiltr@localhost:5432/openfiltr?sslmode=disable"},
		Auth:    Auth{JWTSecret: "change-me-in-production", TokenExpiry: 24},
	}
}

func Load(path string) (*Config, error) {
	cfg := Defaults()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("creating config dir: %w", err)
		}
		data, _ := yaml.Marshal(cfg)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return nil, fmt.Errorf("writing default config: %w", err)
		}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}
	if s := os.Getenv("OPENFILTR_JWT_SECRET"); s != "" {
		cfg.Auth.JWTSecret = s
	}
	if s := os.Getenv("OPENFILTR_DATABASE_URL"); s != "" {
		cfg.Storage.DatabaseURL = s
	}
	return cfg, nil
}
