package server

import (
	"io/fs"
	"net/http"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/server/web"
)

// Static serves files from the given directory under the specified URL prefix.
// The prefix must start and end with a slash (e.g., "/static/").
//
//	srv.Static("/static/", "./public")
func (s *Server) Static(prefix, dir string) {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	fileServer := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
	s.mux.Handle(web.MethodGet.String()+" "+prefix, fileServer)
}

// StaticFS serves files from the given fs.FS under the specified URL prefix.
// This is useful for embedding static files with go:embed.
//
//	//go:embed static
//	var staticFiles embed.FS
//	srv.StaticFS("/static/", staticFiles)
func (s *Server) StaticFS(prefix string, fsys fs.FS) {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	fileServer := http.StripPrefix(prefix, http.FileServerFS(fsys))
	s.mux.Handle(web.MethodGet.String()+" "+prefix, fileServer)
}
