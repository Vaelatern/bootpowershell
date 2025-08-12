package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	taskName = "OnBootRunPowerShell"
)

func installTask(file string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}
	exePath = filepath.Clean(exePath)

	args := []string{
		"/Create",
		"/TN", taskName,
		"/TR", fmt.Sprintf(`%s %s`, exePath, file),
		"/SC", "ONSTART",
		"/RL", "HIGHEST",
		"/RU", "SYSTEM",
	}

	cmd := exec.Command("schtasks", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func loadCommands(dir string) ([]string, error) {
	var allCmds []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".yml") {
			return nil
		}
		cmds, err := parseYAMLFile(path)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil // keep going even on parse error
		} else {
			fmt.Printf("Found file to parse: %s\n", path)
		}
		allCmds = append(allCmds, cmds...)
		return nil
	})
	return allCmds, err
}

func parseYAMLFile(path string) ([]string, error) {
	var root map[string]interface{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("YAML unmarshal error: %w", err)
	}

	raw, ok := root["raw_cmd"]
	if !ok {
		return nil, nil
	}

	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid raw_cmd type in %s", path)
	}

	var cmds []string
	for _, v := range items {
		if str, ok := v.(string); ok {
			cmds = append(cmds, str)
		}
	}
	return cmds, nil
}

func help() {
	fmt.Printf("Usage: %s [install] E:\n", os.Args[0])
	fmt.Printf("\tPass install to use `schtasks` to install.\n")
	fmt.Printf("\tSet the last parameter to specify where yaml files can be found\n")
	fmt.Printf("\tCommands will be a list of strings under raw_cmd, appended in lexical order\n")
}

func main() {
	if len(os.Args) == 3 && os.Args[1] == "install" {
		if err := installTask(os.Args[2]); err != nil {
			fmt.Println("Installation failed:", err)
			os.Exit(1)
		}
		fmt.Println("Task Scheduler installation complete.")
		return
	} else if len(os.Args) == 2 {
		cmds, err := loadCommands(os.Args[1])
		if err != nil {
			fmt.Println("Error loading commands:", err)
			os.Exit(1)
		}

		if len(cmds) == 0 {
			fmt.Println("No commands provided at " + os.Args[1])
		}

		for _, line := range cmds {
			fmt.Println("Running:", line)
			cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", line)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Command failed: %v\n", err)
			}
		}
	} else {
		help()
	}
}
