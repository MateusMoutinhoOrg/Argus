entries of callbacks must have the type anotations to allow the parser 
indentify cli elements

### Args
Args can be a number or "next", if is set to "next" it will get the next unnused unflaged arg.

#### Args "next":
~~~go 

type NumEntries struct {
	a float64 `arg: "next" required: "true"`
	b float64 `arg: "next" required: "true"`
}

func sum(entries NumEntries) int {

    fmt.Println(entries.a + entries.b)
	return 0
}
~~~
in cli :
~~~sh
calc add 10 20 

~~~

#### Numerical Args:

~~~go 

type NumEntries struct {
	a float64 `arg: "0" required: "true"`
	b float64 `arg: "1" required: "true"`
    c float64 `arg: "2" required: "true"`
}

func sum(entries NumEntries) int {

    fmt.Println(entries.a + entries.b + entries.c)
	return 0
}
~~~
in cli :
~~~sh
calc add 10 20 39
~~~



