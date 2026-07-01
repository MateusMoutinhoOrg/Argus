package main

import (
	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

func main() {

	argus := Argus.New(native.New())

}
