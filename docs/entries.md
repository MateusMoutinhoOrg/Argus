entries of callbacks must have the type anotations to allow the parser 
indentify cli elements

### Args
Args can be: "number","next" or "ArrayArg", if is set to "next" it will get the next unnused unflaged arg.

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
	a float64 `type: "Arg"  position: "2"  required: "true"`
	b float64 `type: "Arg"  position: "3"  required: "true"`
    c float64 `type: "Arg"  position: "4"  required: "true"`
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
	nums []int `type: "ArrayArg" start: "0" end: "2" required: "true"`
}

func test_func(entries ArrayEntries) int {

    for _, n := range entries.nums {
        fmt.Println(n)
    }
	return 0
}
~~~

bash sample:

### Flags
flags can also be retrived 

~~~go 

type NumEntries struct {
	a float64 `type: "flag" identifiers:"-a,--a" required:"true"`
	b float64 `type: "flag" identifiers:"-b,--b" required:"true"`
}

func sum(entries NumEntries) int {

    fmt.Println(entries.a + entries.b) 
	return 0
}
~~~
in cli :
~~~sh
calc add --a 10  --b 20 

~~~

Flags can also be arrays

~~~go 

type NumEntries struct {
	nums []float64 `type: "ArrayFlag" identifiers:"-a,--a" required: "true" min_size:"1" max_size:"-1"`
}

func test(entries NumEntries) int {
  
  for _,n := range entries.nums {
      fmt.Println(n)
  }
  
  return 0 
}
~~~
in cli :
~~~
test -a 1 -a 2 -a 3 -a 4 -a 5

~~~

