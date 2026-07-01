package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

type NumEntries struct {
	nuns []float64 `type "ArrayArg" required: "true"`
}

func sum(entries NumEntries) int {
	sum := 0.0
	for _, v := range entries.nuns {
		sum += v
	}
	fmt.Println("sum is : ", sum)
	return 0
}

func sub(entries NumEntries) int {
	sub := 0.0
	for _, v := range entries.nuns {
		sub -= v
	}
	fmt.Println("sub is : ", sub)
	return 0
}
func mul(entries NumEntries) int {
	mul := 0.0
	for _, v := range entries.nuns {
		mul *= v
	}
	fmt.Println("mul is : ", mul)
	return 0
}
func div(entries NumEntries) int {
	div := entries.nuns[0]
	for _, v := range entries.nuns {
		div /= v
	}
	fmt.Println("div is : ", div)
	return 0
}

func main() {

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
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
