package utils

import (
	"fmt"
	"strings"
)

func BuildVersion(version, gitRevision, buildTime string) string {
	builder := strings.Builder{}

	if len(version) == 0 {
		builder.WriteString(gitRevision)
	} else {
		builder.WriteString(fmt.Sprintf("%s (%s)", version, gitRevision))
	}

	builder.WriteString(fmt.Sprintf(" %s", buildTime))

	return builder.String()
}
