package main

import (
	"fmt"
	"log"
	"os"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
)

func main() {
	// Connect to the local LXD daemon
	c, err := lxd.ConnectLXDUnix("", nil)
	if err != nil {
		log.Fatalf("Failed to connect to LXD: %v", err)
	}

	// Define instance creation request
	req := api.InstancesPost{
		Name: "my-sample-container",
		Source: api.InstanceSource{
			Type:  "image",
			Alias: "ubuntu-22.04-ppc64le", // Specify the image alias
		},
		Type: "container",
	}

	// Create the instance
	op, err := c.CreateInstance(req)
	if err != nil {
		log.Fatalf("Failed to create instance: %v", err)
	}

	// Wait for the operation to complete
	err = op.Wait()
	if err != nil {
		log.Fatalf("Failed to wait for instance creation: %v", err)
	}
	fmt.Println("Instance created successfully.")

	// Start the instance
	reqState := api.InstanceStatePut{
		Action:  "start",
		Timeout: -1, // No timeout
	}
	op, err = c.UpdateInstanceState(req.Name, reqState, "")
	if err != nil {
		log.Fatalf("Failed to start instance: %v", err)
	}

	// Wait for the start operation to complete
	err = op.Wait()
	if err != nil {
		log.Fatalf("Failed to wait for instance start: %v", err)
	}
	fmt.Println("Instance started successfully.")

	// Define the command to be executed inside the container
	execReq := api.InstanceExecPost{
		Command:     []string{"echo", "Hello from inside the container"},
		WaitForWS:   true,
		Interactive: false,
	}

	// Set up exec arguments to capture stdout and stderr
	execArgs := lxd.InstanceExecArgs{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	// Execute the command inside the instance
	execOp, err := c.ExecInstance(req.Name, execReq, &execArgs)
	if err != nil {
		log.Fatalf("Failed to execute command in container: %v", err)
	}

	// Wait for the command execution to complete
	err = execOp.Wait()
	if err != nil {
		log.Fatalf("Failed to wait for command execution: %v", err)
	}
	fmt.Println("Command executed successfully.")
}
