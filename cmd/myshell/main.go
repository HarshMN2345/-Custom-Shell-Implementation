package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// parseCommand parses the user input and handles quotes and escaping
func parseCommand(input string) []string {
	var result []string
	var current strings.Builder
	inDoubleQuotes := false
	inSingleQuotes := false
	escapeNext := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		if escapeNext {
			current.WriteByte(char)
			escapeNext = false
			continue
		}

		switch char {
		case '\\':
			if inSingleQuotes {
				current.WriteByte(char)
			} else {
				escapeNext = true
			}
		case '"':
			if inSingleQuotes {
				current.WriteByte(char)
			} else {
				inDoubleQuotes = !inDoubleQuotes
			}
		case '\'':
			if inDoubleQuotes {
				current.WriteByte(char)
			} else {
				inSingleQuotes = !inSingleQuotes
			}
		case ' ':
			if inSingleQuotes || inDoubleQuotes {
				current.WriteByte(char)
			} else if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// handleRedirection handles commands with output redirection
func handleRedirection(command []string) ([]string, string, bool, error) {
	for i, part := range command {
		switch {
		case part == ">", part == "1>":
			if i+1 >= len(command) {
				return nil, "", false, fmt.Errorf("missing output file")
			}
			return command[:i], command[i+1], false, nil
		case part == ">>":
			if i+1 >= len(command) {
				return nil, "", false, fmt.Errorf("missing output file")
			}
			return command[:i], command[i+1], true, nil
		case strings.HasPrefix(part, ">"):
			return command[:i], strings.TrimPrefix(part, ">"), false, nil
		case strings.HasPrefix(part, ">>"):
			return command[:i], strings.TrimPrefix(part, ">>"), true, nil
		}
	}
	return command, "", false, nil
}

// executeCommand handles the execution of commands
func executeCommand(command []string, outputFile string, appendMode bool) (int, error) {
	if len(command) == 0 {
		return 0, nil
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin

	if outputFile != "" {
		flag := os.O_CREATE | os.O_WRONLY
		if appendMode {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}

		file, err := os.OpenFile(outputFile, flag, 0644)
		if err != nil {
			return 1, fmt.Errorf("error with output file: %v", err)
		}
		defer file.Close()
		cmd.Stdout = file
		cmd.Stderr = file
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), err
		}
		return 1, err
	}
	return 0, nil
}

// executeBuiltin handles built-in shell commands
func executeBuiltin(commands []string) (bool, error) {
	switch commands[0] {
	case "cd":
		if len(commands) < 2 {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return true, fmt.Errorf("could not get home directory: %v", err)
			}
			return true, os.Chdir(homeDir)
		}
		path := commands[1]
		if path == "~" {
			var err error
			path, err = os.UserHomeDir()
			if err != nil {
				return true, fmt.Errorf("could not get home directory: %v", err)
			}
		}
		return true, os.Chdir(path)

	case "pwd":
		if dir, err := os.Getwd(); err != nil {
			return true, err
		} else {
			fmt.Println(dir)
		}
		return true, nil

	case "exit":
		os.Exit(0)

	case "type":
		if len(commands) < 2 {
			return true, fmt.Errorf("type: missing operand")
		}
		switch commands[1] {
		case "cd", "pwd", "exit", "type", "history":
			fmt.Printf("%s is a shell builtin\n", commands[1])
		default:
			if path, err := exec.LookPath(commands[1]); err != nil {
				fmt.Printf("%s: not found\n", commands[1])
			} else {
				fmt.Printf("%s is %s\n", commands[1], path)
			}
		}
		return true, nil
	}
	return false, nil
}

func main() {
	var history []string
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		history = append(history, input)
		if input == "history" {
			for i, cmd := range history {
				fmt.Printf("%d: %s\n", i+1, cmd)
			}
			continue
		}

		commands := parseCommand(input)
		if len(commands) == 0 {
			continue
		}

		// Handle built-in commands first
		if isBuiltin, err := executeBuiltin(commands); isBuiltin {
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			continue
		}

		// Handle command with redirection
		cmd, outputFile, appendMode, err := handleRedirection(commands)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		exitCode, err := executeCommand(cmd, outputFile, appendMode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if exitCode != 0 {
			continue
		}
	}
}