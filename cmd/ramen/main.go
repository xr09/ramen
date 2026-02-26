package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

func main() {
	var sizeStr string
	flag.StringVar(&sizeStr, "size", "", "Amount of memory to consume (e.g. 500, 500M, 1G). Defaults to MB.")
	flag.StringVar(&sizeStr, "s", "", "Alias for -size")
	flag.Parse()

	if err := run(sizeStr); err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
		fmt.Printf("Usage: %s -size <size>\n", os.Args[0])
		os.Exit(1)
	}
}

func parseSize(sizeStr string) (int, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("size must be specified")
	}

	str := strings.ToUpper(strings.TrimSpace(sizeStr))
	multiplier := 1 // Default to MB

	if strings.HasSuffix(str, "GB") {
		multiplier = 1024
		str = strings.TrimSuffix(str, "GB")
	} else if strings.HasSuffix(str, "G") {
		multiplier = 1024
		str = strings.TrimSuffix(str, "G")
	} else if strings.HasSuffix(str, "MB") {
		str = strings.TrimSuffix(str, "MB")
	} else if strings.HasSuffix(str, "M") {
		str = strings.TrimSuffix(str, "M")
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid memory size format: %s", sizeStr)
	}

	if val <= 0 {
		return 0, fmt.Errorf("please provide a positive value for memory to consume")
	}

	return val * multiplier, nil
}

func run(sizeStr string) error {
	sizeMB, err := parseSize(sizeStr)
	if err != nil {
		return err
	}

	mem := allocate(sizeMB)
	fmt.Printf("%sSuccessfully allocated and consumed %d MB of memory.%s\n", colorGreen, sizeMB, colorReset)
	fmt.Printf("%sPress Ctrl+C to exit and release the memory.%s\n", colorBlue, colorReset)

	// Wait for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Printf("\n%sExiting. Memory will be released.%s\n", colorYellow, colorReset)
	_ = mem // Keep reference to prevent GC, though it shouldn't happen while wait
	return nil
}

func allocate(sizeMB int) []byte {
	sizeBytes := sizeMB * 1024 * 1024
	fmt.Printf("%sAllocating %d MB of memory...%s\n", colorBlue, sizeMB, colorReset)

	// Allocate memory
	mem := make([]byte, sizeBytes)

	// Touch every 4KB page so that the OS actually provisions physical memory
	for i := 0; i < len(mem); i += 4096 {
		mem[i] = 1
	}
	return mem
}
