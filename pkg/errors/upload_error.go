package errors

import "fmt"

type UploadError struct {
	Code    string
	Message string
	Err     error
}

func (e *UploadError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrNotFound = func(err error) *UploadError {
		return &UploadError{Code: "not_found", Message: "Dosya bulunamadı", Err: err}
	}
	ErrInternal = func(err error) *UploadError {
		return &UploadError{Code: "internal_error", Message: "Sunucu hatası", Err: err}
	}
	ErrChunkNotOpen = func(err error) *UploadError {
		return &UploadError{Code: "chunk_not_open", Message: "Chunk açılamadı", Err: err}
	}
	ErrChunkNotSave = func(err error) *UploadError {
		return &UploadError{Code: "chunk_not_save", Message: "Chunk kaydedilemedi", Err: err}
	}
	ErrTmpFile = func(err error) *UploadError {
		return &UploadError{Code: "tmp_file_error", Message: "Geçici dosya hatası", Err: err}
	}
	ErrFileCantOpen = func(err error) *UploadError {
		return &UploadError{Code: "file_cant_open", Message: "Dosya açılamadı", Err: err}
	}
	ErrInvalidChunk = func(err error) *UploadError {
		return &UploadError{Code: "invalid_chunk", Message: "Geçersiz chunk index", Err: err}
	}
	ErrCannotStat = func(err error) *UploadError {
		return &UploadError{Code: "cannot_stat", Message: "Stat alınamadı", Err: err}
	}
	ErrCannotRemove = func(err error) *UploadError {
		return &UploadError{Code: "cannot_remove", Message: "Dosya kaldırılamadı", Err: err}
	}
)
