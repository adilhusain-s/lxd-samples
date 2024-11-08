package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "github.com/lxc/lxd/client"
    "github.com/lxc/lxd/shared/api"
)

func main() {
    // Connect to the LXD daemon on the host
    client, err := lxd.ConnectLXDUnix("", nil)
    if err != nil {
        log.Fatalf("Error connecting to LXD: %v", err)
    }

    containerName := "my-container" // Replace with your container name
    command := []string{"id", "ubuntu"} // Command to verify group memberships

    // Define the exec request with group mappings
    execRequest := api.InstanceExecPost{
        Command: command,
        Environment: map[string]string{
            "HOME": "/home/ubuntu",
            "USER": "ubuntu",
        },
        User: 1000,  // Set UID for 'ubuntu' user
        Group: 1000, // Set primary GID for 'ubuntu' user

        // Map container groups for the 'ubuntu' user
        GidMappings: []api.InstanceExecPostGidMapping{
            {ContainerID: 4, HostID: 4, Size: 1},       // adm
            {ContainerID: 20, HostID: 20, Size: 1},     // dialout
            {ContainerID: 24, HostID: 24, Size: 1},     // cdrom
            {ContainerID: 25, HostID: 25, Size: 1},     // floppy
            {ContainerID: 27, HostID: 27, Size: 1},     // sudo
            {ContainerID: 29, HostID: 29, Size: 1},     // audio
            {ContainerID: 30, HostID: 30, Size: 1},     // dip
            {ContainerID: 44, HostID: 44, Size: 1},     // video
            {ContainerID: 46, HostID: 46, Size: 1},     // plugdev
            {ContainerID: 119, HostID: 119, Size: 1},   // netdev
            {ContainerID: 120, HostID: 120, Size: 1},   // lxd
        },
        WaitForWS: true,  // Wait for the command to complete before returning
    }

    // Execute the command in the container
    op, err := client.ExecInstance(containerName, execRequest, &lxd.InstanceExecArgs{
        Stdin:    os.Stdin,
        Stdout:   os.Stdout,
        Stderr:   os.Stderr,
        DataDone: make(chan bool),
    })
    if err != nil {
        log.Fatalf("Error executing command in container: %v", err)
    }

    // Wait for the command to complete and get the result
    err = op.Wait(context.Background())
    if err != nil {
        log.Fatalf("Command execution failed: %v", err)
    }

    // Retrieve and print the command's exit code
    exitCode := int(op.Metadata["return"].(float64))
    fmt.Printf("Command executed with exit code: %d\n", exitCode)
}

