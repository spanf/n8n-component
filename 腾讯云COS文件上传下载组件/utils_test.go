package cos

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{"Empty path", "", errors.New("path cannot be empty")},
		{"Contains <", "a<test", fmt.Errorf("path contains invalid character: '%c'", '<')},
		{"Contains >", "a>test", fmt.Errorf("path contains invalid character: '%c'", '>')},
		{"Contains :", "a:test", fmt.Errorf("path contains invalid character: '%c'", ':')},
		{"Contains \"", "a\"test", fmt.Errorf("path contains invalid character: '%c'", '"')},
		{"Contains |", "a|test", fmt.Errorf("path contains invalid character: '%c'", '|')},
		{"Contains ?", "a?test", fmt.Errorf("path contains invalid character: '%c'", '?')},
		{"Contains *", "a*test", fmt.Errorf("path contains invalid character: '%c'", '*')},
		{"Contains \\", "a\\test", fmt.Errorf("path contains invalid character: '%c'", '\\')},
		{"Valid path", "valid/path/123.txt", nil},
		{"Unicode path", "中文路径/こんにちは.txt", nil},
		{"Space path", "path with spaces", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil {
				if err.Error() != tt.wantErr.Error() {
					t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestGenerateTempFileName(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"Basic prefix", "test"},
		{"Empty prefix", ""},
		{"Special chars prefix", "préfix-123"},
		{"Long prefix", strings.Repeat("a", 50)},
	}

	pattern := `^[a-zA-Z0-9_-]{1,50}_\d{8}_\d{6}_\d{6}$`
	regex := regexp.MustCompile(pattern)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateTempFileName(tt.prefix)

			if !regex.MatchString(got) {
				t.Errorf("generateTempFileName() = %v, doesn't match pattern %s", got, pattern)
			}

			if tt.prefix != "" && !strings.HasPrefix(got, tt.prefix+"_") {
				t.Errorf("generateTempFileName() = %v, missing prefix %s", got, tt.prefix)
			}

			parts := strings.Split(got, "_")
			if len(parts) < 3 {
				t.Errorf("generateTempFileName() = %v, invalid format", got)
			}
		})
	}

	t.Run("Uniqueness check", func(t *testing.T) {
		results := make(map[string]bool, 1000)
		for i := 0; i < 1000; i++ {
			name := generateTempFileName("test")
			if results[name] {
				t.Errorf("Duplicate temp file name generated: %s", name)
			}
			results[name] = true
		}
	})
}
