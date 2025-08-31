package progress

import (
	"node/internal/dto/render"
	"node/internal/state"
)

type RenderProgress struct {
	CurrentFrame  int     `json:"current_frame"`
	FramePercent  float64 `json:"frame_percent"`
	FrameCount    int     `json:"frame_count"`
	TimeElapsed   float64 `json:"time_elapsed"`
	TimeRemaining float64 `json:"time_remaining"`
}

type StatusResponse struct {
	IsRendering bool                  `json:"is_rendering"`
	Request     *render.RenderRequest `json:"request"`
	Progress    *RenderProgress       `json:"progress"`
}

func EmptyStatusResponse() StatusResponse {
	return StatusResponse{
		IsRendering: false,
		Request:     nil,
		Progress:    nil,
	}
}

func StatusResponseFromRenderState(state *state.RendererState) StatusResponse {
	if state == nil {
		return EmptyStatusResponse()
	}

	return StatusResponse{
		IsRendering: true,
		Request:     &state.Request,
		Progress: &RenderProgress{
			CurrentFrame:  state.CurrentFrame,
			FramePercent:  state.FramePercent,
			FrameCount:    int(*state.Request.FrameEnd-*state.Request.FrameStart) + 1,
			TimeElapsed:   state.TimeElapsed,
			TimeRemaining: state.TimeRemaining,
		},
	}
}
