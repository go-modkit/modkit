package validation

import "testing"

func TestValidationErrors_AddAndHasErrors(t *testing.T) {
	var errs ValidationErrors
	errs.Add("name", "is required")

	if !errs.HasErrors() {
		t.Fatalf("expected errors")
	}
	if len(errs.Fields) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs.Fields))
	}
	if errs.Fields[0].Name != "name" || errs.Fields[0].Reason != "is required" {
		t.Fatalf("unexpected field error: %+v", errs.Fields[0])
	}
}

func TestProblemDetails_FromValidationErrors(t *testing.T) {
	err := ValidationErrors{}
	err.Add("email", "is required")

	pd := NewProblemDetails("/users", err)

	if pd.Status != 400 {
		t.Fatalf("expected status 400, got %d", pd.Status)
	}
	if len(pd.InvalidParams) != 1 {
		t.Fatalf("expected 1 invalid param, got %d", len(pd.InvalidParams))
	}
	if pd.InvalidParams[0].Name != "email" {
		t.Fatalf("unexpected invalid param: %+v", pd.InvalidParams[0])
	}
}
