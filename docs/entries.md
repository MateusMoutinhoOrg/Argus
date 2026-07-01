entries of callbacks must have the type anotations to allow the parser 
indentify cli elements

### Args
Args can be a number or "next", if is set to "next" it will get the next unnused unflaged arg.

#### Args "next":
~~~go 

type NumEntries struct {
	a float64 `type: "nextArg" required: "true"`
	b float64 `type: "nextArg" required: "true"`
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
	a float64 `type: "NumArg"  position: "0"  required: "true"`
	b float64 `type: "NumArg"  position: "1"  required: "true"`
    c float64 `type: "NumArg"  position: "2"  required: "true"`
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

#### Array Args 

Array args work like slice where you can set start and end (-1 is infinity)

~~~go 

type ArrayEntries struct {
	a []int `type: "ArrayArg" start: "0" end: "2" required: "true"`
	b []int `type: "ArrayArg" start: "3" end: "-1" required: "true"`
}

func test_func(entries ArrayEntries) int {

    for _, v := range entries.a {
        fmt.Println(v)
    }
    for _, v := range entries.b {
        fmt.Println(v)
    }
	return 0
}
~~~

### Flags
