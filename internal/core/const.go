// SPDX-FileCopyrightText: 2025 Hector @sudosu404 <hector@email.gnx>
// SPDX-License-Identifier: AGPL3

package core

import "time"

const (
	// G112 (CWE-400): Potential Slowloris Attack because ReadHeaderTimeout is not configured in the http.Server (Confidence: LOW, Severity: MEDIUM.
	ReadHeaderTimeout = 5 * time.Second
)
