package pageToolsUI

import "strings"

func ValidateURL(url string) bool {
	if !strings.Contains(url, "https://") || !strings.Contains(url, "http://") {
		return false
	}
	return true
}
