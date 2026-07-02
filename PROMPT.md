change the file errors.go to menssages.go 
where instead of struct errors, it will have the struct msg
the msg will work in these way:


~~~go




type Menssages struct {
	MissingFlag  func(string) string
	MissingArg   func(string) string
	UnknowAction func(string) string
	UnknowArg    func(string) string
	NaN          func(string) string
}

var DefaultMessages = Menssages{
	MissingFlag: func(flag string) string {
		return fmt.Sprintf("missing flag %s", flag)
	},
	MissingArg: func(arg string) string {
		return fmt.Sprintf("missing arg %s", arg)
	},
	UnknowAction: func(action string) string {
		if action == "" {
			return "action (argv[1]) not provided"
		}

		return fmt.Sprintf("unknow action %s", action)
	},
	UnknowArg: func(arg string) string {
		return fmt.Sprintf("unknow arg %s", arg)
	},
	NaN: func(flag string) string {
		return fmt.Sprintf("flag %s is not a number", flag)
	},
}




~~~

IMPORTANT: 
also add in menssages the help generator, and all the possible itens that 
are printed by the engine 
if nescessary you can create functions that recives structs as arguments 
