package narc

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type StorageType string

const (
	StorageTypeCSV StorageType = "CSV"
)

type Config struct {
	ServerBaseURL string      `split_words:"true" default:"http://localhost:8080"`
	StorageType   StorageType `split_words:"true" default:"CSV"`
	CSVPath       string      `split_words:"true"`
}

func getConfigPath() (string, error) {

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "narc.yaml"), nil
}

func getDataPath(subpath string) (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, subpath), nil
}

func GetConfig() (*Config, error) {
	c := &Config{ServerBaseURL: "http://localhost:8080", StorageType: StorageTypeCSV}
	if c.StorageType == StorageTypeCSV && c.CSVPath == "" {
		var err error
		c.CSVPath, err = getDataPath(fmt.Sprintf("%d.csv", time.Now().Unix()))
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}
