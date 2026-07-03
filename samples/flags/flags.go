package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

// ServeFlags demonstrates:
//   - Required string flag (Host)
//   - Optional int flag with default (Port)
//   - Boolean presence flag (TLS)
//   - Optional string flag with default (LogLevel)
//
// Every field lives in a Flags sub-struct with an `identifiers` tag; Argus
// infers it's a Flag (or ArrayFlag, for slice fields) automatically.
type ServeFlags struct {
	Host     string `identifiers:"-h,--host" description:"the host address to bind to"`
	Port     int    `identifiers:"-p,--port" default:"8080" description:"the port number to listen on (default: 8080)"`
	TLS      bool   `identifiers:"--tls" description:"enable TLS/HTTPS"`
	LogLevel string `identifiers:"-l,--log-level" default:"info" description:"logging level: debug, info, warn, error (default: info)"`
}

type ServeEntries struct {
	Flags ServeFlags
}

func serve(e ServeEntries) int {
	scheme := "http"
	if e.Flags.TLS {
		scheme = "https"
	}
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("  Server starting…\n")
	fmt.Printf("  Address:   %s://%s:%d\n", scheme, e.Flags.Host, e.Flags.Port)
	fmt.Printf("  TLS:       %v\n", e.Flags.TLS)
	fmt.Printf("  Log level: %s\n", e.Flags.LogLevel)
	fmt.Println(strings.Repeat("─", 40))
	return 0
}

// StatusFlags demonstrates a command with no required flags —
// only boolean presence flags and optional flags.
type StatusFlags struct {
	Verbose bool   `identifiers:"-v,--verbose"`
	Format  string `identifiers:"-f,--format" default:"text"`
}

type StatusEntries struct {
	Flags StatusFlags
}

func status(e StatusEntries) int {
	fmt.Printf("Status (format=%s, verbose=%v)\n", e.Flags.Format, e.Flags.Verbose)
	if e.Flags.Verbose {
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
