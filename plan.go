package terralib

import (
	"os/exec"
	"regexp"
	"strings"
)

// Exported error codes
const (
	ErrInvalidResourceType               string = "errInvalidResourceType"
	ErrCouldNotSatisfyPluginRequirements string = "errCouldNotSatisfyPluginRequirements"
	ErrPlanDefault                       string = "errPlanDefault"
)

var planErrors = map[string]string{
	ErrInvalidResourceType: ("The provider (.*) does not support resource type\n" +
		"\"(.*)\"."),
	ErrCouldNotSatisfyPluginRequirements: ("provider.(.*): no suitable version installed\n" +
		"  version requirements: \"(.*)\"\n" +
		"  versions installed: (.*)"),
}

// PlanError represents an error on the Init command
type PlanError struct {
	Reason string
	Code   string
}

// PlanOutput represents the output of the plan command
type PlanOutput struct {
	Raw string
}

func (e PlanError) Error() string {
	return e.Code
}

// Plan executes the 'terraform plan' command
func (t *Terralib) Plan(options []string) (PlanOutput, error) {

	cmdString := formatCommand("plan", options)
	cmd := exec.Command("sh", "-c", cmdString)
	cmd.Dir = t.ConfigPath
	stdOutputError, _ := cmd.CombinedOutput()
	planError := findPlanError(stdOutputError)
	return PlanOutput{
		Raw: string(stdOutputError),
	}, planError
}

func findPlanError(output []byte) error {
	var planError PlanError
	for k, v := range planErrors {
		r := regexp.MustCompile(v)
		line := r.Find(output)
		if line != nil {
			planError = PlanError{
				Reason: string(line),
				Code:   k,
			}
			return planError
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
				Code:   ErrPlanDefault,
			}
		}
	}
	return nil
}
