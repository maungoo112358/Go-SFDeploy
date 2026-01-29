package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func buildProject(config *Config) bool {
	fmt.Println("Phase 2: Building Project")

	srcDir := filepath.Join(config.SourceDir, "src")
	serverLibDir := filepath.Join(config.TargetDir, "SFS2X", "lib")

	fmt.Println("Cleaning old class files...")
	cleanClassFiles(srcDir)

	fmt.Println("Compiling Java files...")
	javaFiles := findJavaFiles(srcDir)
	if len(javaFiles) == 0 {
		fmt.Println("No Java files found")
		return false
	}

	fmt.Printf("Found %d Java files\n", len(javaFiles))

	classpath := buildClasspath(serverLibDir)

	javacPath := filepath.Join(config.JavaPath, "javac")
	if runtime.GOOS == "windows" {
		javacPath += ".exe"
	}

	args := []string{"-cp", classpath, "-d", srcDir}
	args = append(args, javaFiles...)

	cmd := exec.Command(javacPath, args...)
	cmd.Dir = srcDir

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Compilation failed: %s\n", string(output))
		return false
	}

	fmt.Println("Compilation successful")

	jarPath := filepath.Join(config.JavaPath, "jar")
	if runtime.GOOS == "windows" {
		jarPath += ".exe"
	}

	// Create SpookyCommon.jar from just the common folder
	if config.CommonFile != "" && config.CommonFolder != "" {
		fmt.Printf("Creating %s...\n", config.CommonFile)
		commonJarFile := filepath.Join(config.SourceDir, config.CommonFile)
		commonDir := filepath.Join(srcDir, config.CommonFolder)

		cmd = exec.Command(jarPath, "cf", commonJarFile, ".")
		cmd.Dir = commonDir

		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("JAR creation failed for %s: %s\n", config.CommonFile, string(output))
			return false
		}
		fmt.Printf("%s created successfully\n", config.CommonFile)
	}

	// Create main extension JAR from the whole src folder
	fmt.Printf("Creating %s...\n", config.ExtensionFile)
	extensionJarFile := filepath.Join(config.SourceDir, config.ExtensionFile)

	cmd = exec.Command(jarPath, "cf", extensionJarFile, ".")
	cmd.Dir = srcDir

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("JAR creation failed for %s: %s\n", config.ExtensionFile, string(output))
		return false
	}

	fmt.Printf("%s created successfully\n", config.ExtensionFile)
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
	requiredJars := []string{
		"sfs2x.jar",
		"sfs2x-core.jar",
		"sfs2x-api.jar",
		"slf4j-api*.jar",
		"logback*.jar",
	}

	var classpathParts []string

	jarFiles, _ := filepath.Glob(filepath.Join(serverLibDir, "*.jar"))
	for _, jarFile := range jarFiles {
		classpathParts = append(classpathParts, jarFile)
	}

	if len(classpathParts) == 0 {
		for _, jarPattern := range requiredJars {
			matches, _ := filepath.Glob(filepath.Join(serverLibDir, jarPattern))
			classpathParts = append(classpathParts, matches...)
		}
	}

	if len(classpathParts) == 0 {
		fmt.Printf("Warning: No JAR files found in %s\n", serverLibDir)
		return "."
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	return strings.Join(classpathParts, separator)
}
