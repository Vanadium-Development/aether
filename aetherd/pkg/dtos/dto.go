package dtos

import (
	"encoding/json"
)

type DaemonRequestType int

const (
	// StatusReqeust Returns status of the daemon
	StatusReqeust DaemonRequestType = iota

	// CommitRequest Accepts tracked files, render settings and output path;
	// prepares data  and initiates distributed rendering upon distributing
	// the data to all available nodes
	CommitRequest
)

type DaemonCommitDto struct {
	Files      []string `json:"files"`
	FrameStart uint16   `json:"frame_start"`
	FrameEnd   uint16   `json:"frame_end"`
}

type DaemonCommonDto struct {
	RequestType DaemonRequestType `json:"type"`
	Body        json.RawMessage   `json:"body"`
}
