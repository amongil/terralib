package terralib

import (
	"fmt"
	"strings"
)

func formatCommand(cmd string, options []string) string {
	return fmt.Sprintf("terraform %s %s", cmd, strings.Join(options, " "))
}
