package api

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/facebook/augmented-traffic-control/src/assets"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type BundleAssetManager struct {
	srv *Server
}

func (mgr *BundleAssetManager) Index(w http.ResponseWriter, r *http.Request) {
	data, err := assets.Asset("static/index.html.tmpl")
	if err != nil {
		Log.Printf("Could not find index page asset: %v", err)
		w.WriteHeader(404)
		return
	}

	// Need to do this so GetClientAddr will work
	context.Set(r, srvContextKey, mgr.srv)

	tmpl, err := template.New("root").Parse(string(data))
	if err != nil {
		Log.Printf("Could not parse template for index page: %v", err)
		w.WriteHeader(500)
		return
	}
	tmpl_data, err := mgr.srv.bind_info.templateFor(r)
	if err != nil {
		Log.Printf("Could not generate index page: %v", err)
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	tmpl.Execute(w, tmpl_data)
}

func (mgr *BundleAssetManager) Asset(w http.ResponseWriter, r *http.Request) {
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
	asset_name := fmt.Sprintf("static/%s/%s", folder, name)
	data, err := assets.Asset(asset_name)
	if err != nil {
		Log.Printf("Could not find asset %q: %v", asset_name, err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	_, err = w.Write(data)
	if err != nil {
		Log.Printf("Could not write asset %q: %v", asset_name, err)
		return
	}
}
