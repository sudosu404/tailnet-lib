// SPDX-FileCopyrightText: 2024 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

//go:generate wget https://github.com/selfhst/icons/archive/refs/heads/main.zip
//go:generate unzip -jo main.zip icons-main/svg/* -d public/icons/sh
//go:generate bun run build

//go:embed dist
var dist embed.FS

var Static http.Handler

func init() {
	s, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Fatal(err)
	}

	Static = statigz.FileServer(s.(fs.ReadDirFS), brotli.AddEncoding)
}
