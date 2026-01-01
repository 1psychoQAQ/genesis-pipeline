package filter

import (
	"testing"
	"time"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
)

func TestFilter_Level1_Accepted(t *testing.T) {
	f := NewFilter()

	paper := model.Paper{
		ID:       "2301.00001v1",
		Title:    "Test Paper",
		Abstract: "We conduct extensive experiments and evaluation on benchmark datasets.",
		Comments: "Accepted at ICML 2024",
	}

	result := f.evaluate(paper)

	if !result.PassedLevel1 {
		t.Error("Paper with accepted signal should pass Level 1")
	}
}

func TestFilter_Level1_DOI(t *testing.T) {
	f := NewFilter()

	paper := model.Paper{
		ID:       "2301.00001v1",
		Title:    "Test Paper",
		Abstract: "Our experiments show significant improvements on the evaluation benchmark.",
		DOI:      "10.1234/example",
	}

	result := f.evaluate(paper)

	if !result.PassedLevel1 {
		t.Error("Paper with DOI should pass Level 1")
	}
}

func TestFilter_Level1_StrongEvidence(t *testing.T) {
	f := NewFilter()

	paper := model.Paper{
		ID:       "2301.00001v1",
		Title:    "Test Paper",
		Abstract: "We perform ablation experiments on benchmark datasets with multiple metrics for evaluation.",
	}

	result := f.evaluate(paper)

	if !result.PassedLevel1 {
		t.Error("Paper with strong evidence (>=3 keywords) should pass Level 1")
	}
}

func TestFilter_Level1_Fail_NoEvaluation(t *testing.T) {
	f := NewFilter()

	paper := model.Paper{
		ID:       "2301.00001v1",
		Title:    "A New Framework",
		Abstract: "We propose a novel framework for understanding complex systems.",
		Comments: "Accepted at NeurIPS 2024",
	}

	result := f.evaluate(paper)

	if result.PassedLevel1 {
		t.Error("Paper without evaluation keywords should fail Level 1")
	}
}

func TestFilter_Scoring(t *testing.T) {
	f := NewFilter()

	paper := model.Paper{
		ID:       "2301.00001v2",
		Title:    "Comprehensive Evaluation",
		Abstract: "We conduct ablation experiments on benchmark datasets with baseline comparisons. We discuss the limitations of our approach.",
		Comments: "Accepted at ICML 2024. Code: https://github.com/user/repo",
		DOI:      "10.1234/example",
	}

	result := f.evaluate(paper)

	// Expected: +30 (accepted) +20 (DOI) +15 (>=3 eval) +10 (ablation/baseline) +10 (dataset/benchmark) +10 (code) +5 (limitation) +5 (v2) = 105 -> capped at 100
	if result.Score < 90 {
		t.Errorf("Expected high score (>=90), got %d", result.Score)
	}
}

func TestFilter_Scoring_Negative(t *testing.T) {
	f := NewFilter()

	paper := model.Paper{
		ID:       "2301.00001v1",
		Title:    "Revolutionary Framework",
		Abstract: "This is a groundbreaking framework that changes everything.",
	}

	result := f.evaluate(paper)

	// Should have negative modifiers
	if result.Score >= 50 {
		t.Errorf("Paper with hype words and no evaluation should have low score, got %d", result.Score)
	}
}

func TestFilter_FilterPassed(t *testing.T) {
	f := NewFilter()
	f.MinScore = 30 // Lower threshold for testing

	papers := []model.Paper{
		{
			ID:       "good-paper",
			Title:    "Good Paper",
			Abstract: "We conduct experiments and evaluation on benchmark datasets with ablation studies.",
			Comments: "Accepted at ICML",
		},
		{
			ID:       "bad-paper",
			Title:    "Bad Paper",
			Abstract: "This is a framework.",
		},
	}

	passed := f.FilterPassed(papers)

	if len(passed) != 1 {
		t.Errorf("Expected 1 paper to pass, got %d", len(passed))
	}

	if len(passed) > 0 && passed[0].ID != "good-paper" {
		t.Errorf("Expected good-paper to pass, got %s", passed[0].ID)
	}
}

func TestPaperVersion(t *testing.T) {
	tests := []struct {
		id      string
		version int
	}{
		{"2301.00001v1", 1},
		{"2301.00001v2", 2},
		{"2301.00001v10", 10},
		{"2301.00001", 1},
		{"cs/0001001v3", 3},
	}

	for _, tc := range tests {
		paper := model.Paper{ID: tc.id, UpdatedAt: time.Now()}
		if got := paper.Version(); got != tc.version {
			t.Errorf("Paper{ID: %q}.Version() = %d, want %d", tc.id, got, tc.version)
		}
	}
}
