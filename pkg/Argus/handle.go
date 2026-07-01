package Argus

type Callback struct {
	Starter  string
	Callback any
}
type GenerationProps struct {
	Errors    Errors
	Callbacks []Callback
}

func (l Lib) HandleCli(props GenerationProps) int {

	//return 0 on fail
	return 0
}
