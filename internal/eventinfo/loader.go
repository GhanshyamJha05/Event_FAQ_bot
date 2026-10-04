package eventinfo

import (
	"os"
	"strings"
)

// Load reads the event file and returns its content as a string.
func Load(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
