package storage

import "net"

type AetherNode struct {
	Name    string `json:"name"`
	Address net.IP `json:"address"`
}

type DaemonStorage struct {
	Nodes []AetherNode `json:"nodes"`
}
