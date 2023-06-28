// Package zeusutils provides utilities for use in zeus commands that are implemented in Go.
package zeusutils

import (
	"log"
	"os"
	"strings"
)

// LoadArg attempts to load the named zeus argument from os.Args.
// Args are passed to zeus commands in the name=value format on the commandline.
// Zeus will throw an error if not all required args are provided to your command,
// so this util does not validate the presence of args, but only loads the value.
// In addition, leading and trailing string literals will be trimmed from the value (" and ' characters).
func LoadArg(name string) string {
	for _, arg := range os.Args {
		parts := strings.Split(arg, "=")
		if len(parts) > 1 {
			if parts[0] == name {
				return strings.Trim(strings.Join(parts[1:], "="), "\"'")
			}
		}
	}
	return ""
}

// RequireEnv attempts to load the value of the named environment variable
// If no value for the given name has been found, the program will fatal.
func RequireEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatal("no value provided for required environment variable: " + name)
	}
	return val
}
