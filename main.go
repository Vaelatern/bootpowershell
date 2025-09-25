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

// Commands holds both PowerShell and CMD commands
type Commands struct {
	Powershell []string
	Cmd        []string
}

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

func loadCommands(dir string) (Commands, error) {
	var allCmds Commands

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
		allCmds.Powershell = append(allCmds.Powershell, cmds.Powershell...)
		allCmds.Cmd = append(allCmds.Cmd, cmds.Cmd...)
		return nil
	})
	return allCmds, err
}

func parseYAMLFile(path string) (Commands, error) {
	var root map[string]interface{}
	var cmds Commands

	data, err := os.ReadFile(path)
	if err != nil {
		return cmds, fmt.Errorf("read error: %w", err)
	}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return cmds, fmt.Errorf("YAML unmarshal error: %w", err)
	}

	// Check for PowerShell commands (raw_ps)
	if psRaw, psOk := root["raw_ps"]; psOk {
		items, ok := psRaw.([]interface{})
		if !ok {
			return cmds, fmt.Errorf("invalid raw_ps type in %s", path)
		}
		for _, v := range items {
			if str, ok := v.(string); ok {
				cmds.Powershell = append(cmds.Powershell, str)
			}
		}
	}

	// Check for CMD commands (raw_cmd)
	if cmdRaw, cmdOk := root["raw_cmd"]; cmdOk {
		items, ok := cmdRaw.([]interface{})
		if !ok {
			return cmds, fmt.Errorf("invalid raw_cmd type in %s", path)
		}
		for _, v := range items {
			if str, ok := v.(string); ok {
				cmds.Cmd = append(cmds.Cmd, str)
			}
		}
	}

	return cmds, nil
}

func runPs(line string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", line)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCmd(line string) error {
	cmd := exec.Command("cmd", "/C", line)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func help() {
	fmt.Printf("Usage: %s [install] E:\n", os.Args[0])
	fmt.Printf("\tPass install to use `schtasks` to install.\n")
	fmt.Printf("\tSet the last parameter to specify where yaml files can be found\n")
	fmt.Printf("\tCommands will be a list of strings under raw_ps or raw_cmd, appended in lexical order\n")
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

		if len(cmds.Powershell) == 0 && len(cmds.Cmd) == 0 {
			fmt.Println("No commands provided at " + os.Args[1])
		}

		for _, line := range cmds.Powershell {
			fmt.Println("Running PowerShell:", line)
			if err := runPs(line); err != nil {
				fmt.Printf("PowerShell command failed: %v\n", err)
			}
		}

		for _, line := range cmds.Cmd {
			fmt.Println("Running CMD:", line)
			if err := runCmd(line); err != nil {
				fmt.Printf("CMD command failed: %v\n", err)
			}
		}
	} else {
		help()
	}
}
