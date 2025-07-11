package daemon

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexrjones/narc"
)

type Server struct {
	d          *Daemon
	s          Store
	termSignal chan SignalPacket
}

func NewServer(d *Daemon, s Store, termSignal chan SignalPacket) *Server {
	return &Server{d: d, s: s, termSignal: termSignal}
}

func (s *Server) GetHandler() http.Handler {

	m := http.NewServeMux()
	m.Handle("GET /up", http.HandlerFunc(s.HandleUp))
	m.Handle("POST /start", http.HandlerFunc(s.HandleStartActivity))
	m.Handle("POST /end", http.HandlerFunc(s.HandleStopActivity))
	m.Handle("POST /terminate", http.HandlerFunc(s.HandleTerminate))
	m.Handle("POST /reload", http.HandlerFunc(s.HandleConfigReload))
	m.Handle("GET /status", http.HandlerFunc(s.HandleStatus))
	m.Handle("GET /aggregate", http.HandlerFunc(s.HandleAggregateActivities))
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
	ignoreIdle := false
	ignoreIdleQ := r.URL.Query().Get("ignoreIdle")
	if ignoreIdleParsed, err := strconv.ParseBool(ignoreIdleQ); err == nil {
		ignoreIdle = ignoreIdleParsed
	}

	err = s.d.SetActivity(r.Context(), name, WithIgnoreIdle(ignoreIdle))
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
	s.termSignal <- SignalPacket{Signal: SignalTerm}
}

func (s *Server) HandleStatus(rw http.ResponseWriter, r *http.Request) {

	cur := s.d.getStatus()
	if !cur.validActivity() {
		writeString(rw, "No activity set.")
	} else {
		if cur.validPeriod() {
			writeString(rw, fmt.Sprintf("Current activity: %s\nStarted at: %s\nRunning for: %s", cur.activity, cur.periodStart.Format(time.RFC3339), time.Since(cur.periodStart)))
		} else {
			writeString(rw, fmt.Sprintf("Current activity: %s\nCurrently idle (may not have polled for user activity yet)", cur.activity))
		}
	}
}

func (s *Server) HandleAggregateActivities(rw http.ResponseWriter, r *http.Request) {

	start, end := r.URL.Query().Get("start"), r.URL.Query().Get("end")
	startTime, _ := time.Parse(time.DateOnly, start)
	endTime, _ := time.Parse(time.DateOnly, end)
	roundStr := r.URL.Query().Get("round")
	round := true
	if v, err := strconv.ParseBool(roundStr); err == nil {
		round = v
	}
	activities, err := s.s.GetActivities(r.Context(), startTime, endTime)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rows := narc.Activities(activities).ToDurationRows()
	sb := new(strings.Builder)
	csvw := csv.NewWriter(sb)
	for _, row := range rows {
		dur := row.Duration.Hours()
		if round {
			dur = ceilToNearestFiveCents(dur)
		}
		csvw.Write([]string{row.Date.Format(time.DateOnly), row.Name, fmt.Sprintf("%.2f", dur)})
	}
	csvw.Flush()
	writeString(rw, sb.String())
}

func (s *Server) HandleConfigReload(rw http.ResponseWriter, r *http.Request) {

	cur := s.d.getStatus()
	err := s.d.StopActivity(r.Context(), narc.ChangeReasonDaemonExit)
	if err != nil {
		log.Println("Couldn't stop current activity:", err)
	}
	writeOK(rw)
	s.termSignal <- SignalPacket{Signal: SignalHup, LastActivityName: cur.activity, LastActivityIgnoreIdle: cur.ignoreIdle}
}

func roundToNearestQuarter(f float64) float64 {
	return math.Round(f*4) / 4
}

func ceilToNearestQuarter(f float64) float64 {
	return math.Ceil(f*4) / 4
}

func ceilToNearestFiveCents(f float64) float64 {
	return math.Ceil(f*20) / 20
}

func writeOK(rw http.ResponseWriter) {
	rw.WriteHeader(200)
	rw.Write([]byte("OK"))
}

func writeString(rw http.ResponseWriter, s string) {
	rw.WriteHeader(200)
	rw.Write([]byte(s))
}
