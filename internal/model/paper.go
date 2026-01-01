package model

import "time"

// Paper represents a scientific paper from ArXiv.
type Paper struct {
	ID         string    // ArXiv unique identifier (e.g., "2301.00001v1")
	Title      string    // Paper title
	Abstract   string    // Full abstract text
	Authors    []string  // List of author names
	Categories []string  // Academic category tags (e.g., cs.AI, cond-mat)
	UpdatedAt  time.Time // Last update timestamp

	// Extended fields for quality filtering
	Comments   string // Author comments (may contain "accepted", "to appear", etc.)
	DOI        string // Digital Object Identifier
	JournalRef string // Journal reference
	Links      []Link // Related links (PDF, code repos, etc.)

	// Computed fields (populated by filter)
	Score        int      // Quality score (0-100)
	ScoreDetails []string // Breakdown of score components
}

// Link represents a related link for a paper.
type Link struct {
	URL   string // Full URL
	Type  string // "abstract", "pdf", "code", "data", etc.
	Title string // Optional title/description
}

// Version extracts the version number from the paper ID.
// e.g., "2301.00001v2" -> 2, "2301.00001" -> 1
func (p Paper) Version() int {
	for i := len(p.ID) - 1; i >= 0; i-- {
		if p.ID[i] == 'v' {
			if i+1 < len(p.ID) {
				v := 0
				for j := i + 1; j < len(p.ID); j++ {
					if p.ID[j] >= '0' && p.ID[j] <= '9' {
						v = v*10 + int(p.ID[j]-'0')
					}
				}
				if v > 0 {
					return v
				}
			}
		}
	}
	return 1
}
