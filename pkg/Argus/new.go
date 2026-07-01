package Argus

import (
	Deps "github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

type Lib struct {
	deps Deps.Deps
}

func New(d Deps.Deps) Lib {
	return Lib{deps: d}
}
