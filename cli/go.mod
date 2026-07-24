module github.com/zeitlos/lucity/cli

go 1.26

require (
	github.com/coder/websocket v1.8.15
	github.com/modelcontextprotocol/go-sdk v1.6.1
	github.com/zeitlos/lucity/pkg v0.0.0-20260706084318-3ef7184f18d8
)

replace github.com/zeitlos/lucity/pkg => ../pkg

require (
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)
