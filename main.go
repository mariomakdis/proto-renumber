package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func main() {
	replaceFlag := flag.Bool("replace", false, "Replace the original file")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: proto-renumber [--replace] PROTO_FILE_PATH.proto")
		return
	}

	filePath := flag.Arg(0)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var outputLines []string
	numberStack := []int{1}

	fieldRegex := regexp.MustCompile(`(=\s*)(\d+)(;)`)
	messageRegex := regexp.MustCompile(`^\s*message\s+\w+\s*\{`)
	closeMessageRegex := regexp.MustCompile(`^\s*\}`)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// if we encounter a "message" declaration, push a new numbering context
		if messageRegex.MatchString(line) {
			outputLines = append(outputLines, line)
			numberStack = append(numberStack, 1)
			continue
		}

		// if we encounter a closing brace, pop the current numbering context
		if closeMessageRegex.MatchString(line) {
			outputLines = append(outputLines, line)
			if len(numberStack) > 1 {
				numberStack = numberStack[:len(numberStack)-1]
			}
			continue
		}

		// check if the line has a field number
		if fieldRegex.MatchString(line) {
			currentNumber := numberStack[len(numberStack)-1]
			replacement := "${1}" + strconv.Itoa(currentNumber) + "${3}"
			updatedLine := fieldRegex.ReplaceAllString(line, replacement)
			outputLines = append(outputLines, updatedLine)
			// increment the number for the current context
			numberStack[len(numberStack)-1]++
			continue
		} else {
			outputLines = append(outputLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	outputFilePath := "updated_" + filePath

	// replace the original file if --replace is set
	if *replaceFlag {
		outputFilePath = filePath
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range outputLines {
		writer.WriteString(line + "\n")
	}
	writer.Flush()
	fmt.Printf("Updated file saved to %s\n", outputFilePath)
}
