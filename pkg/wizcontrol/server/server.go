package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/config"
)

const (
	FieldRemoteAddr = "remote_addr"
)

// Instance holds the singleton instance of the server.
var Instance *Server

// Server represents the HTTP server for handling WizControl requests.
type Server struct {
	*common.BaseRunner
	mux    *http.ServeMux
	server *http.Server
	cnf    *config.Config
}

// NewServer creates and returns a new instance of the Server with the provided configuration.
func NewServer(cnf *config.Config) *Server {
	mux := http.NewServeMux()
	srv := &Server{
		BaseRunner: common.NewRunner("server"),
		mux:        mux,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cnf.HTTP.Port()),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12, // Set minimum TLS version
				MaxVersion: tls.VersionTLS13, // Set maximum TLS version
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				},
			},
		},
		cnf: cnf,
	}
	srv.Exec = srv.run

	mux.HandleFunc("GET /", srv.intercept(srv.statusHandler))
	mux.HandleFunc("GET /scenegroups", srv.intercept(srv.scenegroupsHandler))
	mux.HandleFunc("GET /lightgroups", srv.intercept(srv.lightgroupsHandler))
	mux.HandleFunc("GET /colors", srv.intercept(srv.colorsHandler))
	mux.HandleFunc("GET /targets", srv.intercept(srv.targetsHandler))
	mux.HandleFunc("POST /lock", srv.intercept(srv.lockHandler))
	mux.HandleFunc("POST /addscenes", srv.intercept(srv.addScenesHandler))
	mux.HandleFunc("POST /delscenes", srv.intercept(srv.delScenesHandler))
	mux.HandleFunc("POST /unlock", srv.intercept(srv.unlockHandler))
	mux.HandleFunc("POST /reset", srv.intercept(srv.resetHandler))
	mux.HandleFunc("POST /pause", srv.intercept(srv.pauseHandler))
	mux.HandleFunc("POST /unpause", srv.intercept(srv.unpauseHandler))

	return srv

}

// run starts the HTTP server and handles TLS configuration if provided. It blocks until the server stops.
func (s *Server) run() {
	logger := common.Logger.GetLogger("server")
	var err error

	cert, key := "", ""
	if s.cnf.HTTP.TLS() != nil {
		cert, key = s.cnf.HTTP.TLS().CertChainAndKey()
	}

	if cert != "" && key != "" {
		common.Logger.GetLogger("server").Infof("Starting server with TLS on port %d", s.cnf.HTTP.Port())
		err = s.server.ListenAndServeTLS(cert, key)
	} else {
		common.Logger.GetLogger("server").Infof("Starting server without TLS on port %d", s.cnf.HTTP.Port())
		err = s.server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		common.Logger.GetLogger("server").Errorf("Server error: %v", err)
	}

	logger.Info("Server stopped")
	close(s.Closed)
}

// Stop gracefully shuts down the HTTP server with a timeout of 5 seconds.
func (s *Server) Stop() {
	logger := common.Logger.GetLogger("server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		logger.Errorf("Error during server shutdown: %v", err)
	}
	logger.Info("Server indicates shutdown complete")
}
