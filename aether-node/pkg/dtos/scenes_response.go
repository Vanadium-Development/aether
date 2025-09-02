package dtos

import (
	"github.com/google/uuid"
)

type AetherSceneDto struct {
	CreatedAt    int64     `json:"created_at"`
	ID           uuid.UUID `json:"id"`
	OriginalName string    `json:"original_name"`
}

type AetherSceneIndexDto struct {
	Scenes []AetherSceneDto `json:"scenes"`
}
