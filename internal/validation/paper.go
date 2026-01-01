package validation

import (
	"errors"
	"strings"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
)

// ValidationError represents a paper validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidationResult holds the result of validating papers.
type ValidationResult struct {
	Valid   int
	Invalid int
	Errors  []ValidationError
}

// ValidatePaper validates a single paper and returns any errors.
func ValidatePaper(p model.Paper) []ValidationError {
	var errs []ValidationError

	if strings.TrimSpace(p.ID) == "" {
		errs = append(errs, ValidationError{Field: "ID", Message: "cannot be empty"})
	}

	if strings.TrimSpace(p.Title) == "" {
		errs = append(errs, ValidationError{Field: "Title", Message: "cannot be empty"})
	}

	if len(p.Authors) == 0 {
		errs = append(errs, ValidationError{Field: "Authors", Message: "must have at least one author"})
	}

	if p.UpdatedAt.IsZero() {
		errs = append(errs, ValidationError{Field: "UpdatedAt", Message: "cannot be zero"})
	}

	return errs
}

// ValidatePapers validates a batch of papers and returns a summary.
func ValidatePapers(papers []model.Paper) ValidationResult {
	result := ValidationResult{}

	for _, p := range papers {
		errs := ValidatePaper(p)
		if len(errs) > 0 {
			result.Invalid++
			result.Errors = append(result.Errors, errs...)
		} else {
			result.Valid++
		}
	}

	return result
}

// IsValid returns true if the paper passes all validations.
func IsValid(p model.Paper) bool {
	return len(ValidatePaper(p)) == 0
}

// ErrInvalidPaper is returned when a paper fails validation.
var ErrInvalidPaper = errors.New("invalid paper")
