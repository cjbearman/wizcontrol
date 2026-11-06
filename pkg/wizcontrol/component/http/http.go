package http

import "fmt"

type HttpAuth struct {
	username string
	password string
}

type HttpTLS struct {
	certChain string
	key       string
}

type Http struct {
	port int
	auth *HttpAuth
	tls  *HttpTLS
}

func Build(cfg *ConfigYAML) *Http {
	if cfg == nil {
		return nil
	}

	return &Http{
		port: cfg.Port,
		tls: func() *HttpTLS {
			if cfg.CertChain != "" && cfg.Key != "" {
				return &HttpTLS{
					certChain: cfg.CertChain,
					key:       cfg.Key,
				}
			}
			return nil
		}(),
		auth: func() *HttpAuth {
			if cfg.Auth != nil {
				return &HttpAuth{
					username: cfg.Auth.Username,
					password: cfg.Auth.Password,
				}
			}
			return nil
		}(),
	}
}

func (h *Http) Auth() *HttpAuth {
	return h.auth
}

func (h *Http) TLS() *HttpTLS {
	return h.tls
}
func (h *Http) Port() int {
	return h.port
}

func (a *HttpAuth) Username() string {
	return a.username
}

func (a *HttpAuth) Password() string {
	return a.password
}

func (t *HttpTLS) CertChainAndKey() (string, string) {
	return t.certChain, t.key
}

func (h *Http) String() string {
	return fmt.Sprintf("HTTP: port=%d, auth=%t, tls=%t", h.port, h.auth != nil, h.tls != nil)
}
