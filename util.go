package bambulabs_api

import (
	"regexp"
	"strings"
)

func isValidGCode(line string) bool {
	line = strings.Split(line, ";")[0]
	line = strings.TrimSpace(line)

	re := regexp.MustCompile(`^[GM]\d+`)
	if line == "" || !re.MatchString(line) {
		return false
	}

	tokens := strings.Fields(line)
	for _, token := range tokens[1:] {
		paramRe := regexp.MustCompile(`^[A-Z]-?\d+(\.\d+)?$`)
		if !paramRe.MatchString(token) {
			return false
		}
	}

	return true
}
