package node

import (
	"fmt"
	"node/internal/api"
	"node/internal/banner"
	"node/internal/config"
	"node/internal/state"

	"github.com/google/uuid"
)

func InitializeNode(port uint16) {
	fmt.Println(banner.AetherBanner)

	var nodeId, _ = uuid.NewRandom()
	var cfg = config.ParseNodeConfig()
	var n = state.Node{
		ID:    nodeId,
		Name:  cfg.Node.Name,
		Color: state.RandomNodeColor(),
		State: state.State{
			Scene: nil,
		}}

	api.InitializeApi(cfg.Node.Port, n, cfg)
}
