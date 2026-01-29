package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	JavaPath        string   `json:"java_path"`
	SourceDir       string   `json:"source_dir"`
	TargetDir       string   `json:"target_dir"`
	ExtensionFolder string   `json:"extension_folder"`
	ExtensionFile   string   `json:"extension_file"`
	CommonFile      string   `json:"common_file"`
	CommonFolder    string   `json:"common_folder"`
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

func setupDirectories(config *Config) bool {
	fmt.Println("Phase 1: Directory Setup")

	savedConfig, exists := loadConfig()
	if !exists {
		fmt.Println("Config file not found: sfdeploy_config.json")
		return false
	}

	*config = savedConfig

	if !validateSourceDir(config.SourceDir) {
		fmt.Println("Source directory is invalid")
		return false
	}

	if !validateTargetDir(config.TargetDir) {
		fmt.Println("Target directory is invalid")
		return false
	}

	config.JavaPath = findJava11Path()
	if config.JavaPath == "" {
		fmt.Println("Java 11 not found")
		return false
	}

	fmt.Printf("Source: %s\n", config.SourceDir)
	fmt.Printf("Target: %s\n", config.TargetDir)
	fmt.Printf("Extension: %s\n", config.ExtensionFolder)
	fmt.Printf("Java 11: %s\n", config.JavaPath)
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
		fmt.Printf("Warning: sfs2x.jar not found at %s\n", sfs2xJar)
	}

	if _, err := os.Stat(sfs2xCoreJar); os.IsNotExist(err) {
		fmt.Printf("Warning: sfs2x-core.jar not found at %s\n", sfs2xCoreJar)
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
