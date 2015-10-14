package formatter

import (
	"bytes"
	"fmt"

	"github.com/octavore/delta"
)

func ColoredText(d *delta.DiffSolution) string {
	buf := &bytes.Buffer{}
	for _, l := range d.Lines {
		if l[2] == "=" && l[0] == l[1] {
			fmt.Fprintf(buf, " %s \n", l[0])
			continue
		}
		if l[0] != "" {
			fmt.Fprintf(buf, "\x1b[31m-%s\x1b[0m\n", l[0])
		}
		if l[1] != "" {
			fmt.Fprintf(buf, "\x1b[32m+%s\x1b[0m\n", l[1])
		}
	}
	return buf.String()
}
