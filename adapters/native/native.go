package native

import (
	"fmt"
	"os"

	Deps "github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

func Print(s string) {
	fmt.Println(s)
}
func New() Deps.Deps {

	return Deps.Deps{
		Args:  os.Args,
		Print: Print,
	}
}
