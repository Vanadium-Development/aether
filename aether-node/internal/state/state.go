package state

import (
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
)

type Checksum []byte

func (c *Checksum) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	*c = b
	return nil
}

type SceneMetadata struct {
	Checksum   Checksum `json:"checksum"`
	FramesFrom uint16
	FramesTo   uint16
}

type Scene struct {
	Filename string
	FilePath string
	Metadata SceneMetadata
}

type State struct {
	Scene *Scene
}

type Node struct {
	ID    uuid.UUID
	Name  string
	Color RGBColor
	State State
}

func (node *Node) NodeInfoMap() map[string]interface{} {
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
