package enums

import "slices"

type InputType string

const (
	TextInput     InputType = "text"
	TextareaInput InputType = "textarea"
	NumberInput   InputType = "number"
	BooleanInput  InputType = "boolean"
	SelectInput   InputType = "select"
	MultiInput    InputType = "multiselect"
	FileInput     InputType = "file"
)

var inputTypes = []InputType{
	TextInput,
	TextareaInput,
	NumberInput,
	BooleanInput,
	SelectInput,
	MultiInput,
	FileInput,
}

func InputTypes() []InputType {
	return slices.Clone(inputTypes)
}

func (p InputType) Valid() bool {
	return slices.Contains(inputTypes, p)
}

func (p InputType) String() string {
	return string(p)
}
