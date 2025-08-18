package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	SourceDir       string `json:"source_dir"`
	TargetDir       string `json:"target_dir"`
	ExtensionFolder string `json:"extension_folder"`
	JavaPath        string `json:"java_path"`
}

const configFile = "sfdeploy_config.json"

var smartFoxCmdPid string // Global variable to store CMD PID for reuse

func main() {
	fmt.Println("========================================")
	fmt.Println("  SmartFox Hot Deploy CLI Tool")
	fmt.Println("========================================")
	fmt.Println()

	config := Config{}

	if !setupDirectories(&config) {
		return
	}

	if !buildProject(&config) {
		return
	}

	if !deployProject(&config) {
		return
	}

	if !restartServer(&config) {
		return
	}

	if !cleanupProject(&config) {
		return
	}

	fmt.Println("✅ Hot deploy completed successfully!")
	fmt.Println("Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadLine()
}

func setupDirectories(config *Config) bool {
	fmt.Println("📁 Phase 1: Directory Setup")

	if savedConfig, exists := loadConfig(); exists {
		fmt.Println("📋 Found previous configuration:")
		fmt.Printf("   Source: %s\n", savedConfig.SourceDir)
		fmt.Printf("   Target: %s\n", savedConfig.TargetDir)
		fmt.Printf("   Extension: %s\n", savedConfig.ExtensionFolder)
		fmt.Println()

		if askYesNo("Do you want to use the previous configuration? (y/n): ") {
			*config = savedConfig

			if validateSourceDir(config.SourceDir) && validateTargetDir(config.TargetDir) {
				config.JavaPath = findJava11Path()
				if config.JavaPath == "" {
					fmt.Println("❌ Java 11 not found")
					return false
				}

				fmt.Printf("✅ Using previous configuration\n")
				fmt.Printf("✅ Java 11: %s\n", config.JavaPath)
				fmt.Println()
				return true
			} else {
				fmt.Println("⚠️ Previous paths are no longer valid, please enter new ones")
			}
		}
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

	// Auto-detect SmartFox server installation
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

	// If no auto-detection or user declined, ask for manual input
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

func validateSourceDir(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}

	srcDir := filepath.Join(dir, "src")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return false
	}

	// Check for Java files recursively in src directory
	return hasJavaFiles(srcDir)
}

func hasJavaFiles(dir string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".java") {
			found = true
			return filepath.SkipDir // Stop walking once we find a Java file
		}
		return nil
	})
	return found
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

	// Check for SmartFox JAR files in the server's lib directory
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
	// Common SmartFox installation paths
	commonPaths := []string{
		"C:\\Users\\Slint\\SmartFoxServer_2X",        // Your specific path
		"C:\\SmartFoxServer_2X",                      // Common installation path
		"C:\\Program Files\\SmartFoxServer_2X",       // Program Files
		"C:\\Program Files (x86)\\SmartFoxServer_2X", // Program Files x86
	}

	// Check user's home directory for SmartFox installations
	if userHome, err := os.UserHomeDir(); err == nil {
		userPaths := []string{
			filepath.Join(userHome, "SmartFoxServer_2X"),
			filepath.Join(userHome, "SmartFoxServer"),
			filepath.Join(userHome, "Desktop", "SmartFoxServer_2X"),
			filepath.Join(userHome, "Downloads", "SmartFoxServer_2X"),
		}
		commonPaths = append(commonPaths, userPaths...)
	}

	// Check each potential path
	for _, path := range commonPaths {
		if validateTargetDir(path) {
			return path
		}
	}

	// If Windows, search common drive letters
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

				// Also check in Program Files
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

func buildProject(config *Config) bool {
	fmt.Println("🔨 Phase 2: Building Project")

	srcDir := filepath.Join(config.SourceDir, "src")

	// Use SmartFox server's lib directory for JAR files
	serverLibDir := filepath.Join(config.TargetDir, "SFS2X", "lib")

	fmt.Println("🧹 Cleaning old class files...")
	cleanClassFiles(srcDir)

	fmt.Println("⚙️ Compiling Java files...")
	javaFiles := findJavaFiles(srcDir)
	if len(javaFiles) == 0 {
		fmt.Println("❌ No Java files found")
		return false
	}

	fmt.Printf("📋 Found %d Java files\n", len(javaFiles))

	// Build classpath using SmartFox server's lib directory
	classpath := buildClasspath(serverLibDir)
	// fmt.Printf("📚 Using classpath: %s\n", classpath)

	javacPath := filepath.Join(config.JavaPath, "javac")
	if runtime.GOOS == "windows" {
		javacPath += ".exe"
	}

	args := []string{"-cp", classpath, "-d", srcDir}
	args = append(args, javaFiles...)

	cmd := exec.Command(javacPath, args...)
	cmd.Dir = srcDir

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("❌ Compilation failed: %s\n", string(output))
		return false
	}

	fmt.Println("✅ Compilation successful")

	fmt.Println("📦 Creating JAR file...")
	jarPath := filepath.Join(config.JavaPath, "jar")
	if runtime.GOOS == "windows" {
		jarPath += ".exe"
	}

	jarFile := filepath.Join(config.SourceDir, "ServerExtension.jar")

	// Create JAR with all compiled classes (including packages)
	cmd = exec.Command(jarPath, "cf", jarFile, ".")
	cmd.Dir = srcDir

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("❌ JAR creation failed: %s\n", string(output))
		return false
	}

	fmt.Println("✅ JAR file created successfully")
	fmt.Println()

	return true
}

func cleanClassFiles(srcDir string) {
	filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".class") {
			os.Remove(path)
		}
		return nil
	})
}

func findJavaFiles(srcDir string) []string {
	var javaFiles []string
	filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".java") {
			javaFiles = append(javaFiles, path)
		}
		return nil
	})
	return javaFiles
}

func buildClasspath(serverLibDir string) string {
	// Essential SmartFox JAR files
	requiredJars := []string{
		"sfs2x.jar",
		"sfs2x-core.jar",
		"sfs2x-api.jar",
		"slf4j-api*.jar",
		"logback*.jar",
	}

	var classpathParts []string

	// Add all JAR files from server lib directory
	jarFiles, _ := filepath.Glob(filepath.Join(serverLibDir, "*.jar"))
	for _, jarFile := range jarFiles {
		classpathParts = append(classpathParts, jarFile)
	}

	// If no JARs found, try the required ones specifically
	if len(classpathParts) == 0 {
		for _, jarPattern := range requiredJars {
			matches, _ := filepath.Glob(filepath.Join(serverLibDir, jarPattern))
			classpathParts = append(classpathParts, matches...)
		}
	}

	if len(classpathParts) == 0 {
		fmt.Printf("⚠️ Warning: No JAR files found in %s\n", serverLibDir)
		return "."
	}

	// Join classpath with appropriate separator
	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	return strings.Join(classpathParts, separator)
}

func killPort9933() {
	if runtime.GOOS != "windows" {
		return
	}

	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":9933") && strings.Contains(line, "LISTENING") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				pid := parts[len(parts)-1]
				fmt.Printf("🔫 Killing process %s using port 9933\n", pid)
				killCmd := exec.Command("taskkill", "/PID", pid, "/F")
				killCmd.Run()
			}
		}
	}
}

func findAndStoreSmartFoxCmdWindow() {
	if runtime.GOOS != "windows" {
		return
	}

	fmt.Println("🔍 Searching all CMD windows for SmartFox...")

	// Method 1: Check for any CMD that has Java children using port 9933
	javaCmd := exec.Command("wmic", "process", "where", "name='java.exe'", "get", "ProcessId,ParentProcessId,CommandLine", "/format:csv")
	javaOutput, err := javaCmd.Output()
	if err == nil {
		javaLines := strings.Split(string(javaOutput), "\n")
		for _, javaLine := range javaLines {
			if strings.Contains(javaLine, "java.exe") {
				fmt.Printf("🔍 Found Java process: %s\n", strings.TrimSpace(javaLine))

				// Check if this Java process is using port 9933
				parts := strings.Split(javaLine, ",")
				if len(parts) >= 3 {
					javaPid := strings.TrimSpace(parts[len(parts)-1])
					parentPid := strings.TrimSpace(parts[len(parts)-2])

					// Check if this Java process is listening on port 9933
					netstatCmd := exec.Command("netstat", "-ano")
					netstatOutput, err := netstatCmd.Output()
					if err == nil {
						netstatLines := strings.Split(string(netstatOutput), "\n")
						for _, netLine := range netstatLines {
							if strings.Contains(netLine, ":9933") && strings.Contains(netLine, "LISTENING") && strings.Contains(netLine, javaPid) {
								fmt.Printf("🎯 Found SmartFox Java process PID: %s with parent: %s\n", javaPid, parentPid)

								// Check if parent is CMD
								parentCmd := exec.Command("tasklist", "/fi", fmt.Sprintf("PID eq %s", parentPid), "/fo", "csv")
								parentOutput, err := parentCmd.Output()
								if err == nil && strings.Contains(string(parentOutput), "cmd.exe") {
									smartFoxCmdPid = parentPid
									fmt.Printf("✅ Found SmartFox CMD window PID: %s (parent of Java process)\n", smartFoxCmdPid)
									return
								}
								break
							}
						}
					}
				}
			}
		}
	}

	if smartFoxCmdPid == "" {
		fmt.Println("⚠️ Could not find SmartFox CMD window - will create new one")
	}
}

func deployProject(config *Config) bool {
	fmt.Println("🚀 Phase 3: Deploying Project")

	targetExtDir := filepath.Join(config.TargetDir, "SFS2X", "extensions", config.ExtensionFolder)

	if err := os.MkdirAll(targetExtDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create target directory: %v\n", err)
		return false
	}

	fmt.Printf("📁 Deploying to: %s\n", targetExtDir)

	// First, find and preserve CMD window BEFORE killing anything
	findAndStoreSmartFoxCmdWindow()

	// Now kill the processes
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

	fmt.Println("📋 Copying new JAR file...")
	sourceJar := filepath.Join(config.SourceDir, "ServerExtension.jar")
	targetJar := filepath.Join(targetExtDir, "ServerExtension.jar")

	if err := copyFile(sourceJar, targetJar); err != nil {
		fmt.Printf("❌ Failed to copy JAR file: %v\n", err)
		return false
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
		if strings.HasSuffix(strings.ToLower(info.Name()), ".class") {
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

func restartServer(config *Config) bool {
	fmt.Println("🔄 Phase 4: Restarting SmartFox Server")

	startScript := filepath.Join(config.TargetDir, "SFS2X", "sfs2x.bat")

	// Check if we have a stored CMD PID to reuse
	if smartFoxCmdPid != "" {
		fmt.Printf("🔍 Checking if stored CMD window PID %s is still alive...\n", smartFoxCmdPid)

		checkCmd := exec.Command("tasklist", "/fi", fmt.Sprintf("PID eq %s", smartFoxCmdPid), "/fo", "csv")
		checkOutput, err := checkCmd.Output()

		if err == nil && strings.Contains(string(checkOutput), "cmd.exe") {
			fmt.Println("✅ Found existing SmartFox CMD window")
			fmt.Println("🔄 Since we need to see logs, creating new CMD window...")

			// Close the old CMD window since we can't reuse it effectively
			exec.Command("taskkill", "/PID", smartFoxCmdPid, "/F").Run()
			fmt.Printf("🗑️ Closed old CMD window PID: %s\n", smartFoxCmdPid)
		}

		// Reset the stored PID
		smartFoxCmdPid = ""
	}

	// Always create a new window for logs visibility
	fmt.Println("▶️ Creating new CMD window for SmartFox server...")

	// Create a batch file with a nice header
	logBat := filepath.Join(config.TargetDir, "sfs_with_logs.bat")
	logContent := fmt.Sprintf(`@echo off
title SmartFox Server 2X - Hot Deploy
echo.
echo ========================================
echo   SmartFox Server 2X - Hot Deploy
echo   Starting server with logs...
echo ========================================
echo.
cd /d "%s"
call "%s"
echo.
echo ========================================
echo   Server stopped. Press any key to close.
echo ========================================
pause
`, filepath.Join(config.TargetDir, "SFS2X"), startScript)

	if err := os.WriteFile(logBat, []byte(logContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create log batch file: %v\n", err)
		return false
	}

	// Start the new CMD window with the log batch file
	cmd := exec.Command("cmd", "/c", "start", "cmd", "/k", logBat)
	cmd.Dir = filepath.Dir(logBat)

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to start server: %v\n", err)
		return false
	}

	// Clean up the batch file after a delay
	go func() {
		time.Sleep(5 * time.Second)
		os.Remove(logBat)
	}()

	fmt.Println("✅ Server started in new CMD window with logs")
	fmt.Println("📝 Check the new CMD window for server logs and status")
	fmt.Println()

	return true
}
