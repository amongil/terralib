package terralib

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const planOutputInvalidResourceTypeTest string = `
Error: Invalid resource type

  on main.tf line 6, in resource "azurerm_non_existant" "example":
   6: resource "azurerm_non_existant" "example" {

The provider provider.azurerm does not support resource type
"azurerm_non_existant".
`

const planOutputErrDefaultTest string = `
Error: something wrong happened.
`

const planOutputErrCouldNotSatisfyPluginRequirementsTest string = `
Error: Could not satisfy plugin requirements


Plugin reinitialization required. Please run "terraform init".

Plugins are external binaries that Terraform uses to access and manipulate
resources. The configuration provided requires plugins which can't be located,
don't satisfy the version constraints, or are otherwise incompatible.

Terraform automatically discovers provider requirements from your
configuration, including providers used in child modules. To see the
requirements and constraints from each module, run "terraform providers".



Error: provider.non: no suitable version installed
  version requirements: "(any version)"
  versions installed: none
`

func TestFindErrInvalidResourceType(t *testing.T) {
	expected := PlanError{
		Reason: ("The provider provider.azurerm does not support resource type\n" +
			"\"azurerm_non_existant\"."),
		Code: "errInvalidResourceType",
	}
	got := findPlanError([]byte(planOutputInvalidResourceTypeTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrDefaultType(t *testing.T) {
	expected := PlanError{
		Reason: "something wrong happened",
		Code:   "errDefault",
	}
	got := findPlanError([]byte(planOutputErrDefaultTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}

func TestFindErrCouldNotSatisfyPluginRequirements(t *testing.T) {
	expected := PlanError{
		Reason: ("provider.non: no suitable version installed\n" +
			"  version requirements: \"(any version)\"\n" +
			"  versions installed: none"),
		Code: "errCouldNotSatisfyPluginRequirements",
	}
	got := findPlanError([]byte(planOutputErrCouldNotSatisfyPluginRequirementsTest))
	if !cmp.Equal(got, expected) {
		t.Errorf("Got: %+v, Expected: %+v", got, expected)
	}
}
