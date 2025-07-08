package narc

import (
	"os"
	"path/filepath"
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
	root := filepath.Join(dir, ".narc")
	err = os.MkdirAll(root, 0750)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, subpath), nil
}

const defaultPort = "53300"

func GetConfig() (*Config, error) {
	c := &Config{ServerBaseURL: "http://localhost:" + defaultPort, StorageType: StorageTypeCSV}
	if c.StorageType == StorageTypeCSV && c.CSVPath == "" {
		var err error
		c.CSVPath, err = getDataPath("narc.csv")
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}
