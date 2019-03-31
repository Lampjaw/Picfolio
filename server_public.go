package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type PublicServer struct {
	Port       string
	Router     *mux.Router
	Repository *Repository
}

type PublicMainPageData struct {
	Albums []*AlbumRecord
}

type PublicAlbumPageData struct {
	Album  *AlbumRecord
	Images []*ImageRecord
}

func newPublicServer(r *Repository) *PublicServer {
	return &PublicServer{
		Port:       "80",
		Router:     mux.NewRouter(),
		Repository: r,
	}
}

func (s *PublicServer) startListeningPublic(exitCallback chan bool) {
	s.Router.HandleFunc("/", s.handleMainPage)
	s.Router.HandleFunc("/album/{albumID}", s.handleAlbumPage)

	fs := http.FileServer(http.Dir("./www/assets/"))
	s.Router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fs))

	ifs := http.FileServer(http.Dir("./images/"))
	s.Router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", ifs))

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%s", s.Port), s.Router); err != nil {
			panic(err)
		}
		exitCallback <- true
	}()

	log.Printf("Public server started on port %s", s.Port)
}

func (s *PublicServer) handleMainPage(w http.ResponseWriter, r *http.Request) {
	albumRecords, err := s.Repository.getAllAlbumRecords()
	if err != nil {
		log.Println(err)
	}

	data := &PublicMainPageData{
		Albums: albumRecords,
	}

	tmpl := template.Must(template.ParseFiles("www/public/public_wrapper.html", "www/common_head.html", "www/common_foot.html", "www/public/public.html"))

	tmpl.Execute(w, data)
}

func (s *PublicServer) handleAlbumPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["albumID"]

	albumRecord, err := s.Repository.getAlbumRecord(albumID)
	if err != nil {
		log.Println(err)
	}

	imageRecords, err := s.Repository.getAllImageRecordsByAlbumID(albumID)
	if err != nil {
		log.Println(err)
	}

	data := &PublicAlbumPageData{
		Album:  albumRecord,
		Images: imageRecords,
	}

	tmpl := template.Must(template.ParseFiles("www/public/public_wrapper.html", "www/common_head.html", "www/common_foot.html", "www/public/album.html", "www/photoswipe.html"))

	tmpl.Execute(w, data)
}
