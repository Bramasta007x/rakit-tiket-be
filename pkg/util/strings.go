package util

import (
	"regexp"
	"strings"
)

// English and Bahasa Indonesia stop words
var stopWords = map[string]map[string]bool{
	"en": {
		"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
		"is": true, "are": true, "was": true, "were": true, "be": true, "been": true,
		"in": true, "on": true, "at": true, "with": true, "of": true, "for": true,
		"to": true, "from": true, "by": true, "as": true, "that": true, "this": true,
		"it": true, "he": true, "she": true, "they": true, "you": true, "we": true,
		"i": true, "my": true, "your": true, "our": true, "their": true, "his": true,
		"her": true, "its": true, "what": true, "which": true, "who": true, "whom": true,
		"when": true, "where": true, "why": true, "how": true, "so": true, "than": true,
	},
	"id": {
		"dan": true, "atau": true, "tetapi": true, "adalah": true, "ialah": true,
		"merupakan": true, "yang": true, "untuk": true, "pada": true, "dari": true,
		"di": true, "ke": true, "sebagai": true, "dengan": true, "bagi": true,
		"oleh": true, "karena": true, "bahwa": true, "dalam": true, "itu": true,
		"ini": true, "tersebut": true, "tidak": true, "iya": true, "apa": true,
		"siapa": true, "kapan": true, "dimana": true, "mengapa": true, "bagaimana": true,
		"akan": true, "anda": true, "kami": true, "nya": true, "satu-satu": true,
	},
}

func GetAllStopWords() map[string]bool {
	newStopWords := make(map[string]bool)
	for _, stopWordByLang := range stopWords {
		for stopWord, ok := range stopWordByLang {
			newStopWords[stopWord] = ok
		}
	}
	return newStopWords
}

func CleanSentence(sentence string) string {
	words := strings.Fields(sentence)
	importantWords := []string{}

	stopWords := GetAllStopWords()
	for _, word := range words {
		// Convert word to lowercase and check if it is a stop word
		if _, found := stopWords[strings.ToLower(word)]; !found {
			importantWords = append(importantWords, word)
		}
	}
	return strings.Join(importantWords, " ")
}

func CleanSpecialChars(input string) string {
	// Define a regular expression pattern to match non-alphanumeric characters
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

	// Replace all non-alphanumeric characters with an empty string
	cleaned := re.ReplaceAllString(input, "")

	// Optionally, you can trim spaces if you want to remove leading/trailing spaces
	// cleaned = strings.TrimSpace(cleaned)

	return cleaned
}
