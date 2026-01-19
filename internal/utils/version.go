package utils

import (
	"fmt"
	"strings"
)

func GetAppMetaInfo(version, date, commit string) string {
	sb := strings.Builder{}
	if version == "" {
		version = "N/A"
	}
	if date == "" {
		date = "N/A"
	}
	if commit == "" {
		commit = "N/A"
	}
	sb.WriteString(fmt.Sprintf("\nBuild version: %s\n", version))
	sb.WriteString(fmt.Sprintf("Build date: %s\n", date))
	sb.WriteString(fmt.Sprintf("Build commit: %s\n", commit))

	return sb.String()
}
