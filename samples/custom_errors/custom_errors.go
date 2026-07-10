package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
	argus_dep "github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

// This sample demonstrates how to customize the error messages
// displayed when the CLI receives invalid input. Argus uses the
// Messages struct with function fields so you can fully
// localize or restyle the messages.

type GreetEntries struct {
	Name string `type:"NextArg" description:"nome da pessoa a cumprimentar"`
}

func greet(e GreetEntries, deps argus_dep.Deps) int {
	deps.Print(fmt.Sprintf("Olá, %s! Bem-vindo ao sistema.\n", e.Name))
	return 0
}

type MathEntries struct {
	A float64 `type:"Flag" identifiers:"-a" description:"primeiro número"`
	B float64 `type:"Flag" identifiers:"-b" description:"segundo número"`
}

func add(e MathEntries, deps argus_dep.Deps) int {
	deps.Print(fmt.Sprintf("%.2f + %.2f = %.2f\n", e.A, e.B, e.A+e.B))
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
	a := argus.New(native.New())

	// Portuguese error messages as an example of localization
	errosPt := argus.Messages{
		MissingFlag: func(flag, description string) string {
			if description != "" {
				return fmt.Sprintf("erro: flag obrigatória '%s' não foi informada\n  %s", flag, description)
			}
			return fmt.Sprintf("erro: flag obrigatória '%s' não foi informada", flag)
		},
		MissingArg: func(arg, description, position string) string {
			msg := fmt.Sprintf("erro: argumento obrigatório '%s' não foi informado", arg)
			if position != "" {
				msg += fmt.Sprintf(" (posição %s)", position)
			}
			if description != "" {
				msg += fmt.Sprintf("\n  %s", description)
			}
			return msg
		},
		UnknownAction: func(action string) string {
			if action == "" {
				return "erro: ação (argv[1]) não informada. Use 'greet' ou 'add'."
			}
			return fmt.Sprintf("erro: ação desconhecida '%s'. Use 'greet' ou 'add'.", action)
		},
		UnknownArg: func(arg string) string {
			return fmt.Sprintf("erro: argumento inválido '%s'.", arg)
		},
		NaN: func(flag string) string {
			return fmt.Sprintf("erro: flag '%s' não é um número.", flag)
		},
	}

	props := argus.GenerationProps{
		Messages: errosPt,
		Callbacks: []argus.Callback{
			{Starter: "greet", Callback: greet, Description: "Greet a user by name"},
			{Starter: "add", Callback: add, Description: "Add two numbers"},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("Erro de configuração:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}
