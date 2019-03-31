package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/securecookie"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const SessionCookieName string = "picfolio.session"
const SessionDuration int = 86400

type AdminServer struct {
	Port             string
	AdminCredentials *Credentials
	Router           *mux.Router
	CookieStore      *sessions.CookieStore
	Repository       *Repository
}

type Credentials struct {
	Username string
	Password string
}

type LoginPageData struct {
	IsError      bool
	ErrorMessage string
}

type AdminPageData struct {
	Albums []*AlbumRecord
}

type AlbumPageData struct {
	Album  *AlbumRecord
	Images []*ImageRecord
}

func newAdminServer(r *Repository) *AdminServer {
	adminKey := os.Getenv("ADMIN_KEY")

	if adminKey == "" {
		log.Println("Missing ADMIN_KEY.")
	} else {
		//adminCredentials := getBase64Credentials(adminKey)
	}

	adminCredentials := &Credentials{
		Username: "root",
		Password: "root",
	}

	return &AdminServer{
		Port:             "8080",
		AdminCredentials: adminCredentials,
		Router:           mux.NewRouter(),
		CookieStore:      sessions.NewCookieStore(securecookie.GenerateRandomKey(32)),
		Repository:       r,
	}
}

func (s *AdminServer) startListeningAdmin(exitCallback chan bool) {
	if s.AdminCredentials == nil {
		return
	}

	s.Router.Handle("/", http.RedirectHandler("login", http.StatusFound))

	s.Router.HandleFunc("/login", s.handleLoginPage)
	s.Router.HandleFunc("/logout", s.handleLogout)
	s.Router.HandleFunc("/admin", s.handleAdminPage)
	s.Router.HandleFunc("/upload/{albumID}", s.handleUpload).Methods("POST")
	s.Router.HandleFunc("/image/{imageID}", s.handleImageUpdate).Methods("POST")
	s.Router.HandleFunc("/image/{imageID}", s.handleImageDelete).Methods("DELETE")
	s.Router.HandleFunc("/image/{imageID}/rotate", s.handleImageRotate).Methods("POST")
	s.Router.HandleFunc("/album", s.handleAlbumCreate).Methods("POST")
	s.Router.HandleFunc("/album/{albumID}", s.handleAlbumUpdate).Methods("POST")
	s.Router.HandleFunc("/album/{albumID}", s.handleAlbumDelete).Methods("DELETE")
	s.Router.HandleFunc("/album/{albumID}", s.handleAlbumPage)
	s.Router.HandleFunc("/album/{albumID}/edit", s.handleAlbumEditPage)

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

	log.Printf("Admin server started on port %s", s.Port)
}

func (s *AdminServer) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if isAuthorized {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	data := &LoginPageData{}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == s.AdminCredentials.Username && password == s.AdminCredentials.Password {
			session, err := s.CookieStore.Get(r, SessionCookieName)

			if session == nil && err != nil {
				log.Println(err)
				return
			}

			// Expires after 24 hours
			session.Options.MaxAge = SessionDuration

			session.Save(r, w)

			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		data.IsError = true
		data.ErrorMessage = "Incorrect username or password"
	}

	tmpl := template.Must(template.ParseFiles("www/admin/login.html", "www/common_head.html", "www/common_foot.html"))

	tmpl.Execute(w, data)
}

func (s *AdminServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := s.CookieStore.Get(r, SessionCookieName)

	if !session.IsNew {
		session.Options.MaxAge = -1
		session.Save(r, w)
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *AdminServer) handleAdminPage(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	albumRecords, err := s.Repository.getAllAlbumRecords()
	if err != nil {
		log.Println(err)
	}

	data := &AdminPageData{
		Albums: albumRecords,
	}

	tmpl := template.Must(template.ParseFiles("www/admin/admin_wrapper.html", "www/common_head.html", "www/common_foot.html", "www/admin/admin.html"))

	tmpl.Execute(w, data)
}

func (s *AdminServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	albumID := vars["albumID"]

	_, err := uploadFiles(s.Repository, albumID, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/album/%s", albumID), http.StatusFound)
}

func (s *AdminServer) handleAlbumCreate(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	albumID, _ := s.Repository.createAlbumRecord(title, description)

	http.Redirect(w, r, fmt.Sprintf("/album/%s", albumID), http.StatusFound)
}

func (s *AdminServer) handleAlbumPage(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

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

	data := &AlbumPageData{
		Album:  albumRecord,
		Images: imageRecords,
	}

	tmpl := template.Must(template.ParseFiles("www/admin/admin_wrapper.html", "www/common_head.html", "www/common_foot.html", "www/admin/album.html", "www/photoswipe.html"))

	tmpl.Execute(w, data)
}

func (s *AdminServer) handleAlbumUpdate(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	albumID := vars["albumID"]

	currentAlbum, err := s.Repository.getAlbumRecord(albumID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if currentAlbum == nil {
		http.NotFound(w, r)
		return
	}

	title := r.FormValue("title")
	description := nilString(r.FormValue("description"))
	coverPhotoID := nilString(r.FormValue("coverPhotoId"))

	if title == "" {
		title = currentAlbum.Title
	}

	err = s.Repository.updateAlbum(albumID, title, description, coverPhotoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *AdminServer) handleAlbumDelete(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	albumID := vars["albumID"]

	currentAlbum, err := s.Repository.getAlbumRecord(albumID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if currentAlbum == nil {
		http.NotFound(w, r)
		return
	}

	err = deleteAlbum(s.Repository, albumID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *AdminServer) handleImageUpdate(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	imageID := vars["imageID"]

	currentImage, err := s.Repository.getImageRecord(imageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if currentImage == nil {
		http.NotFound(w, r)
		return
	}

	currentImage.Description = nilString(r.FormValue("description"))

	err = s.Repository.updateImage(imageID, currentImage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *AdminServer) handleImageDelete(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	imageID := vars["imageID"]

	currentImage, err := s.Repository.getImageRecord(imageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if currentImage == nil {
		http.NotFound(w, r)
		return
	}

	err = deleteImage(s.Repository, imageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *AdminServer) handleImageRotate(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	imageID := vars["imageID"]

	err := rotateImage(s.Repository, imageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func nilString(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

func (s *AdminServer) handleAlbumEditPage(w http.ResponseWriter, r *http.Request) {
	isAuthorized, _ := checkIsAuthorized(s, r)
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

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

	data := &AlbumPageData{
		Album:  albumRecord,
		Images: imageRecords,
	}

	tmpl := template.Must(template.ParseFiles("www/admin/admin_wrapper.html", "www/common_head.html", "www/common_foot.html", "www/admin/album_edit.html"))

	tmpl.Execute(w, data)
}

func checkIsAuthorized(s *AdminServer, r *http.Request) (bool, error) {
	session, err := s.CookieStore.Get(r, SessionCookieName)

	if session == nil && err != nil {
		return false, err
	}

	return !session.IsNew, nil
}

func getBase64Credentials(encoded string) *Credentials {
	if encoded == "" {
		return nil
	}

	decoded, err := decodeBase64(encoded)
	if err != nil {
		log.Printf("Unable to decode credentials: %v", err)
		return nil
	}

	credParts := strings.Split(decoded, ":")

	return &Credentials{
		Username: credParts[0],
		Password: credParts[1],
	}
}

func decodeBase64(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)

	if err != nil {
		return "", err
	}

	return string(decodedBytes), nil
}
