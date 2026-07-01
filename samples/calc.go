package main

import (
	"fmt"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

type FirstActionEntries struct {
	a string `arg: "next"`
	b string `arg: "next"`
}

func sum(entries FirstActionEntries) {
	fmt.Println("sum is : ", entries.a+entries.b)
}

func main() {

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{}

	argus.Generate(props)

}
