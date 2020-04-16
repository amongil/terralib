# terralib
Terralib is a Go module that allows issuing terraform commands in a Go program by masking the underlying CLI runs.

## Notes
Even though it is a prototype version, It is possible to run a whole Terraform workflow using this library (init, plan, apply).

Features:
* Call Terraform CLI commands in Go programs
* Explicit error codes for each stage (init is quite complete, more work to be done on plan, apply and show)
* Save planned resource changes on Plan command output in a Go struct, so you can check for specific values or give the data another format to send to logging systems

## Example
```Go
package main

import (
	"fmt"
	"log"
	"os"

	terralib "github.com/amongil/terralib"
)

func main() {
    // Init terralib. ConfigPath is the path on where the Terraform config files are
	tf := terralib.Terralib{
		ConfigPath: "terraform-files",
    }
    
    // Terraform init
	log.Println("Running terraform init...")
	options := []string{}
	initOutput, err := tf.Init(options)
	if err != nil {
		defer fmt.Println(initOutput.Raw)
		log.Printf("Error on terraform init: %s\n", err)
		// We can act on specific errors and maybe remediate
		if err.Error() == terralib.ErrProviderNotFound {
			log.Printf("Full error: %s\n", err.(terralib.InitError).Reason)
		}
		return
    }
    
    // Terraform plan
	log.Println("Running terraform plan...")
	planOptions := []string{
		"-out=planfile",
	}
	_, err = tf.Plan(planOptions)
	if err != nil {
		log.Printf("Error on terraform plan: %s\n", err)
		if err.Error() == terralib.ErrInvalidResourceType {
			log.Printf("Full error: %s\n", err.(terralib.PlanError).Reason)
		}
		return
    }
    
    // Issue Terraform Show command on the plan file and save info on memory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	showOutput, err := tf.Show(dir + "/terraform-files/planfile")
	// Having our planned resource changes in a map allows us to make decisions over them
	// maybe even send them to event hubs for posterior analysis?
	plannedValues := (showOutput.PlannedValues).(map[string]interface{})
	for k, v := range plannedValues {
		fmt.Printf("%s: %v\n", k, v.(map[string]interface{})["resources"])
	}

    // Terraform apply
	log.Println("Running terraform apply...")
	applyOptions := []string{
		"-auto-approve",
	}
	applyOutput, err := tf.Apply(applyOptions)
	if err != nil {
		log.Printf("Error on terraform apply: %s\n", err)
		return
	}
	fmt.Println(applyOutput)

}
```