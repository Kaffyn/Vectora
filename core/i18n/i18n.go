package i18n

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"log"
	"sync"
)

//go:embed translations.csv
var translationCSV []byte

var currentLang = "pt"
var mu sync.RWMutex

// dictionary: [language][key] -> value
var dictionary = make(map[string]map[string]string)

func init() {
	r := csv.NewReader(bytes.NewReader(translationCSV))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("Falha crítica ao ler translations.csv: %v", err)
	}

	if len(records) < 2 {
		return
	}

	headers := records[0][1:]
	for _, lang := range headers {
		dictionary[lang] = make(map[string]string)
	}

	for _, row := range records[1:] {
		key := row[0]
		for i, lang := range headers {
			if len(row) > i+1 {
				dictionary[lang][key] = row[i+1]
			}
		}
	}
}

func SetLanguage(lang string) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := dictionary[lang]; ok {
		currentLang = lang
	}
}

func GetCurrentLang() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if val, ok := dictionary[currentLang][key]; ok && val != "" {
		return val
	}
	if val, ok := dictionary["en"][key]; ok && val != "" {
		return val
	}
	return key
}
