package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func deployProject(config *Config) bool {
	fmt.Println("🚀 Phase 3: Deploying Project")

	targetExtDir := filepath.Join(config.TargetDir, "SFS2X", "extensions", config.ExtensionFolder)

	if err := os.MkdirAll(targetExtDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create target directory: %v\n", err)
		return false
	}

	fmt.Printf("📁 Deploying to: %s\n", targetExtDir)

	findAndStoreSmartFoxCmdWindow()

	fmt.Println("🔍 Killing processes on port 9933...")
	killPort9933()

	fmt.Println("⏳ Waiting for file locks to release...")
	time.Sleep(3 * time.Second)

	fmt.Println("🗑️ Removing old JAR files...")
	jarFiles, _ := filepath.Glob(filepath.Join(targetExtDir, "*.jar"))
	for _, file := range jarFiles {
		if err := os.Remove(file); err != nil {
			fmt.Printf("⚠️ Warning: Could not remove %s: %v\n", file, err)
		}
	}

	fmt.Println("Copying JAR files...")

	// Copy common JAR to __lib__ folder if configured
	if config.CommonFile != "" {
		libDir := filepath.Join(config.TargetDir, "SFS2X", "extensions", "__lib__")
		if err := os.MkdirAll(libDir, 0755); err != nil {
			fmt.Printf("Failed to create __lib__ directory: %v\n", err)
			return false
		}

		sourceCommonJar := filepath.Join(config.SourceDir, config.CommonFile)
		targetCommonJar := filepath.Join(libDir, config.CommonFile)

		if err := copyFile(sourceCommonJar, targetCommonJar); err != nil {
			fmt.Printf("Failed to copy %s: %v\n", config.CommonFile, err)
			return false
		}
		fmt.Printf("Copied: %s -> __lib__/\n", config.CommonFile)
	}

	// Copy main extension JAR to extension folder
	sourceJar := filepath.Join(config.SourceDir, config.ExtensionFile)
	targetJar := filepath.Join(targetExtDir, config.ExtensionFile)

	if err := copyFile(sourceJar, targetJar); err != nil {
		fmt.Printf("Failed to copy %s: %v\n", config.ExtensionFile, err)
		return false
	}
	fmt.Printf("Copied: %s -> %s/\n", config.ExtensionFile, config.ExtensionFolder)

	if len(config.DeployJsonFiles) > 0 {
		fmt.Printf("📋 Copying %d JSON files...\n", len(config.DeployJsonFiles))
		for _, jsonFile := range config.DeployJsonFiles {
			jsonFileName := jsonFile + ".json"
			sourceJson := filepath.Join(config.JsonSourceDir, jsonFileName)
			targetJson := filepath.Join(targetExtDir, jsonFileName)

			if _, err := os.Stat(sourceJson); os.IsNotExist(err) {
				fmt.Printf("⚠️ Warning: JSON file not found: %s\n", jsonFileName)
				continue
			}

			if err := copyFile(sourceJson, targetJson); err != nil {
				fmt.Printf("❌ Failed to copy JSON file %s: %v\n", jsonFileName, err)
				return false
			}
			fmt.Printf("   ✅ Copied: %s\n", jsonFileName)
		}
	}

	fmt.Println("✅ Deployment successful")
	fmt.Println()

	return true
}

func cleanupProject(config *Config) bool {
	fmt.Println("🧹 Phase 5: Cleaning Up Project")

	srcDir := filepath.Join(config.SourceDir, "src")

	fmt.Println("🗑️ Removing .class files from source directory...")
	classFilesRemoved := 0
	filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if filepath.Ext(info.Name()) == ".class" {
			if err := os.Remove(path); err == nil {
				classFilesRemoved++
			}
		}
		return nil
	})

	fmt.Printf("🗑️ Removed %d .class files\n", classFilesRemoved)

	fmt.Println("🗑️ Removing JAR files from project root...")
	jarFilesRemoved := 0
	jarFiles, _ := filepath.Glob(filepath.Join(config.SourceDir, "*.jar"))
	for _, file := range jarFiles {
		if err := os.Remove(file); err == nil {
			jarFilesRemoved++
			fmt.Printf("   Removed: %s\n", filepath.Base(file))
		} else {
			fmt.Printf("⚠️ Warning: Could not remove %s: %v\n", filepath.Base(file), err)
		}
	}

	fmt.Printf("🗑️ Removed %d JAR files\n", jarFilesRemoved)

	fmt.Println("✅ Project cleanup completed")
	fmt.Println()

	return true
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
