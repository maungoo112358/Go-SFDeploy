package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var smartFoxCmdPid string

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

	javaCmd := exec.Command("wmic", "process", "where", "name='java.exe'", "get", "ProcessId,ParentProcessId,CommandLine", "/format:csv")
	javaOutput, err := javaCmd.Output()
	if err == nil {
		javaLines := strings.Split(string(javaOutput), "\n")
		for _, javaLine := range javaLines {
			if strings.Contains(javaLine, "java.exe") {
				fmt.Printf("🔍 Found Java process: %s\n", strings.TrimSpace(javaLine))

				parts := strings.Split(javaLine, ",")
				if len(parts) >= 3 {
					javaPid := strings.TrimSpace(parts[len(parts)-1])
					parentPid := strings.TrimSpace(parts[len(parts)-2])

					netstatCmd := exec.Command("netstat", "-ano")
					netstatOutput, err := netstatCmd.Output()
					if err == nil {
						netstatLines := strings.Split(string(netstatOutput), "\n")
						for _, netLine := range netstatLines {
							if strings.Contains(netLine, ":9933") && strings.Contains(netLine, "LISTENING") && strings.Contains(netLine, javaPid) {
								fmt.Printf("🎯 Found SmartFox Java process PID: %s with parent: %s\n", javaPid, parentPid)

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

func restartServer(config *Config) bool {
	fmt.Println("🔄 Phase 4: Restarting SmartFox Server")

	startScript := filepath.Join(config.TargetDir, "SFS2X", "sfs2x.bat")

	if smartFoxCmdPid != "" {
		fmt.Printf("🔍 Checking if stored CMD window PID %s is still alive...\n", smartFoxCmdPid)

		checkCmd := exec.Command("tasklist", "/fi", fmt.Sprintf("PID eq %s", smartFoxCmdPid), "/fo", "csv")
		checkOutput, err := checkCmd.Output()

		if err == nil && strings.Contains(string(checkOutput), "cmd.exe") {
			fmt.Println("✅ Found existing SmartFox CMD window")
			fmt.Println("🔄 Since we need to see logs, creating new CMD window...")

			exec.Command("taskkill", "/PID", smartFoxCmdPid, "/F").Run()
			fmt.Printf("🗑️ Closed old CMD window PID: %s\n", smartFoxCmdPid)
		}

		smartFoxCmdPid = ""
	}

	fmt.Println("▶️ Creating new CMD window for SmartFox server...")

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

	cmd := exec.Command("cmd", "/c", "start", "cmd", "/k", logBat)
	cmd.Dir = filepath.Dir(logBat)

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to start server: %v\n", err)
		return false
	}

	go func() {
		time.Sleep(5 * time.Second)
		os.Remove(logBat)
	}()

	fmt.Println("✅ Server started in new CMD window with logs")
	fmt.Println("📝 Check the new CMD window for server logs and status")
	fmt.Println()

	return true
}
