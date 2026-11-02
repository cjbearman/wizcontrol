package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cjbearman/wizcontrol2/pkg/wizcontrol/common"
)

var ServerPort int

const (
	FieldRemoteAddr   = "remote_addr"
	FieldOperation    = "operation"
	FieldScene        = "scene"
	FieldLight        = "light"
	FieldLightOrGroup = "light_or_group"
	FieldLights       = "lights"
	FieldDuration     = "duration"
	FieldColor        = "color"
	FieldBrightness   = "brightness"
)

var Server *ServerInstance

type ServerInstance struct {
	*common.Runner
	mux    *http.ServeMux
	server *http.Server
}

func NewServer() *ServerInstance {
	mux := http.NewServeMux()
	srv := &ServerInstance{
		Runner: common.NewRunner(),
		mux:    mux,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", ServerPort),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
	srv.Runner.Exec = srv.run

	mux.HandleFunc("GET /", srv.statusHandler)
	mux.HandleFunc("POST /day", srv.dayHandler)
	// mux.HandleFunc("POST /set/{light}", srv.setHandler)
	// mux.HandleFunc("POST /unlock/{light}", srv.unlockHandler)
	// mux.HandleFunc("POST /reset", srv.resetHandler)

	return srv

}
func (s *ServerInstance) run() {
	common.Logger.GetLogger("server").Info("Starting server")
	s.server.ListenAndServe()
	close(s.Closed)
}
func (s *ServerInstance) Stop() {
	logger := common.Logger.GetLogger("server")
	logger.Info("Stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
	logger.Info("Web server stopped")
}

func nocache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
