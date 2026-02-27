package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadProgram(filename string, memory *Memory) (uint16, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("unable to open file: %v", err)
	}
	defer file.Close()

	return readProgramFromFile(file, memory)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var filename string

	for {
		fmt.Print("Enter program filename: ")
		scanner.Scan()
		filename = strings.TrimSpace(scanner.Text())

		if filename == "" {
			fmt.Fprintf(os.Stderr, "Error: Filename cannot be empty\n")
			continue
		}

		// Check if file exists
		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist.\n", filename)
			fmt.Print("Would you like to try again? (y/n): ")
			scanner.Scan()
			response := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if response != "y" && response != "yes" {
				fmt.Println("Exiting program.")
				os.Exit(0)
			}
			continue
		}

		break
	}

	processor, err := NewProcessor()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create processor: %v\n", err)
		os.Exit(1)
	}
	defer processor.Close()

	initialIP, err := loadProgram(filename, processor.memory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load program: %v\n", err)
		os.Exit(1)
	}

	processor.Reset(initialIP)
	processor.Run()
}
