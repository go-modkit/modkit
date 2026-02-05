package validation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestValidationErrors_EmptyHasErrorsFalse(t *testing.T) {
	var errs ValidationErrors
	if errs.HasErrors() {
		t.Fatalf("expected no errors")
	}
}

func TestProblemDetails_MapsMultipleFields(t *testing.T) {
	var errs ValidationErrors
	errs.Add("name", "is required")
	errs.Add("email", "is required")

	pd := NewProblemDetails("/users", errs)

	if len(pd.InvalidParams) != 2 {
		t.Fatalf("expected 2 invalid params, got %d", len(pd.InvalidParams))
	}
	if pd.InvalidParams[0].Name != "name" || pd.InvalidParams[1].Name != "email" {
		t.Fatalf("unexpected invalid params: %+v", pd.InvalidParams)
	}
}

func TestWriteProblemDetails_WritesResponse(t *testing.T) {
	var errs ValidationErrors
	errs.Add("name", "is required")

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()

	WriteProblemDetails(rec, req, errs)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected content-type application/problem+json, got %q", ct)
	}

	var pd ProblemDetails
	if err := json.NewDecoder(rec.Body).Decode(&pd); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(pd.InvalidParams) != 1 || pd.InvalidParams[0].Name != "name" {
		t.Fatalf("unexpected invalid params: %+v", pd.InvalidParams)
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
