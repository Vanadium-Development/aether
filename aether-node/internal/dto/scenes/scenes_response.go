package scenes

import (
	"node/internal/persistence"

	"github.com/google/uuid"
)

type SceneResponse struct {
	CreatedAt    int64     `json:"created_at"`
	ID           uuid.UUID `json:"id"`
	OriginalName string    `json:"original_name"`
}

type SceneIndexResponse struct {
	Scenes []SceneResponse `json:"scenes"`
}

func SceneIndexResponseFromIndex(index *persistence.SceneIndex) SceneIndexResponse {
	var scenes []SceneResponse
	for _, v := range index.Scenes {
		scenes = append(scenes, SceneResponse{
			CreatedAt:    v.CreatedAt,
			ID:           v.ID,
			OriginalName: v.OriginalName,
		})
	}
	return SceneIndexResponse{Scenes: scenes}
}
