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
	taskName  = "OnBootRunPowerShell"
	configDir = `D:\SetUpOnBoot`
)

func installTask() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}
	exePath = filepath.Clean(exePath)

	args := []string{
		"/Create",
		"/TN", taskName,
		"/TR", fmt.Sprintf(`"%s"`, exePath),
		"/SC", "ONSTART",
		"/RL", "HIGHEST",
		"/F",
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

func main() {
	if len(os.Args) > 1 && os.Args[1] == "install" {
		if err := installTask(); err != nil {
			fmt.Println("Installation failed:", err)
			os.Exit(1)
		}
		fmt.Println("Task Scheduler installation complete.")
		return
	}

	cmds, err := loadCommands(configDir)
	if err != nil {
		fmt.Println("Error loading commands:", err)
		os.Exit(1)
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
}
