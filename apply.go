package terralib

import (
	"os/exec"
	"regexp"
	"strings"
)

// Exported error codes
const (
	ErrApplyDefault string = "errApplyDefault"
)

// ApplyOutput represents the output of the apply command
type ApplyOutput struct {
	Raw string
}

var applyErrors = map[string]string{}

// ApplyError represents an error on the Init command
type ApplyError struct {
	Reason string
	Code   string
}

func (e ApplyError) Error() string {
	return e.Code
}

// Apply executes the 'terraform apply' command
func (t *Terralib) Apply(options []string) (ApplyOutput, error) {

	cmdString := formatCommand("apply", options)
	cmd := exec.Command("sh", "-c", cmdString)
	cmd.Dir = t.ConfigPath
	stdOutputError, _ := cmd.CombinedOutput()
	planError := findPlanError(stdOutputError)
	return ApplyOutput{
		Raw: string(stdOutputError),
	}, planError
}

func findApplyError(output []byte) error {
	var applyError ApplyError
	for k, v := range planErrors {
		r := regexp.MustCompile(v)
		line := r.Find(output)
		if line != nil {
			applyError = ApplyError{
				Reason: string(line),
				Code:   k,
			}
			return applyError
		}
	}
	// Get a default error or else return no error
	r := regexp.MustCompile("Error: (.*).")
	line := r.Find(output)
	if line != nil {
		reason := strings.Split(string(line), "Error: ")
		if reason != nil {
			return PlanError{
				Reason: strings.TrimSuffix(reason[1], "."),
				Code:   ErrApplyDefault,
			}
		}
	}
	return nil
}
