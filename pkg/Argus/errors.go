package Argus

type Errors struct {
	MissingFlag  string
	MissingArg   string
	UnknowAction string
	UnknowArg    string
}

var DefaultErrors = Errors{
	MissingFlag:  "missing flag :%s",
	MissingArg:   "missing arg :%s",
	UnknowAction: "unknow action :%s",
	UnknowArg:    "unknow arg :%s",
}
