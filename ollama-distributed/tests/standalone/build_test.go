package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BuildTester tests individual packages to identify build issues
type BuildTester struct {
	packages []string
	results  map[string]BuildResult
}

type BuildResult struct {
	Package  string
	Success  bool
	Error    string
	Duration time.Duration
}

func main() {
	fmt.Println("🔧 OllamaMax Build Issue Identifier")
	fmt.Println("===================================")

	tester := &BuildTester{
		packages: []string{
			"./pkg/config",
			"./pkg/proxy",
			"./cmd/node",
		},
		results: make(map[string]BuildResult),
	}

	tester.runTests()
	tester.printResults()
}

func (bt *BuildTester) runTests() {
	for _, pkg := range bt.packages {
		fmt.Printf("Testing package: %s\n", pkg)
		result := bt.testPackage(pkg)
		bt.results[pkg] = result
	}
}

func (bt *BuildTester) testPackage(pkg string) BuildResult {
	start := time.Now()

	// Set a timeout for the build command
	cmd := exec.Command("timeout", "30s", "go", "build", "-o", "/dev/null", pkg)
	output, err := cmd.CombinedOutput()

	duration := time.Since(start)

	result := BuildResult{
		Package:  pkg,
		Success:  err == nil,
		Duration: duration,
	}

	if err != nil {
		result.Error = string(output)
		if len(result.Error) > 500 {
			result.Error = result.Error[:500] + "... (truncated)"
		}
	}

	return result
}

func (bt *BuildTester) printResults() {
	fmt.Println("\n📊 Build Test Results")
	fmt.Println("====================")

	successCount := 0
	for pkg, result := range bt.results {
		status := "❌ FAILED"
		if result.Success {
			status = "✅ SUCCESS"
			successCount++
		}

		fmt.Printf("%s %s (%.2fs)\n", status, pkg, result.Duration.Seconds())

		if !result.Success && result.Error != "" {
			fmt.Printf("   Error: %s\n", strings.TrimSpace(result.Error))
		}
	}

	fmt.Printf("\n📈 Summary: %d/%d packages built successfully\n", successCount, len(bt.packages))

	if successCount == len(bt.packages) {
		fmt.Println("🎉 All packages build successfully!")
	} else {
		fmt.Println("⚠️  Some packages have build issues that need to be resolved.")
	}
}
