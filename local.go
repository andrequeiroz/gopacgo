package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
)

// pac represents a locally installed package and its current version.
type pac struct {
	name    string
	version string
}

// getPac runs pacman -Qm to list all foreign (AUR) packages installed on the system.
func getPac() (string, error) {
	cmd := exec.Command("pacman", "-Qm")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}

// parsePac parses the raw output of pacman -Qm into a slice of pac structs.
// Each line has the format: "<name> <version>".
func parsePac(output string) []pac {
	var pacs []pac

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		record := scanner.Text()
		if record == "" {
			continue
		}

		fields := strings.Fields(record)
		if len(fields) == 2 {
			pacs = append(pacs, pac{name: fields[0], version: fields[1]})
		}
	}

	return pacs
}
