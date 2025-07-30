package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/tests/mutation"
)

var (
	projectRoot = flag.String("root", ".", "Project root directory")
	packagePath = flag.String("package", "", "Specific package to test (e.g., pkg/consensus)")
	timeout     = flag.Duration("timeout", 30*time.Second, "Test timeout duration")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
	threshold   = flag.Float64("threshold", 70.0, "Minimum mutation score threshold")
	excludeDirs = flag.String("exclude-dirs", "vendor,.git,node_modules,testdata", "Comma-separated list of directories to exclude")
	excludeFiles = flag.String("exclude-files", "*_test.go,*.pb.go,mock_*.go", "Comma-separated list of file patterns to exclude")
	testCmd     = flag.String("test-cmd", "go test -race -timeout=30s", "Test command to execute")
	outputDir   = flag.String("output", "test-artifacts", "Output directory for reports")
	workers     = flag.Int("workers", 4, "Number of parallel workers for mutation testing")
	quick       = flag.Bool("quick", false, "Run quick mutation testing (fewer mutations)")
	reportFormat = flag.String("format", "text", "Report format: text, json, html")
)

func main() {
	flag.Parse()

	if *projectRoot == "" {
		log.Fatal("Project root directory is required")
	}

	// Resolve absolute path
	absRoot, err := filepath.Abs(*projectRoot)
	if err != nil {
		log.Fatalf("Failed to resolve project root: %v", err)
	}

	fmt.Printf("🧬 Ollama Distributed System - Mutation Testing Tool\n")
	fmt.Printf("===================================================\n\n")
	fmt.Printf("Project Root: %s\n", absRoot)
	fmt.Printf("Mutation Score Threshold: %.1f%%\n", *threshold)
	fmt.Printf("Test Timeout: %v\n", *timeout)
	fmt.Printf("Workers: %d\n", *workers)
	fmt.Printf("Quick Mode: %v\n\n", *quick)

	// Create mutation test runner
	runner := mutation.NewMutationTestRunner(absRoot)
	runner.TestTimeout = *timeout
	runner.Verbose = *verbose
	runner.TestCommand = *testCmd

	// Configure exclusions
	if *excludeDirs != "" {
		runner.ExcludeDirs = strings.Split(*excludeDirs, ",")
	}
	if *excludeFiles != "" {
		runner.ExcludeFiles = strings.Split(*excludeFiles, ",")
	}

	// Configure output directory
	outputPath := filepath.Join(absRoot, *outputDir)
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	var mutationErr error
	if *packagePath != "" {
		// Test specific package
		fmt.Printf("🎯 Testing package: %s\n\n", *packagePath)
		mutationErr = runner.RunMutationTestsForPackage(*packagePath)
	} else {
		// Test entire project
		fmt.Printf("🚀 Testing entire project\n\n")
		if *quick {
			mutationErr = runQuickMutationTesting(runner)
		} else {
			mutationErr = runner.RunMutationTests()
		}
	}

	if mutationErr != nil {
		log.Fatalf("Mutation testing failed: %v", mutationErr)
	}

	// Display results
	displayResults(runner)

	// Generate additional reports
	if err := generateAdditionalReports(runner, outputPath); err != nil {
		log.Printf("Warning: Failed to generate additional reports: %v", err)
	}

	// Check threshold
	if runner.GetMutationScore() < *threshold {
		fmt.Printf("\n❌ Mutation score %.2f%% is below threshold %.1f%%\n", 
			runner.GetMutationScore(), *threshold)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Mutation testing completed successfully!\n")
	fmt.Printf("🎯 Mutation Score: %.2f%% (Grade: %s)\n", 
		runner.GetMutationScore(), runner.GetQualityGrade())
}

func runQuickMutationTesting(runner *mutation.MutationTestRunner) error {
	fmt.Printf("⚡ Quick Mode: Testing critical packages only\n")
	
	criticalPackages := []string{
		"pkg/consensus",
		"pkg/p2p", 
		"internal/auth",
		"pkg/api",
	}

	var allErrors []string
	totalMutants := 0
	totalKilled := 0

	for _, pkg := range criticalPackages {
		fmt.Printf("🧬 Testing %s...\n", pkg)
		
		pkgRunner := mutation.NewMutationTestRunner(runner.ProjectRoot)
		pkgRunner.TestTimeout = runner.TestTimeout
		pkgRunner.Verbose = runner.Verbose
		pkgRunner.TestCommand = runner.TestCommand
		pkgRunner.ExcludeDirs = runner.ExcludeDirs
		pkgRunner.ExcludeFiles = runner.ExcludeFiles

		err := pkgRunner.RunMutationTestsForPackage(pkg)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("%s: %v", pkg, err))
			continue
		}

		totalMutants += pkgRunner.Results.TotalMutants
		totalKilled += pkgRunner.Results.KilledMutants

		fmt.Printf("  📊 %s: %.1f%% (%d/%d mutants killed)\n", 
			pkg, pkgRunner.GetMutationScore(), 
			pkgRunner.Results.KilledMutants, pkgRunner.Results.TotalMutants)
	}

	// Aggregate results
	if totalMutants > 0 {
		runner.Results.TotalMutants = totalMutants
		runner.Results.KilledMutants = totalKilled
		runner.Results.SurvivedMutants = totalMutants - totalKilled
		runner.Results.MutationScore = float64(totalKilled) / float64(totalMutants) * 100
	}

	if len(allErrors) > 0 {
		fmt.Printf("\n⚠️  Some packages had errors:\n")
		for _, err := range allErrors {
			fmt.Printf("  - %s\n", err)
		}
	}

	return nil
}

func displayResults(runner *mutation.MutationTestRunner) {
	results := runner.Results
	
	fmt.Printf("\n📊 MUTATION TESTING RESULTS\n")
	fmt.Printf("============================\n")
	fmt.Printf("Total Mutants: %d\n", results.TotalMutants)
	fmt.Printf("Killed Mutants: %d\n", results.KilledMutants)
	fmt.Printf("Survived Mutants: %d\n", results.SurvivedMutants)
	fmt.Printf("Timed Out Mutants: %d\n", results.TimedOutMutants)
	fmt.Printf("Error Mutants: %d\n", results.ErrorMutants)
	fmt.Printf("Execution Time: %v\n", results.ExecutionTime)
	fmt.Printf("Mutation Score: %.2f%%\n", results.MutationScore)
	fmt.Printf("Quality Grade: %s\n", results.QualityGrade)

	if results.SurvivedMutants > 0 {
		fmt.Printf("\n🚨 QUALITY ISSUES DETECTED\n")
		fmt.Printf("==========================\n")
		fmt.Printf("Survived mutants indicate potential test quality issues.\n")
		fmt.Printf("Consider adding more comprehensive tests for:\n")

		// Group survived mutants by file
		fileMap := make(map[string]int)
		for _, mutation := range results.Mutations {
			if mutation.Status == "survived" {
				fileMap[mutation.File]++
			}
		}

		for file, count := range fileMap {
			relPath, _ := filepath.Rel(runner.ProjectRoot, file)
			fmt.Printf("  - %s (%d survived mutants)\n", relPath, count)
		}
	}

	if results.MutationScore < 60 {
		fmt.Printf("\n💡 IMPROVEMENT RECOMMENDATIONS\n")
		fmt.Printf("==============================\n")
		fmt.Printf("Low mutation score indicates test quality issues:\n")
		fmt.Printf("  • Add edge case testing\n")
		fmt.Printf("  • Test boundary conditions\n")
		fmt.Printf("  • Improve assertion coverage\n")
		fmt.Printf("  • Use property-based testing\n")
		fmt.Printf("  • Review survived mutants\n")
	}
}

func generateAdditionalReports(runner *mutation.MutationTestRunner, outputDir string) error {
	switch *reportFormat {
	case "json":
		return generateJSONReport(runner, outputDir)
	case "html":
		return generateHTMLReport(runner, outputDir)
	case "text":
		// Default text report is already generated
		return nil
	default:
		return fmt.Errorf("unsupported report format: %s", *reportFormat)
	}
}

func generateJSONReport(runner *mutation.MutationTestRunner, outputDir string) error {
	// TODO: Implement JSON report generation
	fmt.Printf("📄 JSON report generation not yet implemented\n")
	return nil
}

func generateHTMLReport(runner *mutation.MutationTestRunner, outputDir string) error {
	htmlContent := generateHTMLReportContent(runner)
	
	reportPath := filepath.Join(outputDir, 
		fmt.Sprintf("mutation_report_%s.html", time.Now().Format("20060102_150405")))
	
	err := os.WriteFile(reportPath, []byte(htmlContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}
	
	fmt.Printf("📄 HTML report generated: %s\n", reportPath)
	return nil
}

func generateHTMLReportContent(runner *mutation.MutationTestRunner) string {
	results := runner.Results
	
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Mutation Testing Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { display: flex; gap: 20px; margin: 20px 0; }
        .metric { background: #e8f4fd; padding: 15px; border-radius: 5px; text-align: center; }
        .metric h3 { margin: 0; color: #1976d2; }
        .metric .value { font-size: 24px; font-weight: bold; }
        .grade-A { background: #c8e6c9; color: #2e7d32; }
        .grade-B { background: #fff3c4; color: #f57c00; }
        .grade-C { background: #ffcccb; color: #d32f2f; }
        .grade-D, .grade-F { background: #ffcccb; color: #d32f2f; }
        .mutations { margin: 20px 0; }
        .mutation { border: 1px solid #ddd; margin: 5px 0; padding: 10px; border-radius: 3px; }
        .survived { background: #ffebee; border-color: #f44336; }
        .killed { background: #e8f5e8; border-color: #4caf50; }
        .timeout { background: #fff3e0; border-color: #ff9800; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🧬 Mutation Testing Report</h1>
        <p><strong>Generated:</strong> ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
        <p><strong>Execution Time:</strong> ` + results.ExecutionTime.String() + `</p>
    </div>

    <div class="summary">
        <div class="metric">
            <h3>Total Mutants</h3>
            <div class="value">` + fmt.Sprintf("%d", results.TotalMutants) + `</div>
        </div>
        <div class="metric">
            <h3>Killed Mutants</h3>
            <div class="value">` + fmt.Sprintf("%d", results.KilledMutants) + `</div>
        </div>
        <div class="metric">
            <h3>Survived Mutants</h3>
            <div class="value">` + fmt.Sprintf("%d", results.SurvivedMutants) + `</div>
        </div>
        <div class="metric">
            <h3>Mutation Score</h3>
            <div class="value">` + fmt.Sprintf("%.1f%%", results.MutationScore) + `</div>
        </div>
        <div class="metric ` + getGradeClass(results.QualityGrade) + `">
            <h3>Quality Grade</h3>
            <div class="value">` + results.QualityGrade + `</div>
        </div>
    </div>`

	if results.SurvivedMutants > 0 {
		html += `
    <div class="mutations">
        <h2>🚨 Survived Mutants (Quality Issues)</h2>`
		
		for _, mutation := range results.Mutations {
			if mutation.Status == "survived" {
				html += fmt.Sprintf(`
        <div class="mutation survived">
            <strong>%s:%d:%d</strong> - %s<br>
            <code>%s</code> → <code>%s</code><br>
            <em>Type: %s</em>
        </div>`, mutation.File, mutation.Line, mutation.Column, mutation.Type,
					mutation.Original, mutation.Mutant, mutation.Type)
			}
		}
		
		html += `
    </div>`
	}

	html += `
</body>
</html>`

	return html
}

func getGradeClass(grade string) string {
	switch {
	case strings.HasPrefix(grade, "A"):
		return "grade-A"
	case strings.HasPrefix(grade, "B"):
		return "grade-B"
	case strings.HasPrefix(grade, "C"):
		return "grade-C"
	case strings.HasPrefix(grade, "D"):
		return "grade-D"
	default:
		return "grade-F"
	}
}