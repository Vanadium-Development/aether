package state

import (
	"node/internal/checksum"
	"node/internal/util"
	"node/pkg/dtos"
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
	Request       dtos.AetherRenderDto
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

func (state *RendererState) StatusDto() dtos.AetherStatusDto {
	if state == nil {
		return dtos.EmptyStatusResponse()
	}

	return dtos.AetherStatusDto{
		IsRendering: true,
		Request:     &state.Request,
		Progress: &dtos.AetherRenderProgressDto{
			CurrentFrame:  state.CurrentFrame,
			FramePercent:  state.FramePercent,
			FrameCount:    int(*state.Request.FrameEnd-*state.Request.FrameStart) + 1,
			TimeElapsed:   state.TimeElapsed,
			TimeRemaining: state.TimeRemaining,
		},
	}
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
