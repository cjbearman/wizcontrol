package http

import "fmt"

// HttpAuth represents HTTP authentication credentials
type HttpAuth struct {
	username string
	password string
}

// HttpTLS represents HTTP TLS configuration
type HttpTLS struct {
	certChain string
	key       string
}

// Http represents the HTTP component configuration
type Http struct {
	port int
	auth *HttpAuth
	tls  *HttpTLS
}

// Build returns an HTTP configuration instance from the given ConfigYAML
// Will return nil if HTTP is not configured for use
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

// Auth returns the HTTP authentication configuration, or nil if not configured
func (h *Http) Auth() *HttpAuth {
	return h.auth
}

// TLS returns the HTTP TLS configuration, or nil if not configured
func (h *Http) TLS() *HttpTLS {
	return h.tls
}

// Port returns the HTTP port
func (h *Http) Port() int {
	return h.port
}

// Username returns the HTTP authentication username
func (a *HttpAuth) Username() string {
	return a.username
}

// Password returns the HTTP authentication password
func (a *HttpAuth) Password() string {
	return a.password
}

// CertChainAndKey returns the HTTP TLS certificate chain and key
func (t *HttpTLS) CertChainAndKey() (string, string) {
	return t.certChain, t.key
}

// String returns a string representation of the HTTP configuration
func (h *Http) String() string {
	return fmt.Sprintf("HTTP: port=%d, auth=%t, tls=%t", h.port, h.auth != nil, h.tls != nil)
}
