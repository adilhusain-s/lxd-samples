package main

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"

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

	// Define the command to get the Docker group ID inside the container
	execReq := api.InstanceExecPost{
		Command:     []string{"getent", "group", "docker"},
		WaitForWS:   true,
		Interactive: false,
	}

	// Set up buffers to capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	execArgs := lxd.InstanceExecArgs{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
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

	// Parse the output to extract the Docker group ID
	groupID, err := parseDockerGroupID(stdoutBuf.String())
	if err != nil {
		log.Fatalf("Failed to parse Docker group ID: %v", err)
	}

	fmt.Printf("Docker group ID: %d\n", groupID)
}

// parseDockerGroupID parses the group ID from the `getent group docker` output
func parseDockerGroupID(output string) (uint32, error) {
	// Define the regular expression to match the group ID
	re := regexp.MustCompile(`^docker:x:(\d+):`)
	matches := re.FindStringSubmatch(output)

	// Check if the group ID was captured
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid output format from 'getent group docker': %s", output)
	}

	// Convert the captured group ID to uint32
	groupID, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to convert group ID to uint32: %v", err)
	}

	return uint32(groupID), nil
}
