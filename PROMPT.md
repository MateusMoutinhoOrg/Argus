interface mechanism:

make the deps as a interface , like these: 
~~~go 

type Deps interface {
	GetArgs() []string
	Print(s string) // if quet is true, print nothing
    SetQuiet() //set the aplication in quiet mode
}
~~~ 
