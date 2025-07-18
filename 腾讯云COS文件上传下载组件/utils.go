package cos

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func validatePath(path string) error {
	if len(path) == 0 {
		return errors.New("path cannot be empty")
	}

	invalidChars := `<>:"|?*\`
	for _, char := range path {
		if strings.ContainsRune(invalidChars, char) {
			return fmt.Errorf("path contains invalid character: '%c'", char)
		}
	}
	return nil
}

func generateTempFileName(prefix string) string {
	timestamp := time.Now().Format("20060102_150405")
	randomNum := rand.Intn(1000000)
	return fmt.Sprintf("%s_%s_%06d", prefix, timestamp, randomNum)
}
