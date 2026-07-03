package native

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

func New() deps.Deps {
	quiet := false
	return deps.Deps{
		Args: os.Args,
		Print: func(s string) {
			if quiet {
				return
			}
			fmt.Println(s)
		},
		Quiet: &quiet,
	}
}
