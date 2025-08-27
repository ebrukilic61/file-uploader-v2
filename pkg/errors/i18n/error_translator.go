package i18n

import (
	"embed" //derleme zamanında dosyayı gömülü olarak ekler, çalışma zamanında aramasına gerek kalmaz
	"encoding/json"
	"fmt"
)

//go:embed *.json
var i18nFiles embed.FS

var messages map[string]string

func Load(locale string) error {
	filename := locale + ".json"

	data, err := i18nFiles.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read embedded i18n file %s: %w", filename, err)
	}

	messages = make(map[string]string)
	return json.Unmarshal(data, &messages)
}

func T(code string) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return code // fallback
}
