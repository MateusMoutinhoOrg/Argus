package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

type NumEntries struct {
	Nums []float64 `type:"ArrayArg" start:"0" end:"-1" min_size:"1"`
}

func sum(entries NumEntries) int {
	result := 0.0
	for _, v := range entries.Nums {
		result += v
	}
	fmt.Println("sum is : ", result)
	return 0
}

func sub(entries NumEntries) int {
	if len(entries.Nums) == 0 {
		return 1
	}
	result := entries.Nums[0]
	for _, v := range entries.Nums[1:] {
		result -= v
	}
	fmt.Println("sub is : ", result)
	return 0
}
func mul(entries NumEntries) int {
	result := 1.0
	for _, v := range entries.Nums {
		result *= v
	}
	fmt.Println("mul is : ", result)
	return 0
}
func div(entries NumEntries) int {
	if len(entries.Nums) == 0 {
		return 1
	}
	result := entries.Nums[0]
	for _, v := range entries.Nums[1:] {
		if v == 0 {
			fmt.Println("error: division by zero")
			return 1
		}
		result /= v
	}
	fmt.Println("div is : ", result)
	return 0
}

func main() {

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{
				Starter:  "sum",
				Callback: sum,
			},
			{
				Starter:  "sub",
				Callback: sub,
			},
			{
				Starter:  "mul",
				Callback: mul,
			},
			{
				Starter:  "div",
				Callback: div,
			},
		},
	}

	exitCode := argus.HandleCli(props)
	os.Exit(exitCode)
}
