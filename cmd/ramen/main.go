package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
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
	var growTime int
	var waitTime int
	flag.StringVar(&sizeStr, "size", "", "Amount of memory to consume (e.g. 500, 500M, 1G). Defaults to MB.")
	flag.StringVar(&sizeStr, "s", "", "Alias for -size")
	flag.IntVar(&growTime, "time", 0, "Time in seconds to gradually grow memory to the target size (0 = allocate instantly)")
	flag.IntVar(&growTime, "t", 0, "Alias for -time")
	flag.IntVar(&waitTime, "wait", 0, "Wait time in seconds before starting memory allocation")
	flag.IntVar(&waitTime, "w", 0, "Alias for -wait")
	flag.Parse()

	if err := run(sizeStr, waitTime, growTime); err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
		fmt.Printf("Usage: %s -size <size> [-w <wait_seconds>] [-t <grow_seconds>]\n", os.Args[0])
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

func run(sizeStr string, waitSec, growSec int) error {
	sizeMB, err := parseSize(sizeStr)
	if err != nil {
		return err
	}

	if waitSec < 0 {
		return fmt.Errorf("wait time must be non-negative, got %d", waitSec)
	}
	if growSec < 0 {
		return fmt.Errorf("grow time must be non-negative, got %d", growSec)
	}

	// Set up signal handling early so we can cancel during wait/grow phases
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait phase
	if waitSec > 0 {
		fmt.Printf("%sWaiting %d seconds before starting allocation...%s\n", colorBlue, waitSec, colorReset)
		select {
		case <-time.After(time.Duration(waitSec) * time.Second):
			// done waiting
		case <-sigChan:
			fmt.Printf("\n%sInterrupted during wait phase. Exiting.%s\n", colorYellow, colorReset)
			return nil
		}
	}

	var mem []byte

	if growSec > 0 {
		// Gradual allocation: grow steadily over growSec seconds
		mem, err = allocateGradual(sizeMB, growSec, sigChan)
		if err != nil {
			return err
		}
	} else {
		// Instant allocation
		mem = allocate(sizeMB)
		fmt.Printf("%sSuccessfully allocated and consumed %d MB of memory.%s\n", colorGreen, sizeMB, colorReset)
	}

	fmt.Printf("%sPress Ctrl+C to exit and release the memory.%s\n", colorBlue, colorReset)

	// Wait for Ctrl+C
	<-sigChan

	fmt.Printf("\n%sExiting. Memory will be released.%s\n", colorYellow, colorReset)
	_ = mem // Keep reference to prevent GC
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

func allocateGradual(totalMB, seconds int, sigChan <-chan os.Signal) ([]byte, error) {
	totalBytes := totalMB * 1024 * 1024
	fmt.Printf("%sGradually allocating %d MB over %d seconds...%s\n", colorBlue, totalMB, seconds, colorReset)

	mem := make([]byte, totalBytes)

	// We'll touch pages in chunks spread over the duration.
	// Each tick we touch a proportional number of pages.
	totalPages := (totalBytes + 4095) / 4096
	tickInterval := 100 * time.Millisecond
	totalTicks := int(time.Duration(seconds) * time.Second / tickInterval)
	if totalTicks <= 0 {
		totalTicks = 1
	}

	pagesPerTick := totalPages / totalTicks
	if pagesPerTick <= 0 {
		pagesPerTick = 1
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	touchedPages := 0
	lastReportedMB := 0

	for touchedPages < totalPages {
		select {
		case <-sigChan:
			allocatedMB := (touchedPages * 4096) / (1024 * 1024)
			fmt.Printf("\n%sInterrupted during grow phase at %d MB / %d MB.%s\n", colorYellow, allocatedMB, totalMB, colorReset)
			return mem, nil
		case <-ticker.C:
			end := touchedPages + pagesPerTick
			if end > totalPages {
				end = totalPages
			}
			for p := touchedPages; p < end; p++ {
				offset := p * 4096
				if offset < totalBytes {
					mem[offset] = 1
				}
			}
			touchedPages = end

			// Log progress every 10% (or every MB for small allocations)
			currentMB := (touchedPages * 4096) / (1024 * 1024)
			stepMB := totalMB / 10
			if stepMB < 1 {
				stepMB = 1
			}
			if currentMB >= lastReportedMB+stepMB || touchedPages >= totalPages {
				fmt.Printf("%s  -> %d / %d MB allocated%s\n", colorBlue, currentMB, totalMB, colorReset)
				lastReportedMB = currentMB
			}
		}
	}

	fmt.Printf("%sSuccessfully allocated and consumed %d MB of memory over %d seconds.%s\n", colorGreen, totalMB, seconds, colorReset)
	return mem, nil
}
