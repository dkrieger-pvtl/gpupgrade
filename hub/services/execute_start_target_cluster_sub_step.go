package services

import (
	"io"
)

func (h *Hub) StartTargetCluster(stream messageSender, log io.Writer, args ...string) error {
	return StartCluster(stream, log, h.target)
}
