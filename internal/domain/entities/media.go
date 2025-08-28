package entities

import "time"

type Media struct {
	MediaID      string
	Filename     string
	FileType     string
	FileSize     int64
	Status       string
	OriginalPath string
	Metadata     Metadata
	Jobs         []*MediaJob
}

type MediaJob struct {
	JobID      string
	MediaID    string
	Type       string
	Status     string
	Params     map[string]string
	OutputPath string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Metadata struct {
	Width    int
	Height   int
	Format   string
	Duration int
	Size     int64
}
