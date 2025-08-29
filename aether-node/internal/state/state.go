package state

import (
	"github.com/google/uuid"
)

type Node struct {
	ID   uuid.UUID
	Name string
}

func (node *Node) NodeInfoMap() map[string]string {
	return map[string]string{
		"id":   node.ID.String(),
		"name": node.Name,
	}
}
