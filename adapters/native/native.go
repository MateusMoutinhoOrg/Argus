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
	return deps.Deps{
		Args:  os.Args,
		Print: Print,
	}
}
