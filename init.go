package terralib

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Exported error codes
const (
	ErrInitCopyNotEmpty            string = "errInitCopyNotEmpty"
	ErrProviderNotFound            string = "errProviderNotFound"
	ErrDiscoveryServiceUnreachable string = "errDiscoveryServiceUnreachable"
	ErrProviderVersionsUnsuitable  string = "errProviderVersionsUnsuitable"
	ErrProviderIncompatible        string = "errProviderIncompatible"
	ErrProviderInstallError        string = "errProviderInstallError"
	ErrMissingProvidersNoInstall   string = "errMissingProvidersNoInstall"
	ErrChecksumVerification        string = "errChecksumVerification"
	ErrSignatureVerification       string = "errSignatureVerification"
)

var initErrors = map[string]error{
	ErrInitCopyNotEmpty:            errors.New("The working directory already contains files"),
	ErrProviderNotFound:            errors.New("Provider \"(.*)\" not available for installation"),
	ErrDiscoveryServiceUnreachable: errors.New("Registry service unreachable"),
	ErrProviderVersionsUnsuitable:  errors.New("No provider \"(.*)\" plugins meet the constraint \"(.*)\""),
	ErrProviderIncompatible:        errors.New("Provider \"(.*)\" (.*) is not compatible with Terraform (.*)"),
	ErrProviderInstallError:        errors.New("Error installing provider \"(.*)\": (.*)"),
	ErrMissingProvidersNoInstall: errors.New("The following provider constraints are not met by the currently-installed\n" +
		"provider plugins:\n\n" +
		"(.*)"),
	ErrChecksumVerification:  errors.New("Error verifying checksum for provider \"(.*)\""),
	ErrSignatureVerification: errors.New("Error verifying GPG signature for provider \"(.*)\""),
}

// Provider represents a Terraform provider
type Provider struct {
	Name    string
	Path    string
	Version string
}

// InitOutput represents the output of the init command
type InitOutput struct {
	Raw                  string
	InitializedProviders []Provider
}

// InitError represents an error on the Init command
type InitError struct {
	Reason string
	Code   string
}

func (e InitError) Error() string {
	return e.Code
}

// Init executes the 'terraform init' command
func (t *Terralib) Init(options []string) (InitOutput, error) {
	cmdString := formatCommand("init", options)
	cmd := exec.Command("sh", "-c", cmdString)
	cmd.Dir = t.ConfigPath
	stdOutputError, _ := cmd.CombinedOutput()
	initProviders := getProvidersFromOutput(stdOutputError)
	initError := findInitError(stdOutputError)
	return InitOutput{
		Raw:                  string(stdOutputError),
		InitializedProviders: initProviders,
	}, initError
}

func getProvidersFromOutput(out []byte) []Provider {
	var providers []Provider
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "- Downloading plugin for provider") {
			var name string
			var path string
			var version string
			fmt.Sscanf(line, "- Downloading plugin for provider %q %s %s", &name, &path, &version)
			provider := Provider{
				Name:    name,
				Path:    strings.Trim(path, "()"),
				Version: strings.TrimSuffix(version, "..."),
			}
			providers = append(providers, provider)
		}
	}
	return providers
}

func findInitError(output []byte) error {
	for k, v := range initErrors {
		r := regexp.MustCompile(v.Error())
		line := r.Find(output)
		if line != nil {
			return InitError{
				Reason: strings.TrimRight(string(line), "."),
				Code:   k,
			}
		}
	}
	return nil
}
