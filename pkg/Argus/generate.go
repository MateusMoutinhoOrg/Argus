package Argus

type Callback struct {
	Starter  string
	Callback func(entries any)
}
type GenerationProps struct {
	Errors    Errors
	Callbacks []Callback
}

func (l Lib) Generate(props GenerationProps) int {

	//return 0 on fail
	return 0
}
