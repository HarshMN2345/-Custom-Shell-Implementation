package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
			} else if inDoubleQuotes {
				if i+1 < len(input) && (input[i+1] == '\\' || input[i+1] == '$' || input[i+1] == '"' || input[i+1] == '\n') {
					escapeNext = true
				} else {
					current.WriteByte(char)
				}
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
			} else {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
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
func handleRedirection(command []string) (cmd []string, outputFile string, err error) {
	for i, part := range command {
		if strings.HasPrefix(part, ">") || strings.HasPrefix(part, "1>") {
			outputFile = strings.TrimPrefix(part, "1>")
			outputFile = strings.TrimPrefix(outputFile, ">")

			if i > 0 {
				cmd = command[:i]
			} else {
				cmd = []string{}
			}
			return
		}
	}
	cmd = command
	return
}

func main() {
	fmt.Fprint(os.Stdout, "$ ")

	f := bufio.NewReader(os.Stdin)
	for {
		input, err := f.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		input = strings.TrimSpace(input)

		commands := parseCommand(input)
		if len(commands) == 0 {
			fmt.Fprint(os.Stdout, "$ ")
			continue
		}

		commands, outputFile, err := handleRedirection(commands)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing command: %v\n", err)
			fmt.Fprint(os.Stdout, "$ ")
			continue
		}

		switch commands[0] {
		case "cd":
			if len(commands) < 2 {
				fmt.Fprintln(os.Stdout, "cd: missing operand")
			} else {
				path := commands[1]
				if path == "~" {
					path = os.Getenv("HOME")
				}
				err := os.Chdir(path)
				if err != nil {
					if os.IsNotExist(err) {
						fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", commands[1])
					} else {
						fmt.Fprintf(os.Stderr, "cd: %s: %v\n", commands[1], err)
					}
				}
			}
			fmt.Fprint(os.Stdout, "$ ")

		case "type":
			if len(commands) < 2 {
				fmt.Fprintln(os.Stdout, "type: missing operand")
			} else {
				switch commands[1] {
				case "echo":
					fmt.Fprintln(os.Stdout, "echo is a shell builtin")
				case "exit":
					fmt.Fprintln(os.Stdout, "exit is a shell builtin")
				case "type":
					fmt.Fprintln(os.Stdout, "type is a shell builtin")
				default:
					// Search for the command in PATH
					pathEnv := os.Getenv("PATH")
					paths := strings.Split(pathEnv, string(os.PathListSeparator))
					found := false
					for _, dir := range paths {
						executablePath := filepath.Join(dir, commands[1])
						if _, err := os.Stat(executablePath); err == nil {
							fmt.Fprintf(os.Stdout, "%s is %s\n", commands[1], executablePath)
							found = true
							break
						}
					}
					if !found {
						fmt.Fprintf(os.Stdout, "%s: not found\n", commands[1])
					}
				}
			}
			fmt.Fprint(os.Stdout, "$ ")

		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			} else {
				fmt.Fprintln(os.Stdout, dir)
			}
			fmt.Fprint(os.Stdout, "$ ")

		case "exit":
			os.Exit(0)

		case "echo":
			output := strings.Join(commands[1:], " ")
			if outputFile != "" {
				err := os.WriteFile(outputFile, []byte(output+"\n"), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
				}
			} else {
				fmt.Fprintln(os.Stdout, output)
			}
			fmt.Fprint(os.Stdout, "$ ")

		default:
			pathEnv := os.Getenv("PATH")
			paths := strings.Split(pathEnv, string(os.PathListSeparator))
			var executablePath string
			found := false

			for _, dir := range paths {
				executablePath = filepath.Join(dir, commands[0])
				if _, err := os.Stat(executablePath); err == nil {
					found = true
					break
				}
			}

			if found {
				cmd := exec.Command(executablePath, commands[1:]...)
				if outputFile != "" {
					file, err := os.Create(outputFile)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
						fmt.Fprint(os.Stdout, "$ ")
						continue
					}
					defer file.Close()
					cmd.Stdout = file
					cmd.Stderr = file
				} else {
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
				}

				err := cmd.Run()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stdout, "%s: command not found\n", commands[0])
			}
			fmt.Fprint(os.Stdout, "$ ")
		}
	}
}
