package model

import "time"

// Paper represents a scientific paper from ArXiv.
type Paper struct {
	ID         string    // ArXiv unique identifier
	Title      string    // Paper title
	Abstract   string    // Full abstract text
	Authors    []string  // List of author names
	Categories []string  // Academic category tags (e.g., cs.AI, cond-mat)
	UpdatedAt  time.Time // Last update timestamp
}
