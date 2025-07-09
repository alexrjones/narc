package narc

import (
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type StorageType string

const (
	StorageTypeCSV StorageType = "CSV"
)

type Config struct {
	ServerBaseURL string      `yaml:"serverBaseUrl,omitempty"`
	StorageType   StorageType `yaml:"storageType,omitempty"`
	CSVPath       string      `yaml:"csvPath,omitempty"`
	LogPath       string      `yaml:"logPath,omitempty"`
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
	}
	f, err := os.OpenFile(filepath.Join(dataDir, "config.yaml"), os.O_RDONLY|os.O_CREATE, 0644)
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
	c.mergeInOther(diskConf)
	return c, nil
}

func GetConfig() (*Config, error) {
	dataDir, err := ensureDataDir()
	if err != nil {
		return nil, err
	}
	return loadOrCreateConfig(dataDir)
}
