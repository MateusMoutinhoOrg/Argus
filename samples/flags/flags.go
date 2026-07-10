package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
	argus_dep "github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

// ServeEntries demonstrates:
//   - Required string flag (Host)
//   - Optional int flag with default (Port)
//   - Boolean presence flag (TLS)
//   - Optional string flag with default (LogLevel)
type ServeEntries struct {
	Host     string `type:"Flag" identifiers:"-h,--host" description:"the host address to bind to"`
	Port     int    `type:"Flag" identifiers:"-p,--port" default:"8080" description:"the port number to listen on (default: 8080)"`
	TLS      bool   `type:"Flag" identifiers:"--tls" description:"enable TLS/HTTPS"`
	LogLevel string `type:"Flag" identifiers:"-l,--log-level" default:"info" description:"logging level: debug, info, warn, error (default: info)"`
}

func serve(e ServeEntries, deps argus_dep.Deps) int {
	scheme := "http"
	if e.TLS {
		scheme = "https"
	}
	deps.Print(strings.Repeat("─", 40) + "\n")
	deps.Print("  Server starting…\n")
	deps.Print(fmt.Sprintf("  Address:   %s://%s:%d\n", scheme, e.Host, e.Port))
	deps.Print(fmt.Sprintf("  TLS:       %v\n", e.TLS))
	deps.Print(fmt.Sprintf("  Log level: %s\n", e.LogLevel))
	deps.Print(strings.Repeat("─", 40) + "\n")
	return 0
}

// StatusEntries demonstrates a command with no required flags —
// only boolean presence flags and optional flags.
type StatusEntries struct {
	Verbose bool   `type:"Flag" identifiers:"-v,--verbose"`
	Format  string `type:"Flag" identifiers:"-f,--format" default:"text"`
}

func status(e StatusEntries, deps argus_dep.Deps) int {
	deps.Print(fmt.Sprintf("Status (format=%s, verbose=%v)\n", e.Format, e.Verbose))
	if e.Verbose {
		deps.Print("  PID:    12345\n")
		deps.Print("  Uptime: 3h42m\n")
		deps.Print("  Memory: 128 MB\n")
	} else {
		deps.Print("  running ✓\n")
	}
	return 0
}

// Usage:
//
//	go run samples/flags/flags.go serve --host 0.0.0.0
//	go run samples/flags/flags.go serve --host 0.0.0.0 -p 9090 --tls
//	go run samples/flags/flags.go serve -h 127.0.0.1 --log-level debug
//	go run samples/flags/flags.go status
//	go run samples/flags/flags.go status --verbose --format json
func main() {
	a := argus.New(native.New())

	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{Starter: "serve", Callback: serve, Description: "Serve the application"},
			{Starter: "status", Callback: status, Description: "Get the application status"},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}
