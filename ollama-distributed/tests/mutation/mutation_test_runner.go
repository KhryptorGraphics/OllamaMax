package mutation

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// MutationTestRunner orchestrates mutation testing for code quality validation
type MutationTestRunner struct {
	ProjectRoot   string
	TestTimeout   time.Duration
	Verbose       bool
	ExcludeFiles  []string
	ExcludeDirs   []string
	TargetDirs    []string
	TestCommand   string
	Results       *MutationResults
}

// MutationResults holds the results of mutation testing
type MutationResults struct {
	TotalMutants     int
	KilledMutants    int
	SurvivedMutants  int
	TimedOutMutants  int
	ErrorMutants     int
	MutationScore    float64
	CoverageScore    float64
	QualityGrade     string
	Mutations        []MutationResult
	ExecutionTime    time.Duration
	TestedPackages   []string
}

// MutationResult represents the result of a single mutation
type MutationResult struct {
	ID           int
	File         string
	Line         int
	Column       int
	Type         string
	Original     string
	Mutant       string
	Status       string
	TestOutput   string
	ExecutionTime time.Duration
	KilledBy     []string
}

// MutationType represents different types of mutations
type MutationType struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
	Description string
}

// NewMutationTestRunner creates a new mutation test runner
func NewMutationTestRunner(projectRoot string) *MutationTestRunner {
	return &MutationTestRunner{
		ProjectRoot: projectRoot,
		TestTimeout: 30 * time.Second,
		Verbose:     false,
		ExcludeFiles: []string{
			"*_test.go",
			"*.pb.go",
			"mock_*.go",
			"*_mock.go",
		},
		ExcludeDirs: []string{
			"vendor",
			".git",
			"node_modules",
			"testdata",
			"test-artifacts",
		},
		TargetDirs: []string{
			"pkg",
			"internal",
			"cmd",
		},
		TestCommand: "go test -race -timeout=30s",
		Results: &MutationResults{
			Mutations: make([]MutationResult, 0),
		},
	}
}

// RunMutationTests executes comprehensive mutation testing
func (mtr *MutationTestRunner) RunMutationTests() error {
	startTime := time.Now()
	
	log.Printf("🧬 Starting Mutation Testing for Code Quality Validation")
	log.Printf("📁 Project Root: %s", mtr.ProjectRoot)
	log.Printf("⏱️  Test Timeout: %v", mtr.TestTimeout)
	
	// Step 1: Baseline test to ensure all tests pass
	log.Printf("🧪 Running baseline tests...")
	if err := mtr.runBaselineTests(); err != nil {
		return fmt.Errorf("baseline tests failed: %w", err)
	}
	
	// Step 2: Find target files for mutation
	log.Printf("🔍 Finding target files for mutation...")
	targetFiles, err := mtr.findTargetFiles()
	if err != nil {
		return fmt.Errorf("failed to find target files: %w", err)
	}
	
	log.Printf("📄 Found %d target files for mutation", len(targetFiles))
	
	// Step 3: Generate mutations
	log.Printf("🧬 Generating mutations...")
	mutations, err := mtr.generateMutations(targetFiles)
	if err != nil {
		return fmt.Errorf("failed to generate mutations: %w", err)
	}
	
	log.Printf("🎯 Generated %d mutations", len(mutations))
	mtr.Results.TotalMutants = len(mutations)
	
	// Step 4: Execute mutations
	log.Printf("🚀 Executing mutation tests...")
	if err := mtr.executeMutations(mutations); err != nil {
		return fmt.Errorf("failed to execute mutations: %w", err)
	}
	
	// Step 5: Calculate results
	mtr.calculateResults()
	mtr.Results.ExecutionTime = time.Since(startTime)
	
	// Step 6: Generate report
	log.Printf("📊 Generating mutation test report...")
	if err := mtr.generateReport(); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	
	log.Printf("✅ Mutation testing completed in %v", mtr.Results.ExecutionTime)
	log.Printf("🎯 Mutation Score: %.2f%%", mtr.Results.MutationScore)
	log.Printf("📝 Quality Grade: %s", mtr.Results.QualityGrade)
	
	return nil
}

// runBaselineTests ensures all tests pass before mutation
func (mtr *MutationTestRunner) runBaselineTests() error {
	cmd := exec.Command("sh", "-c", mtr.TestCommand+" ./...")
	cmd.Dir = mtr.ProjectRoot
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("baseline tests failed:\n%s", string(output))
	}
	
	if mtr.Verbose {
		log.Printf("✅ Baseline tests passed")
	}
	
	return nil
}

// findTargetFiles finds all Go files that should be mutated
func (mtr *MutationTestRunner) findTargetFiles() ([]string, error) {
	var targetFiles []string
	
	for _, targetDir := range mtr.TargetDirs {
		dirPath := filepath.Join(mtr.ProjectRoot, targetDir)
		
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			// Skip directories
			if info.IsDir() {
				// Check if directory should be excluded
				for _, excludeDir := range mtr.ExcludeDirs {
					if strings.Contains(path, excludeDir) {
						return filepath.SkipDir
					}
				}
				return nil
			}
			
			// Check if file should be included
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			
			// Check if file should be excluded
			for _, excludeFile := range mtr.ExcludeFiles {
				if matched, _ := filepath.Match(excludeFile, filepath.Base(path)); matched {
					return nil
				}
			}
			
			targetFiles = append(targetFiles, path)
			return nil
		})
		
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	
	return targetFiles, nil
}

// generateMutations creates mutations for the target files
func (mtr *MutationTestRunner) generateMutations(targetFiles []string) ([]MutationResult, error) {
	var mutations []MutationResult
	mutationID := 1
	
	mutationTypes := mtr.getMutationTypes()
	
	for _, file := range targetFiles {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}
		
		lines := strings.Split(string(content), "\n")
		
		for lineNum, line := range lines {
			for _, mutationType := range mutationTypes {
				matches := mutationType.Pattern.FindAllStringSubmatchIndex(line, -1)
				
				for _, match := range matches {
					if len(match) >= 2 {
						original := line[match[0]:match[1]]
						mutant := mutationType.Replacement
						
						// Skip if replacement would be the same as original
						if original == mutant {
							continue
						}
						
						mutations = append(mutations, MutationResult{
							ID:       mutationID,
							File:     file,
							Line:     lineNum + 1,
							Column:   match[0] + 1,
							Type:     mutationType.Name,
							Original: original,
							Mutant:   mutant,
							Status:   "pending",
						})
						
						mutationID++
					}
				}
			}
		}
	}
	
	return mutations, nil
}

// getMutationTypes returns the mutation types to apply
func (mtr *MutationTestRunner) getMutationTypes() []MutationType {
	return []MutationType{
		// Arithmetic Operators
		{
			Name:        "ArithmeticReplacement",
			Pattern:     regexp.MustCompile(`\+`),
			Replacement: "-",
			Description: "Replace + with -",
		},
		{
			Name:        "ArithmeticReplacement",
			Pattern:     regexp.MustCompile(`\-`),
			Replacement: "+",
			Description: "Replace - with +",
		},
		{
			Name:        "ArithmeticReplacement",
			Pattern:     regexp.MustCompile(`\*`),
			Replacement: "/",
			Description: "Replace * with /",
		},
		{
			Name:        "ArithmeticReplacement",
			Pattern:     regexp.MustCompile(`/`),
			Replacement: "*",
			Description: "Replace / with *",
		},
		
		// Comparison Operators
		{
			Name:        "ComparisonReplacement",
			Pattern:     regexp.MustCompile(`==`),
			Replacement: "!=",
			Description: "Replace == with !=",
		},
		{
			Name:        "ComparisonReplacement",
			Pattern:     regexp.MustCompile(`!=`),
			Replacement: "==",
			Description: "Replace != with ==",
		},
		{
			Name:        "ComparisonReplacement",
			Pattern:     regexp.MustCompile(`<`),
			Replacement: ">=",
			Description: "Replace < with >=",
		},
		{
			Name:        "ComparisonReplacement",
			Pattern:     regexp.MustCompile(`>`),
			Replacement: "<=",
			Description: "Replace > with <=",
		},
		{
			Name:        "ComparisonReplacement",
			Pattern:     regexp.MustCompile(`<=`),
			Replacement: ">",
			Description: "Replace <= with >",
		},
		{
			Name:        "ComparisonReplacement",
			Pattern:     regexp.MustCompile(`>=`),
			Replacement: "<",
			Description: "Replace >= with <",
		},
		
		// Logical Operators
		{
			Name:        "LogicalReplacement",
			Pattern:     regexp.MustCompile(`&&`),
			Replacement: "||",
			Description: "Replace && with ||",
		},
		{
			Name:        "LogicalReplacement",
			Pattern:     regexp.MustCompile(`\|\|`),
			Replacement: "&&",
			Description: "Replace || with &&",
		},
		
		// Boolean Constants
		{
			Name:        "BooleanReplacement",
			Pattern:     regexp.MustCompile(`\btrue\b`),
			Replacement: "false",
			Description: "Replace true with false",
		},
		{
			Name:        "BooleanReplacement",
			Pattern:     regexp.MustCompile(`\bfalse\b`),
			Replacement: "true",
			Description: "Replace false with true",
		},
		
		// Numeric Constants
		{
			Name:        "NumericReplacement",
			Pattern:     regexp.MustCompile(`\b0\b`),
			Replacement: "1",
			Description: "Replace 0 with 1",
		},
		{
			Name:        "NumericReplacement",
			Pattern:     regexp.MustCompile(`\b1\b`),
			Replacement: "0",
			Description: "Replace 1 with 0",
		},
		
		// Increment/Decrement
		{
			Name:        "IncrementReplacement",
			Pattern:     regexp.MustCompile(`\+\+`),
			Replacement: "--",
			Description: "Replace ++ with --",
		},
		{
			Name:        "DecrementReplacement",
			Pattern:     regexp.MustCompile(`\-\-`),
			Replacement: "++",
			Description: "Replace -- with ++",
		},
	}
}

// executeMutations runs tests for each mutation
func (mtr *MutationTestRunner) executeMutations(mutations []MutationResult) error {
	for i, mutation := range mutations {
		if mtr.Verbose {
			log.Printf("🧬 Testing mutation %d/%d: %s in %s:%d", 
				i+1, len(mutations), mutation.Type, mutation.File, mutation.Line)
		}
		
		result, err := mtr.executeSingleMutation(mutation)
		if err != nil {
			log.Printf("⚠️ Error executing mutation %d: %v", mutation.ID, err)
			result.Status = "error"
		}
		
		mtr.Results.Mutations = append(mtr.Results.Mutations, result)
		
		// Update counters
		switch result.Status {
		case "killed":
			mtr.Results.KilledMutants++
		case "survived":
			mtr.Results.SurvivedMutants++
		case "timeout":
			mtr.Results.TimedOutMutants++
		case "error":
			mtr.Results.ErrorMutants++
		}
	}
	
	return nil
}

// executeSingleMutation tests a single mutation
func (mtr *MutationTestRunner) executeSingleMutation(mutation MutationResult) (MutationResult, error) {
	startTime := time.Now()
	
	// Read original file
	originalContent, err := ioutil.ReadFile(mutation.File)
	if err != nil {
		return mutation, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Create mutated content
	lines := strings.Split(string(originalContent), "\n")
	if mutation.Line-1 >= len(lines) {
		return mutation, fmt.Errorf("line number out of range")
	}
	
	originalLine := lines[mutation.Line-1]
	mutatedLine := strings.Replace(originalLine, mutation.Original, mutation.Mutant, 1)
	lines[mutation.Line-1] = mutatedLine
	
	mutatedContent := strings.Join(lines, "\n")
	
	// Write mutated file
	if err := ioutil.WriteFile(mutation.File, []byte(mutatedContent), 0644); err != nil {
		return mutation, fmt.Errorf("failed to write mutated file: %w", err)
	}
	
	// Ensure original content is restored
	defer func() {
		ioutil.WriteFile(mutation.File, originalContent, 0644)
	}()
	
	// Run tests with timeout
	ctx, cancel := context.WithTimeout(context.Background(), mtr.TestTimeout)
	defer cancel()
	
	// Determine which package to test
	relPath, _ := filepath.Rel(mtr.ProjectRoot, mutation.File)
	packagePath := "./" + filepath.Dir(relPath)
	
	cmd := exec.CommandContext(ctx, "sh", "-c", mtr.TestCommand+" "+packagePath)
	cmd.Dir = mtr.ProjectRoot
	
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	
	err = cmd.Run()
	
	mutation.ExecutionTime = time.Since(startTime)
	mutation.TestOutput = output.String()
	
	// Determine mutation status
	if ctx.Err() == context.DeadlineExceeded {
		mutation.Status = "timeout"
	} else if err != nil {
		mutation.Status = "killed"
		mutation.KilledBy = mtr.extractFailedTests(mutation.TestOutput)
	} else {
		mutation.Status = "survived"
	}
	
	return mutation, nil
}

// extractFailedTests extracts which tests failed from test output
func (mtr *MutationTestRunner) extractFailedTests(output string) []string {
	var failedTests []string
	
	// Parse Go test output for failed tests
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "FAIL:") || strings.Contains(line, "--- FAIL:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				failedTests = append(failedTests, parts[2])
			}
		}
	}
	
	return failedTests
}

// calculateResults computes final mutation testing results
func (mtr *MutationTestRunner) calculateResults() {
	if mtr.Results.TotalMutants > 0 {
		mtr.Results.MutationScore = float64(mtr.Results.KilledMutants) / float64(mtr.Results.TotalMutants) * 100
	}
	
	// Calculate quality grade based on mutation score
	switch {
	case mtr.Results.MutationScore >= 80:
		mtr.Results.QualityGrade = "A (Excellent)"
	case mtr.Results.MutationScore >= 70:
		mtr.Results.QualityGrade = "B (Good)"
	case mtr.Results.MutationScore >= 60:
		mtr.Results.QualityGrade = "C (Fair)"
	case mtr.Results.MutationScore >= 50:
		mtr.Results.QualityGrade = "D (Poor)"
	default:
		mtr.Results.QualityGrade = "F (Fail)"
	}
}

// generateReport creates a comprehensive mutation testing report
func (mtr *MutationTestRunner) generateReport() error {
	reportPath := filepath.Join(mtr.ProjectRoot, "test-artifacts", 
		fmt.Sprintf("mutation_test_report_%s.txt", time.Now().Format("20060102_150405")))
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}
	
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	
	// Write report header
	fmt.Fprintf(writer, "🧬 MUTATION TESTING REPORT\n")
	fmt.Fprintf(writer, "==========================\n\n")
	fmt.Fprintf(writer, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(writer, "Project: %s\n", mtr.ProjectRoot)
	fmt.Fprintf(writer, "Execution Time: %v\n\n", mtr.Results.ExecutionTime)
	
	// Write summary
	fmt.Fprintf(writer, "📊 SUMMARY\n")
	fmt.Fprintf(writer, "----------\n")
	fmt.Fprintf(writer, "Total Mutants: %d\n", mtr.Results.TotalMutants)
	fmt.Fprintf(writer, "Killed Mutants: %d\n", mtr.Results.KilledMutants)
	fmt.Fprintf(writer, "Survived Mutants: %d\n", mtr.Results.SurvivedMutants)
	fmt.Fprintf(writer, "Timed Out Mutants: %d\n", mtr.Results.TimedOutMutants)
	fmt.Fprintf(writer, "Error Mutants: %d\n", mtr.Results.ErrorMutants)
	fmt.Fprintf(writer, "Mutation Score: %.2f%%\n", mtr.Results.MutationScore)
	fmt.Fprintf(writer, "Quality Grade: %s\n\n", mtr.Results.QualityGrade)
	
	// Write survived mutants (these indicate test gaps)
	if mtr.Results.SurvivedMutants > 0 {
		fmt.Fprintf(writer, "🚨 SURVIVED MUTANTS (Test Quality Issues)\n")
		fmt.Fprintf(writer, "----------------------------------------\n")
		
		for _, mutation := range mtr.Results.Mutations {
			if mutation.Status == "survived" {
				fmt.Fprintf(writer, "ID: %d | %s:%d:%d | %s | %s → %s\n",
					mutation.ID, mutation.File, mutation.Line, mutation.Column,
					mutation.Type, mutation.Original, mutation.Mutant)
			}
		}
		fmt.Fprintf(writer, "\n")
	}
	
	// Write mutation type analysis
	fmt.Fprintf(writer, "📈 MUTATION TYPE ANALYSIS\n")
	fmt.Fprintf(writer, "-------------------------\n")
	typeStats := mtr.analyzeMutationTypes()
	for mutationType, stats := range typeStats {
		fmt.Fprintf(writer, "%s: %d total, %d killed (%.1f%%)\n",
			mutationType, stats.Total, stats.Killed, 
			float64(stats.Killed)/float64(stats.Total)*100)
	}
	fmt.Fprintf(writer, "\n")
	
	// Write recommendations
	fmt.Fprintf(writer, "💡 RECOMMENDATIONS\n")
	fmt.Fprintf(writer, "------------------\n")
	mtr.writeRecommendations(writer)
	
	log.Printf("📄 Mutation test report saved to: %s", reportPath)
	return nil
}

// MutationTypeStats holds statistics for a mutation type
type MutationTypeStats struct {
	Total   int
	Killed  int
	Survived int
}

// analyzeMutationTypes analyzes results by mutation type
func (mtr *MutationTestRunner) analyzeMutationTypes() map[string]MutationTypeStats {
	typeStats := make(map[string]MutationTypeStats)
	
	for _, mutation := range mtr.Results.Mutations {
		stats := typeStats[mutation.Type]
		stats.Total++
		
		if mutation.Status == "killed" {
			stats.Killed++
		} else if mutation.Status == "survived" {
			stats.Survived++
		}
		
		typeStats[mutation.Type] = stats
	}
	
	return typeStats
}

// writeRecommendations writes improvement recommendations
func (mtr *MutationTestRunner) writeRecommendations(writer *bufio.Writer) {
	fmt.Fprintf(writer, "Based on the mutation testing results:\n\n")
	
	if mtr.Results.MutationScore < 70 {
		fmt.Fprintf(writer, "🔴 LOW MUTATION SCORE (%.1f%%)\n", mtr.Results.MutationScore)
		fmt.Fprintf(writer, "- Add more edge case tests\n")
		fmt.Fprintf(writer, "- Test boundary conditions\n")
		fmt.Fprintf(writer, "- Improve assertion coverage\n\n")
	}
	
	if mtr.Results.SurvivedMutants > 0 {
		fmt.Fprintf(writer, "🔴 SURVIVED MUTANTS (%d)\n", mtr.Results.SurvivedMutants)
		fmt.Fprintf(writer, "- Review survived mutants above\n")
		fmt.Fprintf(writer, "- Add tests for uncovered logic\n")
		fmt.Fprintf(writer, "- Strengthen existing assertions\n\n")
	}
	
	if mtr.Results.TimedOutMutants > 5 {
		fmt.Fprintf(writer, "⚠️ MANY TIMEOUTS (%d)\n", mtr.Results.TimedOutMutants)
		fmt.Fprintf(writer, "- Some mutations may cause infinite loops\n")
		fmt.Fprintf(writer, "- Consider increasing test timeout\n")
		fmt.Fprintf(writer, "- Review loop termination conditions\n\n")
	}
	
	fmt.Fprintf(writer, "✅ GENERAL RECOMMENDATIONS:\n")
	fmt.Fprintf(writer, "- Aim for mutation score > 80%%\n")
	fmt.Fprintf(writer, "- Focus on testing business logic\n")
	fmt.Fprintf(writer, "- Use property-based testing for algorithms\n")
	fmt.Fprintf(writer, "- Regular mutation testing in CI/CD\n")
}

// RunMutationTestsForPackage runs mutation tests for a specific package
func (mtr *MutationTestRunner) RunMutationTestsForPackage(packagePath string) error {
	mtr.TargetDirs = []string{packagePath}
	return mtr.RunMutationTests()
}

// GetMutationScore returns the current mutation score
func (mtr *MutationTestRunner) GetMutationScore() float64 {
	return mtr.Results.MutationScore
}

// GetQualityGrade returns the quality grade based on mutation score
func (mtr *MutationTestRunner) GetQualityGrade() string {
	return mtr.Results.QualityGrade
}