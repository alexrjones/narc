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
	ServerBaseURL string
	StorageType   StorageType
	CSVPath       string
	LogPath       string
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

func GetConfig() (*Config, error) {
	dataDir, err := ensureDataDir()
	if err != nil {
		return nil, err
	}
	c := &Config{
		ServerBaseURL: "http://localhost:" + defaultPort,
		StorageType:   StorageTypeCSV,
		CSVPath:       filepath.Join(dataDir, "narc.csv"),
		LogPath:       filepath.Join(dataDir, "narc.log"),
	}
	return c, nil
}
