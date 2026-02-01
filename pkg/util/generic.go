package util

import (
	"fmt"
	"regexp"
	"strings"
)

func ToSnakeCase(value string) string {
	return strings.ReplaceAll(strings.ToLower(value), " ", "-")
}

func ToCanonicalUri(values ...string) string {
	var result string
	for i, value := range values {
		if i > 0 {
			result = fmt.Sprintf("%s ", result)
		}
		result = fmt.Sprintf("%s%s", result, strings.ToLower(value))
	}
	// Define a regular expression to match all special characters, spaces, and white spaces
	regex := regexp.MustCompile("[^a-zA-Z0-9]+")
	// Replace all matches with a dash ("-")
	result = regex.ReplaceAllString(result, "-")
	// Define a regular expression to remove trailing dashes ("-")
	trailingDashRegex := regexp.MustCompile("-+$")
	// Remove trailing dashes ("-")
	return trailingDashRegex.ReplaceAllString(result, "")
}

func ValidateRegex(regexFormula, value string) bool {
	re := regexp.MustCompile(regexFormula)
	return re.MatchString(value)
}

func FindMatchRegex(regexFormula, value string) []string {
	re := regexp.MustCompile(regexFormula)
	return re.FindStringSubmatch(value)
}
