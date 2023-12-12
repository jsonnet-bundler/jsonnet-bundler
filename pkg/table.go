package pkg

import (
	"fmt"
	"io"
	"strings"
)

func PrintRow(writer io.Writer, line []string) {
	fmt.Fprintln(writer, strings.Join(line, "\t"))
}

func PrintHeader(writer io.Writer, headers []string) {
	PrintRow(writer, headers)
	PrintRow(writer, replaceRow(headers))
}

func replaceRow(row []string) []string {
	replaced := make([]string, len(row))
	for i, v := range row {
		replaced[i] = strings.Repeat("-", len(v))
	}
	return replaced
}
