// Package plugindemo a demo plugin.
package plugindemo

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Config the plugin configuration.
type Config struct {
	Attempts int
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Listener is used to inform about retry attempts.
type Listener interface {
	// Retried will be called when a retry happens, with the request attempt passed to it.
	// For the first retry this will be attempt 2.
	Retried(req *http.Request, attempt int)
}

// Listeners is a convenience type to construct a list of Listener and notify
// each of them about a retry attempt.
type Listeners []Listener

// Retry a Demo plugin.
type Retry struct {
	attempts int
	next     http.Handler
	// listener Listener
	name string
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.Attempts <= 0 {
		return nil, fmt.Errorf("incorrect (or empty) value for attempt (%d)", config.Attempts)
	}
	return &Retry{
		attempts: config.Attempts,
		next:     next,
		// listener: listener,
		name: name,
	}, nil
}

func (r *Retry) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()
	sw := statusWriter{ResponseWriter: rw}
	r.next.ServeHTTP(rw, req)
	duration := time.Now().Sub(start)
	log.Printf("host: %v request: %v [%v] (%v)", req.Host, req.URL, sw.status, duration)
	// Log(LogEntry{
	// 	Host:       r.Host,
	// 	RemoteAddr: r.RemoteAddr,
	// 	Method:     r.Method,
	// 	RequestURI: r.RequestURI,
	// 	Proto:      r.Proto,
	// 	Status:     sw.status,
	// 	ContentLen: sw.length,
	// 	UserAgent:  r.Header.Get("User-Agent"),
	// 	Duration:   duration,
	// })
}
