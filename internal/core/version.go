// SPDX-FileCopyrightText: 2025 Hector @sudosu404 <hector@email.gnx>
// SPDX-License-Identifier: AGPL3

package core

import (
	"runtime/debug"
	"strings"
)

var (
	version        string
	realVersion    *string
	isDirty        *bool
	AppNameVersion = AppName + "-" + GetVersion()
)

const (
	AppName   = "Tailnet"
	AppAuthor = "Hector sudosu404 <hector@email.gnx>"
)

func GetVersion() string {
	if realVersion == nil {
		tempVersion := strings.TrimSpace(version)
		if getIsDirty() {
			tempVersion += "-dirty"
		}
		realVersion = &tempVersion
	}
	return *realVersion
}

func getIsDirty() bool {
	if isDirty != nil {
		return *isDirty
	}

	bi, ok := debug.ReadBuildInfo()
	if ok {
		modified := false

		for _, v := range bi.Settings {
			if v.Key == "vcs.modified" {
				if v.Value == "true" {
					modified = true
				}
			}
		}
		isDirty = &modified
	}
	return *isDirty
}
