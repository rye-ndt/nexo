package enums

import "slices"

type ParamType string

const (
	TextParam     ParamType = "text"
	TextareaParam ParamType = "textarea"
	NumberParam   ParamType = "number"
	BooleanParam  ParamType = "boolean"
	SelectParam   ParamType = "select"
	FileParam     ParamType = "file"
)

var paramTypes = []ParamType{
	TextParam,
	TextareaParam,
	NumberParam,
	BooleanParam,
	SelectParam,
	FileParam,
}

func ParamTypes() []ParamType {
	return slices.Clone(paramTypes)
}

func (p ParamType) Valid() bool {
	return slices.Contains(paramTypes, p)
}

func (p ParamType) String() string {
	return string(p)
}
