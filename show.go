package terralib

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"
)

// Exported error codes
const (
	ErrShowDefault string = "errShowDefault"
)

var showErrors = map[string]string{}

// ShowError represents an error on the Init command
type ShowError struct {
	Reason string
	Code   string
}

func (e ShowError) Error() string {
	return e.Code
}

// ShowOutput represents the output of the show command
type ShowOutput struct {
	FormatVersion    string      `json:"format_version,omitempty"`
	TerraformVersion string      `json:"terraform_version,omitempty"`
	PlannedValues    interface{} `json:"planned_values,omitempty"`
	ResourceChanges  interface{} `json:"resource_changes,omitempty"`
	Configuration    interface{} `json:"configuration,omitempty"`
}

// Show executes the 'terraform show' command
func (t *Terralib) Show(path string) (ShowOutput, error) {
	options := []string{
		"-no-color",
		"-json",
		path,
	}
	cmdString := formatCommand("show", options)
	cmd := exec.Command("sh", "-c", cmdString)
	cmd.Dir = t.ConfigPath
	stdOutputError, _ := cmd.CombinedOutput()
	showError := findShowError(stdOutputError)
	// Unmarshal data
	var output ShowOutput
	json.Unmarshal([]byte(stdOutputError), &output)
	return output, showError
}

func findShowError(output []byte) error {
	var showError ShowError
	for k, v := range showErrors {
		r := regexp.MustCompile(v)
		line := r.Find(output)
		if line != nil {
			showError = ShowError{
				Reason: string(line),
				Code:   k,
			}
			return showError
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
				Code:   ErrShowDefault,
			}
		}
	}
	return nil
}
