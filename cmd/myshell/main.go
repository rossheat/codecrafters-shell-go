package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var _ = fmt.Fprint

func exitGuard(command string) {
	commandParts := strings.Split(command, " ")
	if len(commandParts) > 0 {
		if commandParts[0] == "exit" {
			os.Exit(0)
		}
	}
}

func handleEcho(commandParts []string) {
	fmt.Println(strings.Join(commandParts[1:], " "))
}

func getPathEnvVarPaths() []string {
	for _, envVar := range os.Environ() {
		envVarParts := strings.Split(envVar, "=")
		if len(envVarParts) > 1 && envVarParts[0] == "PATH" {
			return strings.Split(envVarParts[1], ":")
		}
	}
	return []string{}
}

func searchPaths(paths []string, commandToInspect string) string {
	for _, path := range paths {
		possiblePathToCommandFile := filepath.Join(path, commandToInspect)
		if _, err := os.Stat(possiblePathToCommandFile); err == nil {
			return possiblePathToCommandFile
		}
	}
	return ""
}

func handleType(commandParts []string) {
	shellBuiltins := []string{"echo", "type", "exit", "pwd", "cd"}
	if len(commandParts) > 1 {
		commandToInspect := commandParts[1]

		// Search builtin shell commands
		for _, shellBuiltin := range shellBuiltins {
			if commandToInspect == shellBuiltin {
				fmt.Printf("%v is a shell builtin\n", commandToInspect)
				return
			}
		}

		// Search command programs on this machine
		paths := getPathEnvVarPaths()
		path := searchPaths(paths, commandToInspect)
		if path != "" {
			fmt.Printf("%v is %v\n", commandToInspect, path)
			return
		}

		// Command not found
		fmt.Printf("%v: not found\n", commandToInspect)
	}
}

func handlePwd() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get working directory")
	}
	fmt.Println(wd)
}

func getCommandParts(command string) []string {
	commandParts := strings.Split(command, " ")
	filteredCommandParts := make([]string, 0)
	for _, commandPart := range commandParts {
		// Filter out empty strings
		if commandPart != "" {
			filteredCommandParts = append(filteredCommandParts, commandPart)
		}
	}
	return filteredCommandParts
}

func handleExternalProgram(commandParts []string) bool {
	externalProgram := commandParts[0]
	paths := getPathEnvVarPaths()
	path := searchPaths(paths, externalProgram)
	if path != "" {
		output, err := exec.Command(path, commandParts[1:]...).Output()
		if err != nil {
			fmt.Printf("Error running external program %v:%v\n", externalProgram, err.Error())
		} else {
			fmt.Print(string(output))
		}
		return true
	}
	return false
}

func handleCd(commandParts []string) {

	// Nothing for the cd command to do
	if len(commandParts) < 2 {
		return
	}

	path := commandParts[1]

	var err error
	// Home directory expansion
	if path == "~" {
		path, err = os.UserHomeDir()
		if err != nil {
			fmt.Printf("Failed to find user's home directory %v", err.Error())
			return
		}
	}

	if err := os.Chdir(path); err != nil {
		fmt.Printf("cd: %v: No such file or directory\n", path)
	}

}

func main() {

	for {
		fmt.Fprint(os.Stdout, "$ ")
		// Wait for user input
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		command = strings.TrimSpace(command)

		if err != nil {
			fmt.Println("Error reading command: ", err.Error())
			os.Exit(1)
		}

		exitGuard(command)

		commandParts := getCommandParts(command)

		if len(commandParts) < 1 {
			// User entered an empty command
			continue
		}

		switch commandParts[0] {
		// Remember to update shellBuiltins slice
		case "echo":
			handleEcho(commandParts)
		case "type":
			handleType(commandParts)
		case "pwd":
			handlePwd()
		case "cd":
			handleCd(commandParts)
		default:
			found := handleExternalProgram(commandParts)
			if !found {
				fmt.Printf("%v: command not found\n", command)
			}
		}
	}
}
