improve the msgs generated
1. add a description element on the callback entries like:
~~~go
	Depth int    `type:"Flag" identifiers:"--depth" required:"false" description:"specifies the depth of the clone"`
~~~

2. format the help to be more professional and complete
3. format the errors to indicate each flag is missing  and its description
3. format the errors to indicate each args its missing and it possitions 