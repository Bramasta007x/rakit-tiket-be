package util

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type (
	results    map[int]string
	resultsMap map[string]results
)

func (m resultsMap) GetByName(inputName string) results {
	if results, ok := m[inputName]; ok {
		return results
	}
	return nil
}

func (m resultsMap) LenByName(inputName string) int {
	if results, ok := m[inputName]; ok {
		return len(results)
	}
	return 0
}

func (m resultsMap) Len() int {
	if len(m) > 0 {
		for _, results := range m {
			return len(results)
		}
	}
	return 0
}

func (m results) Convert() []string {
	var newResults []string
	for _, result := range m {
		newResults = append(newResults, result)
	}
	return newResults
}

func (m results) GetByIdx(idx int) string {
	if result, ok := m[idx]; ok {
		return result
	}
	return ""
}

func GetFormRepeaterInput(formParams url.Values, repeaterName string, inputNames ...string) resultsMap {
	var resultMap resultsMap
	for key, values := range formParams {
		for _, inputName := range inputNames {
			if strings.Contains(key, repeaterName) {
				// Regular expression to capture "inputName[1][external_link_name]"
				re := regexp.MustCompile(fmt.Sprintf(`^%s\[(\d+)\]\[(\w+)\]$`, repeaterName))
				matches := re.FindStringSubmatch(key)

				// Convert index part (matches[1]) to an integer
				index, err := strconv.Atoi(matches[1])
				if err != nil {
					return nil
				}

				// Extract field name part
				field := matches[2]
				if !strings.EqualFold(field, inputName) {
					continue
				}

				// Initialize inner map if it doesn't exist
				if resultMap == nil {
					resultMap = make(resultsMap)
				}
				if _, exists := resultMap[inputName]; !exists {
					resultMap[inputName] = make(map[int]string)
				}

				// Set the value in the nested map
				for _, value := range values {
					resultMap[inputName][index] = value
				}
			}
		}
	}
	return resultMap
}
