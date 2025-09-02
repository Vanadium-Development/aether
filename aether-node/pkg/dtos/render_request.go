package dtos

import "github.com/google/uuid"

type AetherRenderDto struct {
	ID         *uuid.UUID `json:"id"`
	FrameStart *uint16    `json:"frame_start"`
	FrameEnd   *uint16    `json:"frame_end"`
}
