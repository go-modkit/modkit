package users

import (
	"strings"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/validation"
)

type CreateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (i CreateUserInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors
	if strings.TrimSpace(i.Name) == "" {
		errs.Add("name", "is required")
	}
	if strings.TrimSpace(i.Email) == "" {
		errs.Add("email", "is required")
	}
	return errs
}

type UpdateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (i UpdateUserInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors
	if strings.TrimSpace(i.Name) == "" {
		errs.Add("name", "is required")
	}
	if strings.TrimSpace(i.Email) == "" {
		errs.Add("email", "is required")
	}
	return errs
}
