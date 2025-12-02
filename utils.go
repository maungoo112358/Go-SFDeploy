package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func askYesNo(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		} else {
			fmt.Println("Please enter 'y' or 'n'")
		}
	}
}

func findJava11Path() string {
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		javacPath := filepath.Join(javaHome, "bin", "javac")
		if runtime.GOOS == "windows" {
			javacPath += ".exe"
		}
		if _, err := os.Stat(javacPath); err == nil {
			if isJava11(javacPath) {
				return filepath.Dir(javacPath)
			}
		}
	}

	if path, err := exec.LookPath("javac"); err == nil {
		if isJava11(path) {
			return filepath.Dir(path)
		}
	}

	if runtime.GOOS == "windows" {
		commonPaths := []string{
			"C:\\Program Files\\Eclipse Adoptium\\jdk-11*\\bin\\javac.exe",
			"C:\\Program Files\\Java\\jdk-11*\\bin\\javac.exe",
			"C:\\Program Files\\OpenJDK\\jdk-11*\\bin\\javac.exe",
			"C:\\Program Files (x86)\\Eclipse Adoptium\\jdk-11*\\bin\\javac.exe",
		}

		for _, pattern := range commonPaths {
			matches, _ := filepath.Glob(pattern)
			for _, path := range matches {
				if _, err := os.Stat(path); err == nil {
					if isJava11(path) {
						return filepath.Dir(path)
					}
				}
			}
		}
	}

	fmt.Println("❌ Java 11 not found automatically")
	fmt.Print("Please enter the path to Java 11 bin directory (or press Enter to skip): ")
	reader := bufio.NewReader(os.Stdin)
	userPath, _ := reader.ReadString('\n')
	userPath = strings.TrimSpace(userPath)

	if userPath != "" {
		javacPath := filepath.Join(userPath, "javac")
		if runtime.GOOS == "windows" {
			javacPath += ".exe"
		}
		if _, err := os.Stat(javacPath); err == nil {
			return userPath
		}
	}

	return ""
}

func findSmartFoxServer() string {
	var searchPaths []string

	if userHome, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(userHome, "SmartFoxServer_2X"),
			filepath.Join(userHome, "SmartFoxServer"),
			filepath.Join(userHome, "Desktop", "SmartFoxServer_2X"),
			filepath.Join(userHome, "Downloads", "SmartFoxServer_2X"),
		)
	}

	for _, path := range searchPaths {
		if validateTargetDir(path) {
			return path
		}
	}

	if runtime.GOOS == "windows" {
		drives := []string{"C:", "D:", "E:", "F:"}
		patterns := []string{
			"SmartFoxServer_2X",
			"SmartFoxServer",
			"SFS2X",
		}

		for _, drive := range drives {
			for _, pattern := range patterns {
				searchPath := filepath.Join(drive+"\\", pattern)
				if validateTargetDir(searchPath) {
					return searchPath
				}

				programFilesPath := filepath.Join(drive+"\\Program Files", pattern)
				if validateTargetDir(programFilesPath) {
					return programFilesPath
				}

				programFilesx86Path := filepath.Join(drive+"\\Program Files (x86)", pattern)
				if validateTargetDir(programFilesx86Path) {
					return programFilesx86Path
				}
			}
		}
	}

	return ""
}

func isJava11(javacPath string) bool {
	cmd := exec.Command(javacPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "11.") ||
		strings.Contains(outputStr, "javac 11")
}
