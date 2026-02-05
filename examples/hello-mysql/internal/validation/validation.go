package validation

type FieldError struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type ValidationErrors struct {
	Fields []FieldError
}

func (v *ValidationErrors) Add(name, reason string) {
	v.Fields = append(v.Fields, FieldError{Name: name, Reason: reason})
}

func (v ValidationErrors) HasErrors() bool {
	return len(v.Fields) > 0
}
