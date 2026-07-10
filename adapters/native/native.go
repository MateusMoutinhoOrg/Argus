package native

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

type nativeDeps struct {
	args  []string
	quiet bool
}

func (n *nativeDeps) GetArgs() []string {
	return n.args
}

func (n *nativeDeps) Print(s string) {
	if !n.quiet {
		fmt.Println(s)
	}
}

func (n *nativeDeps) SetQuiet() {
	n.quiet = true
}

func New() deps.Deps {
	return &nativeDeps{
		args: os.Args,
	}
}
