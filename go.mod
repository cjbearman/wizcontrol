module github.com/cjbearman/wizcontrol

go 1.26.0

require (
	github.com/cjbearman/logger v0.0.0-20261105180923-ca92117cd857
	github.com/cjbearman/wiz v1.0.1
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/google/uuid v1.6.0
	github.com/rickar/cal/v2 v2.1.32
	github.com/stretchr/testify v1.12.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/prometheus-community/pro-bing v0.9.1 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/cjbearman/wiz => ../wiz	