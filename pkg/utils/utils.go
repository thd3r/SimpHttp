package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func IsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func VerbosePrint(verbose bool, text string) {
	if verbose {
		fmt.Fprintf(os.Stderr, "%s", text)
	}
}

func ReadLines(reader io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(reader)
	filtered := make(map[string]bool)

	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line != "" && !filtered[line] {
			filtered[line] = true
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}

	return lines, nil
}
