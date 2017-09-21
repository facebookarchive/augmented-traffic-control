package api

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type LocalAssetManager struct {
	srv *Server
	dir string
}

func (mgr *LocalAssetManager) path(els ...string) string {
	els = append([]string{mgr.dir}, els...)
	return path.Clean(path.Join(els...))
}

func (mgr *LocalAssetManager) Index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(mgr.path("index.html.tmpl"))
	if err != nil {
		Log.Printf("Could not parse template for index page: %v", err)
		w.WriteHeader(500)
		return
	}

	// Need to do this so GetClientAddr will work
	context.Set(r, srv_context_key, mgr.srv)

	tmpl_data, err := mgr.srv.bind_info.templateFor(r)
	if err != nil {
		Log.Printf("Could not generate index page: %v", err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	tmpl.Execute(w, tmpl_data)
}

func (mgr *LocalAssetManager) Asset(w http.ResponseWriter, r *http.Request) {
	name, ok := mux.Vars(r)["name"]
	if !ok {
		w.WriteHeader(404)
		return
	}
	folder, ok := mux.Vars(r)["folder"]
	if !ok {
		w.WriteHeader(404)
		return
	}
	asset_name := mgr.path(folder, name)
	f, err := os.Open(asset_name)
	if err != nil {
		Log.Printf("Could not find asset %q: %v", asset_name, err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	_, err = io.Copy(w, f)
	if err != nil {
		Log.Printf("Could not write asset %q: %v", asset_name, err)
		return
	}
}
