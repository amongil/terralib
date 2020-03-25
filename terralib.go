package terralib

import (
	"os/exec"
)

// Terralib struct holds the configuration for terralib
type Terralib struct {
	configPath string
}

// Init executes the 'terraform init' command
func (t *Terralib) Init() (string, error) {
	cmd := exec.Command("sh", "-c", "terraform init")
	stdoutStderr, err := cmd.CombinedOutput()
	return string(stdoutStderr), err
}
