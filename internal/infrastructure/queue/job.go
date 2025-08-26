package queue

import "mime/multipart"

type JobType string

const (
	JobSaveChunk JobType = "save_chunk"
	JobMerge     JobType = "merge_chunks"
	JobCleanup   JobType = "cleanup"
)

type Job struct {
	UploadID    string
	Type        JobType
	Filename    string
	ChunkIndex  int
	File        multipart.File
	TotalChunks int
}
