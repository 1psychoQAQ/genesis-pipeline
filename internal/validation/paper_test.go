package validation

import (
	"testing"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
)

func TestValidatePaper_Valid(t *testing.T) {
	paper := model.Paper{
		ID:        "2301.00001",
		Title:     "Test Paper",
		Abstract:  "Test abstract",
		Authors:   []string{"John Doe"},
		UpdatedAt: time.Now(),
	}

	errs := ValidatePaper(paper)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidatePaper_EmptyID(t *testing.T) {
	paper := model.Paper{
		ID:        "",
		Title:     "Test Paper",
		Authors:   []string{"John Doe"},
		UpdatedAt: time.Now(),
	}

	errs := ValidatePaper(paper)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Field != "ID" {
		t.Errorf("expected ID field error, got %s", errs[0].Field)
	}
}

func TestValidatePaper_NoAuthors(t *testing.T) {
	paper := model.Paper{
		ID:        "2301.00001",
		Title:     "Test Paper",
		Authors:   []string{},
		UpdatedAt: time.Now(),
	}

	errs := ValidatePaper(paper)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Field != "Authors" {
		t.Errorf("expected Authors field error, got %s", errs[0].Field)
	}
}

func TestValidatePapers_Batch(t *testing.T) {
	papers := []model.Paper{
		{ID: "1", Title: "Valid", Authors: []string{"A"}, UpdatedAt: time.Now()},
		{ID: "", Title: "Invalid", Authors: []string{"B"}, UpdatedAt: time.Now()},
		{ID: "2", Title: "Also Valid", Authors: []string{"C"}, UpdatedAt: time.Now()},
	}

	result := ValidatePapers(papers)

	if result.Valid != 2 {
		t.Errorf("expected 2 valid, got %d", result.Valid)
	}
	if result.Invalid != 1 {
		t.Errorf("expected 1 invalid, got %d", result.Invalid)
	}
}

func TestIsValid(t *testing.T) {
	valid := model.Paper{
		ID:        "2301.00001",
		Title:     "Test",
		Authors:   []string{"Author"},
		UpdatedAt: time.Now(),
	}

	invalid := model.Paper{
		ID:    "",
		Title: "Test",
	}

	if !IsValid(valid) {
		t.Error("expected valid paper to be valid")
	}

	if IsValid(invalid) {
		t.Error("expected invalid paper to be invalid")
	}
}

func BenchmarkValidatePaper(b *testing.B) {
	paper := model.Paper{
		ID:        "2301.00001",
		Title:     "Test Paper Title",
		Abstract:  "This is a test abstract for benchmarking purposes.",
		Authors:   []string{"John Doe", "Jane Smith"},
		UpdatedAt: time.Now(),
	}

	for i := 0; i < b.N; i++ {
		ValidatePaper(paper)
	}
}
