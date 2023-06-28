// Package zeusutils provides utilities for use in zeus commands that are implemented in Go.
package zeusutils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// LoadArg attempts to load the named zeus argument from os.Args.
// Args are passed to zeus commands in the name=value format on the commandline.
// Zeus will throw an error if not all required args are provided to your command,
// so this util does not validate the presence of args, but only loads the value.
// In addition, leading and trailing whitespace will be trimmed.
func LoadArg(name string) string {
	for _, arg := range os.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			if parts[0] == name {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// LoadArgs loads all zeus arguments into a map, with the argument label as keys.
func LoadArgs() map[string]string {
	args := make(map[string]string)
	for _, arg := range os.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			args[parts[0]] = strings.TrimSpace(parts[1])
		}
	}
	return args
}

// RequireEnv attempts to load the value of the named environment variable
// If no value for the given name has been found, the program will fatal.
func RequireEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatal("no value provided for required environment variable: ", name)
	}
	return val
}

// Prompt asks the user to provide input for the value with the given name,
// and returns the raw value to the caller.
func Prompt(name string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter %s: ", name)

	value, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("failed to read string from stdin: ", err)
	}

	return strings.ReplaceAll(
		strings.TrimSpace(value),
		"\n",
		"",
	)
}

// TrimStringLiterals will remove string literals, if the provided string starts AND ends with a literal.
// Eg:
// "value" => value
// 'value' => value
// "value => "value
func TrimStringLiterals(value string) string {
	if strings.HasPrefix(value, "'") &&
		strings.HasSuffix(value, "'") {
		return strings.Trim(value, "'")
	}
	if strings.HasPrefix(value, "\"") &&
		strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\"")
	}
	return value
}
