package filter

import (
	"regexp"
	"strings"

	"github.com/1psychoQAQ/genesis-pipeline/internal/model"
)

// Keyword patterns for filtering
var (
	// Strong signals - acceptance/publication
	acceptedPattern = regexp.MustCompile(`(?i)(accepted|to appear|camera[- ]?ready|proceedings)`)

	// Evaluation keywords
	evaluationKeywords = []string{
		"evaluation", "experiment", "benchmark", "ablation",
		"baseline", "dataset", "metric",
	}

	// Reproducibility keywords
	reproducibilityKeywords = []string{
		"dataset", "benchmark", "reproduce", "replication",
		"artifact", "supplementary",
	}

	// Risk/hype keywords
	hypeKeywords = []string{
		"revolutionary", "groundbreaking", "first ever", "first-ever",
	}

	// Framework-only keywords (without evaluation = risky)
	frameworkKeywords = []string{
		"framework", "perspective",
	}

	// Limitation keywords (good sign)
	limitationKeywords = []string{
		"limitation", "assumption", "constraint",
	}

	// Code repository pattern
	codeRepoPattern = regexp.MustCompile(`https?://(github\.com|gitlab\.com)/\S+`)
)

// Filter applies quality filtering to papers.
type Filter struct {
	MinScore int // Minimum score to pass (default: 60)
}

// NewFilter creates a new filter with default settings.
func NewFilter() *Filter {
	return &Filter{MinScore: 60}
}

// FilterResult contains the filtering outcome for a paper.
type FilterResult struct {
	Paper        model.Paper
	PassedLevel1 bool
	Score        int
	Details      []string
}

// Apply filters papers and returns results.
func (f *Filter) Apply(papers []model.Paper) []FilterResult {
	results := make([]FilterResult, 0, len(papers))

	for _, paper := range papers {
		result := f.evaluate(paper)
		results = append(results, result)
	}

	return results
}

// FilterPassed returns only papers that passed both levels.
func (f *Filter) FilterPassed(papers []model.Paper) []model.Paper {
	passed := make([]model.Paper, 0)

	for _, paper := range papers {
		result := f.evaluate(paper)
		if result.PassedLevel1 && result.Score >= f.MinScore {
			paper.Score = result.Score
			paper.ScoreDetails = result.Details
			passed = append(passed, paper)
		}
	}

	return passed
}

func (f *Filter) evaluate(paper model.Paper) FilterResult {
	result := FilterResult{Paper: paper}

	// Count evaluation keywords in abstract
	evalCount := countKeywords(paper.Abstract, evaluationKeywords)

	// Level 1: Hard gate
	hasAcceptedSignal := acceptedPattern.MatchString(paper.Comments)
	hasDOI := paper.DOI != ""
	hasJournalRef := paper.JournalRef != ""
	hasStrongEvidence := evalCount >= 3

	// Must satisfy at least one strong signal
	hasStrongSignal := hasAcceptedSignal || hasDOI || hasJournalRef || hasStrongEvidence

	// AND must have at least 2 evaluation keywords
	hasMinEvaluation := evalCount >= 2

	result.PassedLevel1 = hasStrongSignal && hasMinEvaluation

	// Level 2: Scoring
	score := 0
	details := make([]string, 0)

	// Positive signals
	if hasAcceptedSignal {
		score += 30
		details = append(details, "+30 接收信号")
	}

	if hasDOI || hasJournalRef {
		score += 20
		details = append(details, "+20 DOI/期刊引用")
	}

	if evalCount >= 3 {
		score += 15
		details = append(details, "+15 强实证(评估词>=3)")
	}

	if containsAny(paper.Abstract, []string{"ablation", "baseline"}) {
		score += 10
		details = append(details, "+10 消融/基线实验")
	}

	if containsAny(paper.Abstract, []string{"dataset", "benchmark"}) {
		score += 10
		details = append(details, "+10 数据集/基准测试")
	}

	if hasCodeLink(paper) {
		score += 10
		details = append(details, "+10 代码链接")
	}

	if containsAny(paper.Abstract, limitationKeywords) {
		score += 5
		details = append(details, "+5 局限性讨论")
	}

	if paper.Version() >= 2 {
		score += 5
		details = append(details, "+5 多版本迭代")
	}

	// Negative signals
	if containsAny(paper.Abstract, hypeKeywords) || containsAny(paper.Title, hypeKeywords) {
		score -= 10
		details = append(details, "-10 夸大营销词")
	}

	if containsAny(paper.Abstract, frameworkKeywords) && evalCount == 0 {
		score -= 25
		details = append(details, "-25 纯框架无评估")
	}

	// Ensure score is in valid range
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	result.Score = score
	result.Details = details

	return result
}

func countKeywords(text string, keywords []string) int {
	text = strings.ToLower(text)
	count := 0
	for _, kw := range keywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			count++
		}
	}
	return count
}

func containsAny(text string, keywords []string) bool {
	text = strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func hasCodeLink(paper model.Paper) bool {
	// Check in abstract
	if codeRepoPattern.MatchString(paper.Abstract) {
		return true
	}
	// Check in comments
	if codeRepoPattern.MatchString(paper.Comments) {
		return true
	}
	// Check in links
	for _, link := range paper.Links {
		if link.Type == "code" {
			return true
		}
		if codeRepoPattern.MatchString(link.URL) {
			return true
		}
	}
	return false
}
