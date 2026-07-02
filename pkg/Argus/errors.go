package Argus

type Errors struct {
	MissingFlag  string
	MissingArg   string
	UnknowAction string
	UnknowArg    string
	NaN          string
}

var DefaultErrors = Errors{
	MissingFlag:  "missing flag :%s",
	MissingArg:   "missing arg :%s",
	UnknowAction: "unknow action :%s",
	UnknowArg:    "unknow arg :%s",
	NaN:          "flag %s is not a number",
}
