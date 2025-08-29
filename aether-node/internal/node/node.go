package node

import (
	"node/internal/api"
	"node/internal/state"

	"github.com/google/uuid"
)

func InitializeNode(port uint16) {
	var nodeId, _ = uuid.NewRandom()
	var n = state.Node{nodeId, "Linux Blender Node"}

	api.InitializeApi(port, n)
}
