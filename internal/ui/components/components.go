// SPDX-FileCopyrightText: 2025 Hector @sudosu404 <hector@email.gnx>
// SPDX-License-Identifier: AGPL3

package components

func IconURL(name string) string {
	if name == "" {
		name = "tailnet"
	}
	return "/icons/" + name + ".svg"
}
