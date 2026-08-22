package enums

import "slices"

type Language string

const (
	English    Language = "en"
	Vietnamese Language = "vi"
)

var languages = []Language{English, Vietnamese}

func (l Language) String() string {
	return string(l)
}

func (l Language) Valid() bool {
	return slices.Contains(languages, l)
}
