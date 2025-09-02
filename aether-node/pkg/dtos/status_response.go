package dtos

type AetherRenderProgressDto struct {
	CurrentFrame  int     `json:"current_frame"`
	FramePercent  float64 `json:"frame_percent"`
	FrameCount    int     `json:"frame_count"`
	TimeElapsed   float64 `json:"time_elapsed"`
	TimeRemaining float64 `json:"time_remaining"`
}

type AetherStatusDto struct {
	IsRendering bool                     `json:"is_rendering"`
	Request     *AetherRenderDto         `json:"request"`
	Progress    *AetherRenderProgressDto `json:"progress"`
}

func EmptyStatusResponse() AetherStatusDto {
	return AetherStatusDto{
		IsRendering: false,
		Request:     nil,
		Progress:    nil,
	}
}
