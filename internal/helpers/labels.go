package helpers

import "fmt"

func Labels[T fmt.Stringer](items []T) []string {
	names := make([]string, 0, len(items))

	for _, item := range items {
		names = append(names, item.String())
	}

	return names
}
