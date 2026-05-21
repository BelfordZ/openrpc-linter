package functions

import "strings"

func resultPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return []string{path}
}

func fieldPath(basePath string, fieldName string) string {
	if basePath == "" {
		return ""
	}
	return basePath + pathNameSegment(fieldName)
}

func pathNameSegment(name string) string {
	escaped := strings.ReplaceAll(name, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return "['" + escaped + "']"
}
