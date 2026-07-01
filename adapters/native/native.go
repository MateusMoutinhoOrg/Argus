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

	var args []string
	for _, arg := range os.Args {
		args = append(args, arg)
	}
	return Deps.Deps{
		Args:  args,
		Print: Print,
	}
}
