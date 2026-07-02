package argus

import (
	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

type Lib struct {
	deps deps.Deps
}

func New(d deps.Deps) Lib {
	return Lib{deps: d}
}
