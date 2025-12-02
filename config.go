package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	SourceDir       string   `json:"source_dir"`
	TargetDir       string   `json:"target_dir"`
	ExtensionFolder string   `json:"extension_folder"`
	JavaPath        string   `json:"java_path"`
	JsonSourceDir   string   `json:"json_source_dir"`
	DeployJsonFiles []string `json:"deploy_json_files"`
}

const configFile = "sfdeploy_config.json"

func loadConfig() (Config, bool) {
	var config Config

	data, err := os.ReadFile(configFile)
	if err != nil {
		return config, false
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, false
	}

	return config, true
}

func saveConfig(config Config) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(configFile, data, 0644)
}

func setupDirectories(config *Config) bool {
	fmt.Println("📁 Phase 1: Directory Setup")

	if savedConfig, exists := loadConfig(); exists {
		*config = savedConfig

		if validateSourceDir(config.SourceDir) && validateTargetDir(config.TargetDir) {
			config.JavaPath = findJava11Path()
			if config.JavaPath == "" {
				fmt.Println("❌ Java 11 not found")
				return false
			}

			fmt.Printf("✅ Source: %s\n", config.SourceDir)
			fmt.Printf("✅ Target: %s\n", config.TargetDir)
			fmt.Printf("✅ Extension: %s\n", config.ExtensionFolder)
			fmt.Printf("✅ Java 11: %s\n", config.JavaPath)
			fmt.Println()
			return true
		}
		fmt.Println("⚠️ Config paths are no longer valid, please enter new ones")
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter source directory (SmartFox project): ")
		sourceDir, _ := reader.ReadString('\n')
		config.SourceDir = strings.TrimSpace(sourceDir)

		if validateSourceDir(config.SourceDir) {
			fmt.Printf("✅ Valid source directory: %s\n", config.SourceDir)
			break
		} else {
			fmt.Printf("❌ Invalid source directory: %s\n", config.SourceDir)
			fmt.Println("   Please ensure the directory contains 'src' folder with .java files")
		}
	}

	autoDetectedTarget := findSmartFoxServer()
	if autoDetectedTarget != "" {
		fmt.Printf("🔍 Auto-detected SmartFox server: %s\n", autoDetectedTarget)
		if askYesNo("Do you want to use this SmartFox server installation? (y/n): ") {
			config.TargetDir = autoDetectedTarget
			fmt.Printf("✅ Using auto-detected target directory: %s\n", config.TargetDir)
		} else {
			config.TargetDir = ""
		}
	}

	if config.TargetDir == "" {
		for {
			fmt.Print("Enter target directory (SmartFox server): ")
			targetDir, _ := reader.ReadString('\n')
			config.TargetDir = strings.TrimSpace(targetDir)

			if validateTargetDir(config.TargetDir) {
				fmt.Printf("✅ Valid target directory: %s\n", config.TargetDir)
				break
			} else {
				fmt.Printf("❌ Invalid target directory: %s\n", config.TargetDir)
				fmt.Println("   Please ensure the directory contains 'SFS2X/sfs2x.bat' and 'SFS2X/lib/sfs2x.jar'")
			}
		}
	}

	for {
		fmt.Print("Enter extension folder name (e.g., SFServer, MyExtension): ")
		extensionFolder, _ := reader.ReadString('\n')
		config.ExtensionFolder = strings.TrimSpace(extensionFolder)

		if config.ExtensionFolder != "" && !strings.Contains(config.ExtensionFolder, ":") && !strings.Contains(config.ExtensionFolder, "\\") && !strings.Contains(config.ExtensionFolder, "/") {
			fmt.Printf("✅ Valid extension folder: %s\n", config.ExtensionFolder)
			break
		} else {
			fmt.Println("❌ Invalid extension folder name")
			fmt.Println("   Please enter a simple folder name (no paths, colons, or slashes)")
		}
	}

	config.JavaPath = findJava11Path()
	if config.JavaPath == "" {
		fmt.Println("❌ Java 11 not found")
		return false
	}

	fmt.Printf("✅ Java 11: %s\n", config.JavaPath)
	fmt.Println()

	saveConfig(*config)
	fmt.Println("💾 Configuration saved for next time")
	fmt.Println()

	return true
}

func validateSourceDir(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}

	srcDir := filepath.Join(dir, "src")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return false
	}

	return hasJavaFiles(srcDir)
}

func validateTargetDir(dir string) bool {
	sfsDir := filepath.Join(dir, "SFS2X")
	if _, err := os.Stat(sfsDir); os.IsNotExist(err) {
		return false
	}

	startScript := filepath.Join(sfsDir, "sfs2x.bat")
	if _, err := os.Stat(startScript); os.IsNotExist(err) {
		return false
	}

	libDir := filepath.Join(sfsDir, "lib")
	sfs2xJar := filepath.Join(libDir, "sfs2x.jar")
	sfs2xCoreJar := filepath.Join(libDir, "sfs2x-core.jar")

	if _, err := os.Stat(sfs2xJar); os.IsNotExist(err) {
		fmt.Printf("⚠️ Warning: sfs2x.jar not found at %s\n", sfs2xJar)
	}

	if _, err := os.Stat(sfs2xCoreJar); os.IsNotExist(err) {
		fmt.Printf("⚠️ Warning: sfs2x-core.jar not found at %s\n", sfs2xCoreJar)
	}

	return true
}

func hasJavaFiles(dir string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".java") {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	return found
}
