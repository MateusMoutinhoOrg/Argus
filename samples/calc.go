package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

type NumEntries struct {
	a float64 `arg: "next" required: "true"`
	b float64 `arg: "next" required: "true"`
}

func sum(entries NumEntries) int {
	fmt.Println("sum is : ", entries.a+entries.b)
	return 0
}

func sub(entries NumEntries) int {
	fmt.Println("sub is : ", entries.a-entries.b)
	return 0
}
func mul(entries NumEntries) int {
	fmt.Println("mul is : ", entries.a*entries.b)
	return 0
}
func div(entries NumEntries) int {
	fmt.Println("div is : ", entries.a/entries.b)
	return 0
}

func main() {

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
		Errors: Argus.DefaultErrors,
		Callbacks: []Argus.Callback{
			Argus.Callback{
				Starter:  "sum",
				Callback: sum,
			},
			Argus.Callback{
				Starter:  "sub",
				Callback: sub,
			},
			Argus.Callback{
				Starter:  "mul",
				Callback: mul,
			},
			Argus.Callback{
				Starter:  "div",
				Callback: div,
			},
		},
	}

	exitCode := argus.HandleCli(props)
	os.Exit(exitCode)
}
