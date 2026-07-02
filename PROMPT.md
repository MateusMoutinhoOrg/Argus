change the file errors.go to menssages.go 
where instead of struct errors, it will have the struct msg
the msg will work in these way:


~~~go

type Menssages struct {
	MissingFlag  func(map[string]string) string
	MissingArg   func(map[string]string) string
	UnknowAction func(map[string]string) string
	UnknowArg    func(map[string]string) string
	NaN          func(map[string]string) string
}

var DefaultMessages = Menssages{
	MissingFlag: func(m map[string]string) string {
		return fmt.Sprintf("missing flag %s", m["flag"])
	},
	MissingArg: func(m map[string]string) string {
		return fmt.Sprintf("missing arg %s", m["arg"])
	},
	UnknowAction: func(m map[string]string) string {
		if m["action"] == nil {
			return "action (argv[1]) not provided"
		}

		return fmt.Sprintf("unknow action %s", m["action"])
	},
	UnknowArg: func(m map[string]string) string {
		return fmt.Sprintf("unknow arg %s", m["arg"])
	},
	NaN: func(m map[string]string) string {
		return fmt.Sprintf("flag %s is not a number", m["flag.1"])
	},
}




~~~

IMPORTANT: 
also add in menssages the help generator, and all the possible itens that 
are printed by the engine 