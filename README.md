# Custom Shell in Golang

This is a custom shell implemented in Golang that supports various basic functionalities such as quoting, navigation, and execution of commands. The project is designed to give an understanding of how shell environments work under the hood by implementing features like handling quotes, paths, and external program execution.

## Features

- **Quoting Support**: Handles both single and double quotes, allowing commands with spaces or special characters to be parsed and executed correctly.
- **Path Navigation**: Supports `cd` (change directory) and `pwd` (print working directory).
- **Command Execution**: Allows running external commands with argument parsing.
- **Command Chaining**: Implements logical operators like `&&` (AND), `||` (OR).
- **Redirection**: Supports input/output redirection using `>`, `<`, and `>>`.
- **Error Handling**: Provides error messages for invalid commands or directories.
- **History**: Stores and retrieves previously executed commands (if implemented).
- **Interactive Mode**: Allows interactive command-line input.
- **Script Mode**: Allows execution of shell scripts.

## Installation

To run the custom shell on your system, follow these steps:

### Prerequisites

- Go 1.18 or later
- Linux, macOS, or Windows with WSL2 (for Windows)

### Clone the Repository

```bash
git clone https://github.com/HarshMN2345/Custom-Shell-Implementation
cd Shell
