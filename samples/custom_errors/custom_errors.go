package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

// This sample demonstrates how to customize the error messages
// displayed when the CLI receives invalid input. Argus uses the
// Errors struct with Go format strings (%s) so you can fully
// localize or restyle the messages.

type GreetEntries struct {
	Name string `type:"NextArg"`
}

func greet(e GreetEntries) int {
	fmt.Printf("Olá, %s! Bem-vindo ao sistema.\n", e.Name)
	return 0
}

type MathEntries struct {
	A float64 `type:"Flag" identifiers:"-a"`
	B float64 `type:"Flag" identifiers:"-b"`
}

func add(e MathEntries) int {
	fmt.Printf("%.2f + %.2f = %.2f\n", e.A, e.B, e.A+e.B)
	return 0
}

// Usage:
//
//	go run samples/custom_errors/custom_errors.go greet Mateus
//	go run samples/custom_errors/custom_errors.go add -a 10 -b 20
//
//	# Trigger custom error messages:
//	go run samples/custom_errors/custom_errors.go unknown
//	go run samples/custom_errors/custom_errors.go greet
//	go run samples/custom_errors/custom_errors.go add -a 10
func main() {

	argus := Argus.New(native.New())

	// Portuguese error messages as an example of localization
	errosPt := Argus.Errors{
		MissingFlag:  "Erro: a flag obrigatória '%s' não foi informada.",
		MissingArg:   "Erro: o argumento obrigatório '%s' não foi informado.",
		UnknowAction: "Erro: ação desconhecida '%s'. Use 'greet' ou 'add'.",
		UnknowArg:    "Erro: argumento inválido '%s'.",
	}

	props := Argus.GenerationProps{
		Errors: errosPt,
		Callbacks: []Argus.Callback{
			{Starter: "greet", Callback: greet},
			{Starter: "add", Callback: add},
		},
	}

	exitCode, err := argus.HandleCli(props)
	if err != nil {
		fmt.Println("Erro de configuração:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}
