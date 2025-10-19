// SPDX-FileCopyrightText: 2025 Hector @sudosu404 <hector@email.gnx>
// SPDX-License-Identifier: AGPL3

package docker

import (
	"errors"
)

type NoValidTargetFoundError struct {
	containerName string
}

func (n *NoValidTargetFoundError) Error() string {
	return "no valid target found for " + n.containerName
}

var (
	ErrNoPortFoundInContainer              = errors.New("no port found in container")
	ErrNoValidTargetFoundForInternalPorts  = errors.New("no valid target found for internal ports")
	ErrNoValidTargetFoundForPublishedPorts = errors.New("no valid target found for exposed ports")
)
