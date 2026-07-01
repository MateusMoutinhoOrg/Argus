package native

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

func Print(s string) {
	fmt.Println(s)
}
func New() deps.Deps {

	var args []string
	for _, arg := range os.Args {
		args = append(args, arg)
	}
	return deps.Deps{
		Args:  os.Args,
		Print: Print,
	}
}
