package validation

import (
	"fmt"
	"net/http"
)

type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type ProblemDetails struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance"`
	InvalidParams []InvalidParam `json:"invalidParams"`
}

func NewProblemDetails(instance string, errs ValidationErrors) ProblemDetails {
	invalid := make([]InvalidParam, 0, len(errs.Fields))
	for _, f := range errs.Fields {
		invalid = append(invalid, InvalidParam{Name: f.Name, Reason: f.Reason})
	}

	return ProblemDetails{
		Type:          fmt.Sprintf("https://httpstatuses.com/%d", http.StatusBadRequest),
		Title:         http.StatusText(http.StatusBadRequest),
		Status:        http.StatusBadRequest,
		Detail:        "validation failed",
		Instance:      instance,
		InvalidParams: invalid,
	}
}
