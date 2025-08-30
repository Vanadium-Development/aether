package state

import (
	"node/internal/checksum"
	"sync"

	"github.com/google/uuid"
)

type SceneMetadata struct {
	Checksum     checksum.Checksum `json:"checksum"`
	CreatedAt    int64             `json:"created_at"`
	Filename     string            `json:"filename"`
	OriginalName string            `json:"original_name"`
	ID           uuid.UUID         `json:"id"`
}

type State struct {
	Scene      *SceneMetadata
	UploadLock sync.Mutex
	RenderLock sync.Mutex
}

type Platform int

const (
	Windows Platform = iota
	Unix
)

type AetherNode struct {
	ID       uuid.UUID
	Name     string
	Port     uint16
	Color    RGBColor
	State    State
	Platform Platform
}

func (node *AetherNode) NodeInfoMap() map[string]interface{} {
	return map[string]interface{}{
		"id":   node.ID.String(),
		"name": node.Name,
		"color": map[string]interface{}{
			"r": node.Color.R,
			"g": node.Color.G,
			"b": node.Color.B,
		},
	}
}
