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
