package storage

import (
	"aetherd/internal/constants"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// ensureStorageDir Returns `true` if a new directory was created, `false` otherwise
func ensureStorageDir() (bool, error) {
	info, err := os.Stat(constants.StorageDir)
	if err == nil && !info.IsDir() {
		err = os.Remove(constants.StorageDir)
		if err != nil {
			logrus.Errorf("Failed to remove conflicting storage file (%s): %s", constants.StorageDir, err)
			return false, err
		}
	}

	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(constants.StorageDir, 0777)
		if err != nil {
			logrus.Errorf("Failed to mkdir %s: %s\n", constants.StorageDir, err)
			return false, err
		}
		logrus.Infof("Created storage directory %s\n", constants.StorageDir)
		return true, nil
	}

	return false, nil
}

func emptyStorage() DaemonStorage {
	return DaemonStorage{Nodes: []AetherNode{}}
}

func SaveStorage(storage *DaemonStorage) error {
	f, err := os.OpenFile(filepath.Join(constants.StorageDir, constants.StorageFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		logrus.Errorf("Could not save storage file: %s\n", err)
		return err
	}
	err = json.NewEncoder(f).Encode(*storage)
	if err != nil {
		logrus.Errorf("Could not encode storage file: %s\n", err)
		return err
	}
	return nil
}

func createEmptyStorage() (*DaemonStorage, error) {
	empty := emptyStorage()
	return &empty, SaveStorage(&empty)
}

func ReadStorage() (*DaemonStorage, error) {
	// If the directory was just created, we need to create an empty config file as well
	created, err := ensureStorageDir()
	if err != nil {
		return nil, err
	}
	if created {
		return createEmptyStorage()
	}

	// Otherwise we need to parse the existing file
	path := filepath.Join(constants.StorageDir, constants.StorageFile)
	file, err := os.Open(path)
	if err != nil {
		// Create empty storage file if it doesn't exist
		if os.IsNotExist(err) {
			return createEmptyStorage()
		}

		logrus.Errorf("Failed to open storage file (%s): %s\n", path, err)
		return nil, err
	}

	// Parse the storage
	var storage DaemonStorage
	err = json.NewDecoder(file).Decode(&storage)
	if err != nil {
		logrus.Errorf("Failed to parse storage file (%s): %s\n", path, err)
		return nil, err
	}

	return &storage, nil
}
