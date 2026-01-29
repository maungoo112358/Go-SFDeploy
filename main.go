package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("====  SpookyZone Hot Deploy CLI Tool ====")
	fmt.Println()

	config := Config{}

	if !setupDirectories(&config) {
		waitAndExit()
		return
	}

	if !buildProject(&config) {
		waitAndExit()
		return
	}

	if !deployProject(&config) {
		waitAndExit()
		return
	}

	if !restartServer(&config) {
		waitAndExit()
		return
	}

	if !cleanupProject(&config) {
		waitAndExit()
		return
	}

	fmt.Println("Hot deploy completed successfully!")
	waitAndExit()
}

func waitAndExit() {
	fmt.Println()
	fmt.Println("Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadLine()
}
