package state

import (
	"node/internal/checksum"
	"node/internal/dto/render"
	"node/internal/util"
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

type RendererState struct {
	Scene         SceneMetadata
	Request       render.RenderRequest
	CurrentFrame  int
	FramePercent  float64
	TimeElapsed   float64
	TimeRemaining float64
}

type State struct {
	RendererState *RendererState
	UploadLock    sync.Mutex
	RenderLock    sync.Mutex
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
	Color    util.RGBColor
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
