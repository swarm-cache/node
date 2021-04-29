package server

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Server template
type Server struct {
	Instance *http.Server
	Upgrader *websocket.Upgrader
	Mux      *http.ServeMux
}

// Returns a brand new server instance with WS upgrader
func Init() *Server {
	mux := http.NewServeMux()

	return &Server{
		Mux: mux,
		Instance: &http.Server{
			Handler: mux,
		},
		Upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}
