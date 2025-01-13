package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func main() {
	replaceFlag := flag.Bool("replace", false, "Replace the original file")
	ignoreMessages := flag.String("ignore", "", "Ignore some Message blocks")
	var ignoreMsgSlice []string
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: proto-renumber [--replace] PROTO_FILE_PATH.proto")
		return
	}

	// clean up the list
	if *ignoreMessages != "" {
		ignoreMsgSlice = strings.Split(*ignoreMessages, ",")
		for i := 0; i < len(ignoreMsgSlice); i++ {
			ignoreMsgSlice[i] = strings.TrimSpace(ignoreMsgSlice[i])
		}
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
	messageRegex := regexp.MustCompile(`^\s*message\s+(\w+)\s*\{`)
	closeMessageRegex := regexp.MustCompile(`^\s*\}`)
	scanner := bufio.NewScanner(file)

	ignoreStack := []bool{false}

	for scanner.Scan() {
		line := scanner.Text()

		// if we encounter a "message" declaration, push a new numbering context
		if messageRegex.MatchString(line) {
			// ignore if needed

			messageName := messageRegex.FindStringSubmatch(line)
			ignoreCurrent := false
			// means we have a match and we need to ignore it
			if len(ignoreMsgSlice) > 0 && len(messageName) > 1 && slices.Contains(ignoreMsgSlice, messageName[1]) {
				fmt.Println("Ignoring", messageName[1])
				ignoreCurrent = true
			}

			ignoreStack = append(ignoreStack, ignoreCurrent)
			numberStack = append(numberStack, 1)

			outputLines = append(outputLines, line)
			continue
		}

		// if we encounter a closing brace, pop the current numbering context
		if closeMessageRegex.MatchString(line) {
			outputLines = append(outputLines, line)
			if len(numberStack) > 1 {
				numberStack = numberStack[:len(numberStack)-1]
			}
			if len(ignoreStack) > 1 {
				ignoreStack = ignoreStack[:len(ignoreStack)-1]
			}
			continue
		}

		// check if the line has a field number
		if fieldRegex.MatchString(line) && !ignoreStack[len(ignoreStack)-1] {
			currentNumber := numberStack[len(numberStack)-1]
			replacement := "${1}" + strconv.Itoa(currentNumber) + "${3}"
			updatedLine := fieldRegex.ReplaceAllString(line, replacement)
			outputLines = append(outputLines, updatedLine)
			// increment the number for the current context
			numberStack[len(numberStack)-1]++
			continue
		}

		// Default case: Add the line without modifications
		outputLines = append(outputLines, line)
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
