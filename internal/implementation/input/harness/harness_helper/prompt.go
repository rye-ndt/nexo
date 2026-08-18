package harness_helper

import "strings"

func PromptText(base []byte, extra []string) string {
	parts := make([]string, 0, len(extra)+1)

	for _, prompt := range append([]string{string(base)}, extra...) {
		if trimmed := strings.TrimSpace(prompt); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}

	return strings.Join(parts, "\n\n")
}
