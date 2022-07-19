package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

const DefaultFileName = "rsnap-config.toml"

type (
	Config struct {
		JiraCrd  *JiraCrd  `toml:"jira"`
		Projects *Projects `toml:"projects"`
		Epic     *Epic     `toml:"epic"`
	}

	Projects struct {
		Names []string `toml:"names"`
	}

	JiraCrd struct {
		User      string `toml:"user"`
		AccountID string `toml:"account_id"`
		BaseURL   string `toml:"base_url"`
		Token     string `toml:"token"`
	}

	Epic struct {
		CustomFieldStartDate string `toml:"start_date_field"`
	}
)

func LoadConfig(filepath string) (*Config, error) {
	f, err := os.Open(filepath)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %s. %s", filepath, err)
	}

	var cfg Config
	if err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return &cfg, nil
}
