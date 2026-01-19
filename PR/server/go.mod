module pr-server

go 1.19

replace pr-common => ../common

require (
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	gopkg.in/yaml.v3 v3.0.1
	pr-common v0.0.0
)
