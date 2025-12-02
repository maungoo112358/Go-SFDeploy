package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("====  SmartFox Hot Deploy CLI Tool ====")
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
