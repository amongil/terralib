package terralib

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const initOutputSuccessWithProviders string = `
Initializing the backend...

Initializing provider plugins...
- Checking for available provider plugins...
- Downloading plugin for provider "azurerm" (hashicorp/azurerm) 2.2.0...
- Downloading plugin for provider "random" (hashicorp/random) 2.2.1...

The following providers do not have any version constraints in configuration,
so the latest version was installed.

To prevent automatic upgrades to new major versions that may contain breaking
changes, it is recommended to add version = "..." constraints to the
corresponding provider blocks in configuration, with the constraint strings
suggested below.

* provider.random: version = "~> 2.2"

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.`

const errProviderNotFoundTest string = `
Initializing the backend...

Initializing provider plugins...
- Checking for available provider plugins...
- Downloading plugin for provider "random" (hashicorp/random) 2.2.1...

Provider "azurm" not available for installation.

A provider named "azurm" could not be found in the Terraform Registry.

This may result from mistyping the provider name, or the given provider may
be a third-party provider that cannot be installed automatically.

In the latter case, the plugin must be installed manually by locating and
downloading a suitable distribution package and placing the plugin's executable
file in the following directory:
    terraform.d/plugins/darwin_amd64

Terraform detects necessary plugins by inspecting the configuration and state.
To view the provider versions requested by each module, run
"terraform providers".

- Downloading plugin for provider "azurerm" (hashicorp/azurerm) 2.3.0...

Error: no provider exists with the given name
`

const errDiscoveryServiceUnreachableTest string = `
Initializing the backend...

Initializing provider plugins...
- Checking for available provider plugins...

Registry service unreachable.

This may indicate a network issue, or an issue with the requested Terraform Registry.

Error: registry service is unreachable, check https://status.hashicorp.com/ for status updates
`

const errProviderVersionsUnsuitableTest string = `
Initializing the backend...

Initializing provider plugins...
- Checking for available provider plugins...

No provider "azurerm" plugins meet the constraint "=2.5.0".

The version constraint is derived from the "version" argument within the
provider "azurerm" block in configuration. Child modules may also apply
provider version constraints. To view the provider versions requested by each
module in the current configuration, run "terraform providers".

To proceed, the version constraints for this provider must be relaxed by
either adjusting or removing the "version" argument in the provider blocks
throughout the configuration.


Error: no suitable version is available
`

const errProviderIncompatibleTest string = `
Initializing the backend...

Initializing provider plugins...
- Checking for available provider plugins...

Provider "azurerm" v0.1.0 is not compatible with Terraform 0.12.24.

Provider version 1.27.0 is the earliest compatible version. Select it with
the following version constraint:

	version = "~> 1.27"

Terraform checked all of the plugin versions matching the given constraint:
    =0.1

Consult the documentation for this provider for more information on
compatibility between provider and Terraform versions.


Error: incompatible provider version
`

const errProviderInstallErrorTest string = `
Initializing the backend...

Initializing provider plugins...

Checking for available provider plugins...
Downloading plugin for provider "AWS" (hashicorp/aws) 2.31.0...
Error installing provider "AWS": failed to find installed plugin version 2.31.0; this is a bug in Terraform and should be reported.

Terraform analyses the configuration and state and automatically downloads
plugins for the providers used. However, when attempting to download this
plugin an unexpected error occurred.

This may be caused if for some reason Terraform is unable to reach the
plugin repository. The repository may be unreachable if access is blocked
by a firewall.

If automatic installation is not possible or desirable in your environment,
you may alternatively manually install plugins by downloading a suitable
distribution package and placing the plugin's executable file in the
following directory:
terraform.d/plugins/linux_amd64
`

const errMissingProvidersNoInstallTest string = `
Initializing provider plugins...

Missing required providers.

The following provider constraints are not met by the currently-installed
provider plugins:

* rancher2 (any version)

Terraform can automatically download and install plugins to meet the given
constraints, but this step was skipped due to the use of -get-plugins=false
and/or -plugin-dir on the command line.

If automatic installation is not possible or desirable in your environment,
you may manually install plugins by downloading a suitable distribution package
and placing the plugin's executable file in one of the directories given in
by -plugin-dir on the command line, or in the following directory if custom
plugin directories are not set:
    terraform.d/plugins/linux_amd64
`

const errChecksumVerificationTest string = `
Error verifying checksum for provider "AWS"
The checksum for provider distribution from the Terraform Registry
did not match the source. This may mean that the distributed files
were changed after this version was released to the Registry.
`

const errSignatureVerificationTest string = `
Error verifying GPG signature for provider "AWS"
Terraform was unable to verify the GPG signature of the downloaded provider
files using the keys downloaded from the Terraform Registry. This may mean that
the publisher of the provider removed the key it was signed with, or that the
distributed files were changed after this version was released.
`
const errInitCopyNotEmptyTest string = `
The working directory already contains files. The -from-module option requires
an empty directory into which a copy of the referenced module will be placed.
To initialize the configuration already in this working directory, omit the
-from-module option.
`

func TestGetProvidersFromOutput(t *testing.T) {
	expected := []Provider{
		{
			Name:    "azurerm",
			Path:    "hashicorp/azurerm",
			Version: "2.2.0",
		},
		{
			Name:    "random",
			Path:    "hashicorp/random",
			Version: "2.2.1",
		},
	}
	got := getProvidersFromOutput([]byte(initOutputSuccessWithProviders))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %v, Expected: %v", got, expected)
	}
}

func TestFormatCommand(t *testing.T) {
	expected := "terraform init -verify-plugins=true -no-color"
	options := []string{"-verify-plugins=true", "-no-color"}
	got := formatCommand("init", options)
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %v, Expected: %v", got, expected)
	}
}

func TestFindErrProviderNotFound(t *testing.T) {
	expected := InitError{
		Reason: "Provider \"azurm\" not available for installation",
		Code:   "errProviderNotFound",
	}
	got := findInitError([]byte(errProviderNotFoundTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrDiscoveryServiceUnreachable(t *testing.T) {
	expected := InitError{
		Reason: "Registry service unreachable",
		Code:   "errDiscoveryServiceUnreachable",
	}
	got := findInitError([]byte(errDiscoveryServiceUnreachableTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrProviderVersionsUnsuitable(t *testing.T) {
	expected := InitError{
		Reason: "No provider \"azurerm\" plugins meet the constraint \"=2.5.0\"",
		Code:   "errProviderVersionsUnsuitable",
	}
	got := findInitError([]byte(errProviderVersionsUnsuitableTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrProviderIncompatible(t *testing.T) {
	tfVersion, err := getTerraformVersion()
	if err != nil {
		t.Errorf("Could not get terraform version on this machine: " + err.Error())
	}
	expected := InitError{
		Reason: fmt.Sprintf("Provider \"azurerm\" v0.1.0 is not compatible with Terraform %s", tfVersion),
		Code:   "errProviderIncompatible",
	}
	got := findInitError([]byte(errProviderIncompatibleTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrProviderInstallError(t *testing.T) {
	expected := InitError{
		Reason: ("Error installing provider \"AWS\": failed to find installed plugin version 2.31.0; " +
			"this is a bug in Terraform and should be reported"),
		Code: "errProviderInstallError",
	}
	got := findInitError([]byte(errProviderInstallErrorTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrMissingProvidersNoInstallTest(t *testing.T) {
	expected := InitError{
		Reason: ("The following provider constraints are not met by the currently-installed\n" +
			"provider plugins:\n\n" +
			"* rancher2 (any version)"),
		Code: "errMissingProvidersNoInstall",
	}
	got := findInitError([]byte(errMissingProvidersNoInstallTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrChecksumVerificationTest(t *testing.T) {
	expected := InitError{
		Reason: "Error verifying checksum for provider \"AWS\"",
		Code:   "errChecksumVerification",
	}
	got := findInitError([]byte(errChecksumVerificationTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrSignatureVerificationTest(t *testing.T) {
	expected := InitError{
		Reason: "Error verifying GPG signature for provider \"AWS\"",
		Code:   "errSignatureVerification",
	}
	got := findInitError([]byte(errSignatureVerificationTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrInitCopyNotEmptyTest(t *testing.T) {
	expected := InitError{
		Reason: "The working directory already contains files",
		Code:   "errInitCopyNotEmpty",
	}
	got := findInitError([]byte(errInitCopyNotEmptyTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func getTerraformVersion() (string, error) {
	cmd := exec.Command("sh", "-c", "terraform version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var version string
	_, err = fmt.Sscanf(string(output), "Terraform v%s", &version)
	if err != nil {
		return "", err
	}
	return version, err
}
