package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

var ServerPort int
var CertChain string
var Key string
var AuthUsername string
var AuthPassword string

const (
	FieldRemoteAddr = "remote_addr"
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
	}
	srv.Runner.Exec = srv.run

	mux.HandleFunc("GET /", intercept(srv.statusHandler))
	mux.HandleFunc("GET /scenes", intercept(srv.scenesHandler))
	mux.HandleFunc("GET /colors", intercept(srv.colorsHandler))
	mux.HandleFunc("POST /lock", intercept(srv.lockHandler))
	mux.HandleFunc("POST /addscenes", intercept(srv.addScenesHandler))
	mux.HandleFunc("POST /delscenes", intercept(srv.delScenesHandler))
	mux.HandleFunc("POST /unlock", intercept(srv.unlockHandler))
	mux.HandleFunc("POST /reset", intercept(srv.resetHandler))

	return srv

}
func (s *ServerInstance) run() {
	var err error
	if CertChain != "" && Key != "" {
		common.Logger.GetLogger("server").Infof("Starting server with TLS on port %d", ServerPort)
		err = s.server.ListenAndServeTLS(CertChain, Key)
	} else {
		common.Logger.GetLogger("server").Infof("Starting server without TLS on port %d", ServerPort)
		err = s.server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		common.Logger.GetLogger("server").Errorf("Server error: %v", err)
	}
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
