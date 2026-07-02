package config

import (
	"bufio"
	"os"
	"strings"
)

// loadDotenv reads a simple KEY=VALUE file and populates the environment for any
// keys not already set (real environment wins). Quotes around values are
// stripped. Lines starting with '#' are comments.
func loadDotenv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = strings.Trim(val, `"'`)
		if key == "" {
			continue
		}
		if _, set := os.LookupEnv(key); !set {
			_ = os.Setenv(key, val)
		}
	}
}
