package queue

type JobType string

const (
	JobSaveChunk JobType = "save_chunk"
	JobMerge     JobType = "merge_chunks"
	JobCleanup   JobType = "cleanup"
	JobRetry     JobType = "retry_merge"
)

type Job struct {
	UploadID   string
	Type       JobType
	Filename   string
	ChunkIndex int
	//File        multipart.File
	FilePath    string `json:"file_path,omitempty"` // chunk dosya yolu
	TotalChunks int
	ChunkHash   string `json:"chunk_hash,omitempty"`
	FileContent []byte `json:"file_content,omitempty"`
	ErrorType   string `json:"error_type,omitempty"` // Hata türü (örneğin, "missing_chunk")
	LastError   string `json:"last_error,omitempty"`
	Status      string `json:"status,omitempty"`
	RetryCount  int    `json:"retry_count,omitempty"`
	//OnMergeSuccess func(uploadID, filename, mergedFilePath string)
}

type ProcessedJob struct {
	UploadID       string `json:"upload_id"`
	Filename       string `json:"filename"`
	MergedFilePath string `json:"merged_file_path"`
	TotalChunks    int    `json:"total_chunks"`
}
