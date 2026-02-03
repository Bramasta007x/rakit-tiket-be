package util

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
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

func CleanUsername(username string) string {
	// Trim leading and trailing white spaces
	username = strings.TrimSpace(username)

	// Define a regular expression to remove non-alphanumeric characters, except spaces
	reg := regexp.MustCompile(`[^\p{L}\p{N}\s]`)
	username = reg.ReplaceAllString(username, "")

	// Remove all emoticons
	username = RemoveEmoticons(username)

	// Trim spaces again in case there are multiple spaces left
	username = strings.TrimSpace(username)

	return username
}

func RemoveEmoticons(input string) string {
	var result []rune
	for _, r := range input {
		if !IsEmoticon(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

func IsEmoticon(r rune) bool {
	// Check if the character is an emoticon using unicode properties
	return unicode.Is(unicode.So, r) || unicode.Is(unicode.Sk, r)
}

func ShortenString(text string, length int) string {
	if len(text) <= length {
		return text
	}
	// Find the last space within the allowed length
	shortened := text[:length]
	lastSpace := strings.LastIndex(shortened, " ")
	if lastSpace == -1 {
		// If no space found, return the original slice (could cut a word)
		return shortened
	}
	// Otherwise, return up to the last space
	return shortened[:lastSpace] + "..."
}

func AtoF32(str string) float32 {
	result, _ := strconv.ParseFloat(str, 32)
	return float32(result)
}

func AtoF64(str string) float64 {
	result, _ := strconv.ParseFloat(str, 64)
	return result
}

func BtoA(val bool) string {
	return strconv.FormatBool(val)
}

func AtoB(val string) bool {
	valB, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return valB
}

func AtoI(val string) int {
	valB, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return valB
}
