package utils

import (
	"fmt"
	"strings"
)

func ColoredText(c, text string) string {
	switch strings.ToLower(c) {
	case "red":
		return fmt.Sprintf("\033[31m%s\033[0m", text)
	case "gray":
		return fmt.Sprintf("\033[2m%s\033[0m", text)
	case "blue":
		return fmt.Sprintf("\033[34m%s\033[0m", text)
	case "magenta":
		return fmt.Sprintf("\033[35m%s\033[0m", text)
	case "cyan":
		return fmt.Sprintf("\033[36m%s\033[0m", text)
	case "green":
		return fmt.Sprintf("\033[32m%s\033[0m", text)
	case "yellow":
		return fmt.Sprintf("\033[33m%s\033[0m", text)
	default:
		return text
	}
}
