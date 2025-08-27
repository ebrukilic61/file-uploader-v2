package main

import (
	"fmt"
	"log"

	"file-uploader/pkg/errors"
	"file-uploader/pkg/errors/i18n"
)

func main() {
	// 1️⃣ i18n dosyasını yükle (örn: "tr")
	if err := i18n.Load("tr"); err != nil {
		log.Fatalf("i18n load failed: %v", err)
	}

	// 2️⃣ Hata oluştur
	uploadErr := errors.ErrNotFound(nil) // burada nil yerine low-level error de verilebilir

	// 3️⃣ Mesajı i18n üzerinden al
	localizedMessage := i18n.T(uploadErr.Code)
	fmt.Printf("UploadError Code: %s\n", uploadErr.Code)
	fmt.Printf("UploadError Message: %s\n", uploadErr.Message)
	fmt.Printf("i18n Message: %s\n", localizedMessage)

	// 4️⃣ Test: handler yerine direkt çıktıyı simüle et
	status := 404 // not_found için
	response := map[string]string{
		"error":   uploadErr.Code,
		"message": localizedMessage,
	}
	fmt.Printf("HTTP %d Response: %+v\n", status, response)
}
