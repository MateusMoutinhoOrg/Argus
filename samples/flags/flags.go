package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
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

func serve(e ServeEntries) int {
	scheme := "http"
	if e.TLS {
		scheme = "https"
	}
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("  Server starting…\n")
	fmt.Printf("  Address:   %s://%s:%d\n", scheme, e.Host, e.Port)
	fmt.Printf("  TLS:       %v\n", e.TLS)
	fmt.Printf("  Log level: %s\n", e.LogLevel)
	fmt.Println(strings.Repeat("─", 40))
	return 0
}

// StatusEntries demonstrates a command with no required flags —
// only boolean presence flags and optional flags.
type StatusEntries struct {
	Verbose bool   `type:"Flag" identifiers:"-v,--verbose"`
	Format  string `type:"Flag" identifiers:"-f,--format" default:"text"`
}

func status(e StatusEntries) int {
	fmt.Printf("Status (format=%s, verbose=%v)\n", e.Format, e.Verbose)
	if e.Verbose {
		fmt.Println("  PID:    12345")
		fmt.Println("  Uptime: 3h42m")
		fmt.Println("  Memory: 128 MB")
	} else {
		fmt.Println("  running ✓")
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

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{Starter: "serve", Callback: serve, Description: "Serve the application"},
			{Starter: "status", Callback: status, Description: "Get the application status"},
		},
	}

	exitCode, err := argus.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}
