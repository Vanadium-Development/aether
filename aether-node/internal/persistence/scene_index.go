package persistence

import (
	"encoding/json"
	"node/internal/checksum"
	"node/internal/config"
	"node/internal/state"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SceneIndex struct {
	CreatedAt int64                 `json:"created_at"`
	Scenes    []state.SceneMetadata `json:"scenes"`
}

func (store *SceneIndex) AddScene(scene state.SceneMetadata, cfg *config.NodeConfig) {
	store.Scenes = append(store.Scenes, scene)
	StoreIndex(cfg, store)
}

func (store *SceneIndex) FindSceneByChecksum(checksum checksum.Checksum) *state.SceneMetadata {
	for i := range store.Scenes {
		scene := store.Scenes[i]
		if scene.Checksum.IsSame(&checksum) {
			return &scene
		}
	}

	return nil
}

func (store *SceneIndex) FindSceneById(id uuid.UUID) *state.SceneMetadata {
	for i := range store.Scenes {
		scene := store.Scenes[i]
		if scene.ID == id {
			return &scene
		}
	}

	return nil
}

func (store *SceneIndex) EnsureSceneIndex(cfg *config.NodeConfig) bool {
	_, err := os.Stat(cfg.Data.SceneIndex)
	if err == nil {
		return false
	}
	if !os.IsNotExist(err) {
		return false
	}
	_, err = os.Create(cfg.Data.SceneIndex)
	if err != nil {
		logrus.Fatal("Failed to create scene index: ", err)
		return false
	}

	logrus.Infof("Created scene index file \"%s\".\n", cfg.Data.SceneIndex)
	StoreIndex(cfg, store)

	return true
}

func LoadStoredScenes(cfg *config.NodeConfig) SceneIndex {
	var store = SceneIndex{
		Scenes:    []state.SceneMetadata{},
		CreatedAt: time.Now().UnixNano(),
	}

	// If the file was just created, the index is going to be empty. No need to proceed
	if (&store).EnsureSceneIndex(cfg) {
		return store
	}

	file, err := os.ReadFile(cfg.Data.SceneIndex)
	if err != nil {
		return store
	}

	err = json.Unmarshal(file, &store)
	if err != nil {
		logrus.Errorf("Could not load scene index: %s\n", err)
		return store
	}

	sceneCount := len(store.Scenes)

	if sceneCount > 0 {
		logrus.Infof("Loaded %d scenes.\n", sceneCount)
	} else {
		logrus.Infof("No scenes to load.\n")
	}

	return store
}

func StoreIndex(cfg *config.NodeConfig, store *SceneIndex) {
	b, err := json.MarshalIndent(store, "", "\t")
	if err != nil {
		logrus.Errorf("Could not marshal scenes: %s", err)
		return
	}

	err = os.WriteFile(cfg.Data.SceneIndex, b, os.ModePerm)
	if err != nil {
		return
	}

	logrus.Infof("Written scene index (%s)\n", humanize.Bytes(uint64(len(b))))
}
