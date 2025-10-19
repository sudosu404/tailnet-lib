// SPDX-FileCopyrightText: 2025 Hector @sudosu404 <hector@email.gnx>
// SPDX-License-Identifier: AGPL3

package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

// --- Embed prebuilt frontend assets and icons ---
//go:embed web/dist/* web/public/icons/*
var dist embed.FS

var Static http.Handler
const DefaultIcon = "tailnet"

func init() {
	staticFS, err := fs.Sub(dist, "web/dist")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open dist directory")
	}
	Static = statigz.FileServer(staticFS.(fs.ReadDirFS), brotli.AddEncoding)
}

func GuessIcon(name string) string {
	nameParts := strings.Split(name, "/")
	lastPart := nameParts[len(nameParts)-1]
	baseName := strings.SplitN(lastPart, ":", 2)[0]
	baseName = strings.SplitN(baseName, "@", 2)[0]

	var foundFile string
	err := fs.WalkDir(dist, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".svg") {
			if strings.TrimSuffix(d.Name(), ".svg") == baseName {
				foundFile = path
				return fs.SkipDir
			}
		}
		return nil
	})
	if err != nil || foundFile == "" {
		return DefaultIcon
	}
	icon := strings.TrimPrefix(foundFile, "web/dist/icons/")
	return strings.TrimSuffix(icon, ".svg")
}