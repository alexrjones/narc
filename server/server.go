package server

import (
	"io"
	"net/http"

	"github.com/alexrjones/narc/daemon"
)

type Server struct {
	d *daemon.Daemon
}

func New(d *daemon.Daemon) *Server {
	return &Server{d: d}
}

func (s *Server) GetHandler() http.Handler {

	m := http.NewServeMux()
	m.Handle("GET /up", http.HandlerFunc(s.HandleUp))
	m.Handle("POST /start", http.HandlerFunc(s.HandleStartActivity))
	m.Handle("POST /end", http.HandlerFunc(s.HandleStopActivity))
	return m
}

func (s *Server) HandleUp(rw http.ResponseWriter, r *http.Request) {

	writeOK(rw)
}

func (s *Server) HandleStartActivity(rw http.ResponseWriter, r *http.Request) {

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	name := string(b)
	if name == "" {
		http.Error(rw, "empty name sent in start activity", http.StatusBadRequest)
		return
	}
	err = s.d.SetActivity(r.Context(), name)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	writeOK(rw)
}

func (s *Server) HandleStopActivity(rw http.ResponseWriter, r *http.Request) {

	err := s.d.StopActivity(r.Context())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	writeOK(rw)
}

func writeOK(rw http.ResponseWriter) {
	rw.WriteHeader(200)
	rw.Write([]byte("OK"))
}
