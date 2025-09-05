package queue

type JobType string

const (
	JobSaveChunk JobType = "save_chunk"
	JobMerge     JobType = "merge_chunks"
	JobCleanup   JobType = "cleanup"
)

type Job struct {
	UploadID   string
	Type       JobType
	Filename   string
	ChunkIndex int
	//File        multipart.File
	FilePath    string `json:"file_path,omitempty"` // chunk dosya yolu
	TotalChunks int

	//OnMergeSuccess func(uploadID, filename, mergedFilePath string)
}

type ProcessedJob struct {
	UploadID       string `json:"upload_id"`
	Filename       string `json:"filename"`
	MergedFilePath string `json:"merged_file_path"`
	TotalChunks    int    `json:"total_chunks"`
}
