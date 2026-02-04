package mysql

import "testing"

func TestOpen_RejectsEmptyDSN(t *testing.T) {
	_, err := Open("")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestOpen_AcceptsDSN(t *testing.T) {
	_, err := Open("user:pass@tcp(localhost:3306)/app")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
