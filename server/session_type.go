package server

import (
	"fmt"
	"strings"
)

// SessionType determines the kind of ssh session
type SessionType int

const (
	SessionType_CONTROL SessionType = iota
	SessionType_NODE
	SessionType_JUMP
)

func (st SessionType) String() string {
	switch st {
	case SessionType_CONTROL:
		return "control"
	case SessionType_NODE:
		return "node"
	case SessionType_JUMP:
		return "jump"
	default:
		return ""
	}
}

// SessionTypeFromString converts strings to SessionType
func SessionTypeFromString(s string) (SessionType, error) {
	switch strings.ToLower(s) {
	case "control":
		return SessionType_CONTROL, nil
	case "node":
		return SessionType_NODE, nil
	case "jump":
		return SessionType_JUMP, nil
	default:
		return 0, fmt.Errorf("Unknown session type: %s", s)
	}
}
