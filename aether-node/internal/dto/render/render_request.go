package render

import "github.com/google/uuid"

type RenderRequest struct {
	ID         *uuid.UUID `json:"id"`
	FrameStart *uint16    `json:"frame_start"`
	FrameEnd   *uint16    `json:"frame_end"`
}
