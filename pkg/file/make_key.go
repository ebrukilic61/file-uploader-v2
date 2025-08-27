package file

import "strings"

func MakeKey(uploadID, filename string) string {
	cleanUploadID := strings.TrimPrefix(uploadID, "upload-")
	return cleanUploadID + "_" + filename
}
