// SPDX-FileCopyrightText: 2024 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT
package components

import "github.com/almeidapaulopt/tsdproxy/web"

func IconURL(name string) string {
	if name == "" {
		name = "tsdproxy"
	}
	return "/icons/" + name + ".svg"
}

func EmbededSVGIcon(name string) string {
	return web.GetFile("icons/" + name + ".svg")
}
