package workspace_history

import (
	"strconv"
	"strings"

	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

const patchHeader = "diff --git "

func parseNameStatus(out string) []*input_itf.FileChange {
	fields := splitRecords(out)
	changes := make([]*input_itf.FileChange, 0, len(fields)/2)

	for i := 0; i < len(fields); {
		status := fields[i]
		i++
		if status == "" {
			continue
		}

		change := &input_itf.FileChange{ChangeType: changeType(status)}
		switch status[0] {
		case 'R', 'C':
			if i+1 >= len(fields) {
				return changes
			}
			change.OldPath = fields[i]
			change.Path = fields[i+1]
			i += 2
		default:
			if i >= len(fields) {
				return changes
			}
			change.Path = fields[i]
			i++
		}

		changes = append(changes, change)
	}

	return changes
}

func applyNumStat(changes []*input_itf.FileChange, out string) {
	fields := splitRecords(out)
	index := 0

	for i := 0; i < len(fields) && index < len(changes); {
		parts := strings.SplitN(fields[i], "\t", 3)
		i++
		if len(parts) != 3 {
			continue
		}

		if parts[2] == "" {
			i += 2
		}

		changes[index].Additions = count(parts[0])
		changes[index].Deletions = count(parts[1])
		index++
	}
}

func applyPatches(changes []*input_itf.FileChange, patch string) {
	patches := splitPatches(patch)
	for i := range changes {
		if i >= len(patches) {
			return
		}
		changes[i].UnifiedDiff = patches[i]
	}
}

func splitPatches(patch string) []string {
	start := strings.Index(patch, patchHeader)
	if start < 0 {
		return nil
	}

	body := patch[start:]
	parts := strings.Split(body, "\n"+patchHeader)
	patches := make([]string, 0, len(parts))

	for i, part := range parts {
		if i > 0 {
			part = patchHeader + part
		}
		patches = append(patches, strings.TrimRight(part, "\n"))
	}

	return patches
}

func splitRecords(out string) []string {
	out = strings.TrimLeft(out, "\n")
	if out == "" {
		return nil
	}

	fields := strings.Split(out, "\x00")
	for len(fields) > 0 && fields[len(fields)-1] == "" {
		fields = fields[:len(fields)-1]
	}

	return fields
}

func count(raw string) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}

	return value
}

func changeType(status string) enums.FileChangeType {
	switch status[0] {
	case 'A':
		return enums.FileAdded
	case 'D':
		return enums.FileDeleted
	case 'R':
		return enums.FileRenamed
	case 'C':
		return enums.FileAdded
	default:
		return enums.FileModified
	}
}
