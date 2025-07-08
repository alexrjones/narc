package daemon

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/alexrjones/narc"
)

type Server struct {
	d          *Daemon
	termSignal chan struct{}
}

func NewServer(d *Daemon, termSignal chan struct{}) *Server {
	return &Server{d: d, termSignal: termSignal}
}

func (s *Server) GetHandler() http.Handler {

	m := http.NewServeMux()
	m.Handle("GET /up", http.HandlerFunc(s.HandleUp))
	m.Handle("POST /start", http.HandlerFunc(s.HandleStartActivity))
	m.Handle("POST /end", http.HandlerFunc(s.HandleStopActivity))
	m.Handle("POST /terminate", http.HandlerFunc(s.HandleTerminate))
	m.Handle("GET /status", http.HandlerFunc(s.HandleStatus))
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

	err := s.d.StopActivity(r.Context(), narc.ChangeReasonExplicitStop)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	writeOK(rw)
}

func (s *Server) HandleTerminate(rw http.ResponseWriter, r *http.Request) {

	err := s.d.StopActivity(r.Context(), narc.ChangeReasonDaemonExit)
	if err != nil {
		log.Println("Couldn't stop current activity:", err)
	}
	writeOK(rw)
	s.termSignal <- struct{}{}
}

func (s *Server) HandleStatus(rw http.ResponseWriter, r *http.Request) {

	cur := s.d.getStatus()
	if !cur.valid() {
		writeString(rw, "No activity set.")
	} else {
		writeString(rw, fmt.Sprintf("Current activity: %s\nStarted at: %s\nRunning for: %s", cur.activity, cur.periodStart.Format(time.RFC3339), time.Since(cur.periodStart)))
	}
}

func writeOK(rw http.ResponseWriter) {
	rw.WriteHeader(200)
	rw.Write([]byte("OK"))
}

func writeString(rw http.ResponseWriter, s string) {
	rw.WriteHeader(200)
	rw.Write([]byte(s))
}
