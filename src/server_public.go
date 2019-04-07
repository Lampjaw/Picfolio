package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type PublicServer struct {
	Port         string
	Router       *mux.Router
	AppState     *AppState
	ImageManager *ImageManager
	AlbumManager *AlbumManager
}

type PublicMainPageData struct {
	Albums []*AlbumRecord
}

type PublicAlbumPageData struct {
	Album  *AlbumRecord
	Images []*ImageRecord
}

func newPublicServer(a *AppState) *PublicServer {
	return &PublicServer{
		Port:         "80",
		Router:       mux.NewRouter(),
		AppState:     a,
		ImageManager: a.ImageManager,
		AlbumManager: a.AlbumManager,
	}
}

func (s *PublicServer) startListeningPublic() {
	s.Router.HandleFunc("/", s.handleMainPage)
	s.Router.HandleFunc("/album/{albumID}", s.handleAlbumPage)

	s.addCommonRoutes()

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%s", s.Port), s.Router); err != nil {
			panic(err)
		}
		s.AppState.exitCallback <- true
	}()

	log.Printf("Public server started on port %s", s.Port)
}

func (s *PublicServer) handleMainPage(w http.ResponseWriter, r *http.Request) {
	albumRecords, err := s.AlbumManager.getAllAlbums()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &PublicMainPageData{
		Albums: albumRecords,
	}

	tmpl := template.Must(template.ParseFiles("www/public/public_wrapper.html", "www/common_head.html", "www/common_foot.html", "www/album_list.html"))

	tmpl.Execute(w, data)
}

func (s *PublicServer) handleAlbumPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["albumID"]

	albumRecord, err := s.AlbumManager.getAlbum(albumID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if albumRecord == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	imageRecords, err := s.ImageManager.getAllImagesByAlbumID(albumID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &PublicAlbumPageData{
		Album:  albumRecord,
		Images: imageRecords,
	}

	tmpl := template.Must(template.ParseFiles("www/public/public_wrapper.html", "www/common_head.html", "www/common_foot.html", "www/public/album.html", "www/photoswipe.html"))

	tmpl.Execute(w, data)
}
