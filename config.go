package narc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type StorageType string

const (
	StorageTypeCSV StorageType = "CSV"
)

type Config struct {
	ServerBaseURL string        `yaml:"serverBaseUrl,omitempty" validate:"omitempty,url"`
	StorageType   StorageType   `yaml:"storageType,omitempty" validate:"omitempty,oneof=CSV"`
	CSVPath       string        `yaml:"csvPath,omitempty" validate:"omitempty,filepath"`
	LogPath       string        `yaml:"logPath,omitempty" validate:"omitempty,filepath"`
	IdleTimeout   time.Duration `yaml:"idleTimeout,omitempty" validate:"omitempty,gte=1s"`
}

func (c *Config) mergeInOther(o *Config) {
	if o.ServerBaseURL != "" {
		c.ServerBaseURL = o.ServerBaseURL
	}
	if o.StorageType != "" {
		c.StorageType = o.StorageType
	}
	if o.CSVPath != "" {
		c.CSVPath = o.CSVPath
	}
	if o.LogPath != "" {
		c.LogPath = o.LogPath
	}
	if o.IdleTimeout != 0 {
		c.IdleTimeout = o.IdleTimeout
	}
}

func (c *Config) String() string {

	bytes, err := yaml.MarshalWithOptions(c, yaml.Indent(2))
	if err != nil {
		return fmt.Sprintf("failed to marshal config: %s", err)
	}
	return string(bytes)
}

func (c *Config) PropertyByName(name string) string {

	bytes, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("failed to marshal config: %s", err)
	}
	m := make(map[string]interface{})
	err = yaml.Unmarshal(bytes, &m)
	if err != nil {
		return fmt.Sprintf("failed to unmarshal config to map: %s", err)
	}
	val, ok := m[name]
	if !ok {
		return fmt.Sprintf("no property named '%s' exists", name)
	}
	return fmt.Sprint(val)
}

func ensureDataDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(dir, ".narc")
	err = os.MkdirAll(root, 0750)
	if err != nil {
		return "", err
	}
	return root, nil
}

func getDataPath(subpath string) (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(dir, ".narc")
	err = os.MkdirAll(root, 0750)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, subpath), nil
}

const defaultPort = "53300"

func loadOrCreateConfig(dataDir string) (*Config, error) {
	c := &Config{
		ServerBaseURL: "http://localhost:" + defaultPort,
		StorageType:   StorageTypeCSV,
		CSVPath:       filepath.Join(dataDir, "narc.csv"),
		LogPath:       filepath.Join(dataDir, "narc.log"),
		IdleTimeout:   time.Second * 300,
	}
	diskConf, err := loadDiskConfig(filepath.Join(dataDir, "config.yaml"))
	if err != nil {
		return nil, err
	}
	c.mergeInOther(diskConf)
	return c, nil
}

func loadDiskConfig(path string) (*Config, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	diskConf := new(Config)
	err = yaml.Unmarshal(bytes, diskConf)
	if err != nil {
		return nil, err
	}
	return diskConf, nil
}

func loadDiskConfigToMap(path string) (map[string]interface{}, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	diskConf := make(map[string]interface{})
	err = yaml.Unmarshal(bytes, &diskConf)
	if err != nil {
		return nil, err
	}
	return diskConf, nil
}

func SetConfigOption(name, val string) error {
	if val == "" {
		return errors.New("option value was empty")
	}
	op := "set"
	if val == "default" {
		op = "del"
	}
	dataDir, err := ensureDataDir()
	if err != nil {
		return err
	}
	diskConfPath := filepath.Join(dataDir, "config.yaml")
	dc, err := loadDiskConfigToMap(diskConfPath)
	if err != nil {
		return err
	}
	return updateDiskConfig(dc, name, val, op, diskConfPath)
}

func updateDiskConfig(m map[string]interface{}, name, value, op, path string) error {
	if m == nil {
		m = make(map[string]interface{})
	}
	if op == "set" {
		m[name] = value
	} else if op == "del" {
		delete(m, name)
	}

	var action func(f *os.File) error
	if len(m) > 0 {
		// Marshal the new map to bytes
		newBytes, err := yaml.Marshal(m)
		if err != nil {
			return err
		}
		// Unmarshal it back into a config with strict validation
		newConf := new(Config)
		err = yaml.UnmarshalWithOptions(newBytes, newConf,
			yaml.DisallowUnknownField(),
			yaml.Validator(validator.New(validator.WithRequiredStructEnabled())))
		if err != nil {
			return err
		}
		action = func(f *os.File) error {
			err := f.Truncate(0)
			if err != nil {
				return err
			}
			return yaml.NewEncoder(f).Encode(newConf)
		}
	} else {
		action = func(f *os.File) error {
			return f.Truncate(0)
		}
	}
	// If that passed, then let's save the new value back to disk
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return action(f)
}

func GetConfig() (*Config, error) {
	dataDir, err := ensureDataDir()
	if err != nil {
		return nil, err
	}
	return loadOrCreateConfig(dataDir)
}
