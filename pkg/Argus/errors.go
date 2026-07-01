package Argus

type Errors struct {
	MissingFlag  string
	MissingArg   string
	UnknowAction string
	UnknowArg    string
}

var DefaultErrors = Errors{
	MissingFlag:  "missing flag",
	MissingArg:   "missing arg",
	UnknowAction: "unknow action",
	UnknowArg:    "unknow arg",
}
