package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *AdminServer) addCommonRoutes() {
	addCommonRoutes(s.AppState, s.Router)
}

func (s *PublicServer) addCommonRoutes() {
	addCommonRoutes(s.AppState, s.Router)
}

func addCommonRoutes(a *AppState, r *(mux.Router)) {
	fs := http.FileServer(http.Dir("./www/assets/"))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fs))

	ifs := http.FileServer(http.Dir(a.imageDirectoryPath))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", ifs))

	r.NotFoundHandler = http.RedirectHandler("/", http.StatusFound)
}
