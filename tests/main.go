package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// TestRunner manages test execution
type TestRunner struct {
	tests   []TestCase
	config  *EnvironmentConfig
	results *TestResults
	mu      sync.Mutex
}

// TestCase represents a single test
type TestCase struct {
	Name     string
	Duration time.Duration
	Error    error
	Passed   bool
	ErrorMsg string
}

// TestResults holds aggregated results
type TestResults struct {
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Tests    []TestCase
	StartedAt time.Time
	FinishedAt time.Time
}

// NewTestRunner creates a new test runner
func NewTestRunner(config *EnvironmentConfig) *TestRunner {
	return &TestRunner{
		config:  config,
		results: &TestResults{},
	}
}

// RunTest executes a single test
func (r *TestRunner) RunTest(name string, fn func() error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	testCase := TestCase{
		Name:     name,
		Duration: duration,
		Error:    nil, // Don't serialize the error
		Passed:   err == nil,
		ErrorMsg: errorMsg,
	}

	if testCase.Passed {
		r.results.Passed++
		fmt.Printf("  ✅ %-50s %s\n", name, FormatDuration(duration))
	} else {
		r.results.Failed++
		fmt.Printf("  ❌ %-50s %s\n", name, FormatDuration(duration))
		fmt.Printf("     Error: %v\n", err)
	}

	r.results.Total++
	r.results.Tests = append(r.results.Tests, testCase)
}

// Run executes all test suites
func (r *TestRunner) Run() int {
	r.results.StartedAt = time.Now()

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("VECTORA CLI LOCAL TEST SUITE")
	fmt.Println("========================================")
	fmt.Println()

	// Load configuration
	config := r.config
	if config == nil {
		config = NewEnvironmentConfig()
		r.config = config
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Core Binary: %s\n", config.CoreBinaryPath)
	fmt.Printf("  Workspace: %s\n", config.WorkspaceDir)
	fmt.Printf("  Log Level: %s\n", config.LogLevel)
	fmt.Printf("  Timeout: %ds\n", config.TimeoutSeconds)
	fmt.Println()

	// Run test suites
	fmt.Println("Running CLI Tests...")
	TestCLI(config, r)
	TestCLIWithValidation(config, r)
	TestCLIErrorHandling(config, r)
	TestCLIIntegration(config, r)

	fmt.Println()
	fmt.Println("Running Integration Tests...")
	TestIntegration(config, r)
	TestCoreIntegration(config, r)
	TestCommandIntegration(config, r)
	TestWorkspaceIntegration(config, r)
	TestEndToEndFlow(config, r)
	TestProcessIntegration(config, r)
	TestErrorRecovery(config, r)

	fmt.Println()
	fmt.Println("Running ACP Protocol Tests...")
	TestACP(config, r)
	TestACPMessages(config, r)
	TestACPErrors(config, r)
	TestACPStreaming(config, r)
	TestACPTools(config, r)
	TestACPProtocol(config, r)

	fmt.Println()
	fmt.Println("Running Feature Tests...")
	TestFeatures(config, r)
	TestTokenCounting(config, r)
	TestCostTracking(config, r)
	TestModelSwitching(config, r)
	TestChatFeatures(config, r)
	TestRAGFeatures(config, r)
	TestStreamingFeatures(config, r)
	TestAdvancedFeatures(config, r)

	r.results.FinishedAt = time.Now()
	r.results.Duration = r.results.FinishedAt.Sub(r.results.StartedAt)

	// Print summary
	r.printSummary()

	// Save results
	if err := r.saveResults(); err != nil {
		fmt.Printf("Warning: Failed to save results: %v\n", err)
	}

	// Return exit code
	if r.results.Failed > 0 {
		return 1
	}
	return 0
}

func (r *TestRunner) printSummary() {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("Results: %d/%d PASSED", r.results.Passed, r.results.Total)

	if r.results.Failed > 0 {
		fmt.Printf(" | %d FAILED", r.results.Failed)
	}
	if r.results.Skipped > 0 {
		fmt.Printf(" | %d SKIPPED", r.results.Skipped)
	}
	fmt.Println()

	if r.results.Failed == 0 {
		fmt.Println("Status: ✅ ALL TESTS PASSED")
	} else {
		fmt.Printf("Status: ❌ %d TEST(S) FAILED\n", r.results.Failed)
	}

	fmt.Printf("Total Time: %s\n", FormatDuration(r.results.Duration))
	fmt.Println("========================================")
	fmt.Println()
}

func (r *TestRunner) saveResults() error {
	reportDir := "reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return err
	}

	// Save JSON results
	resultsFile := filepath.Join(reportDir, "results.json")
	data, err := json.MarshalIndent(r.results, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(resultsFile, data, 0644); err != nil {
		return err
	}

	// Save timing report
	timingFile := filepath.Join(reportDir, "timing.txt")
	var timingReport string

	// Sort tests by duration (longest first)
	tests := append([]TestCase{}, r.results.Tests...)
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].Duration > tests[j].Duration
	})

	timingReport = "Test Execution Times\n"
	timingReport += "====================\n\n"

	for i, test := range tests {
		status := "PASS"
		if !test.Passed {
			status = "FAIL"
		}
		timingReport += fmt.Sprintf("%2d. %s [%s] %s\n",
			i+1, test.Name, status, FormatDuration(test.Duration))
	}

	timingReport += fmt.Sprintf("\nTotal: %s\n", FormatDuration(r.results.Duration))

	if err := os.WriteFile(timingFile, []byte(timingReport), 0644); err != nil {
		return err
	}

	// Print report locations
	fmt.Printf("Reports saved to:\n")
	fmt.Printf("  - %s\n", resultsFile)
	fmt.Printf("  - %s\n", timingFile)

	return nil
}

// parseArgs parses command line arguments
func parseArgs() (testFilter string, verbose bool) {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-verbose", "-v":
			verbose = true
		case "-test":
			// Next arg is the test filter
			for i, a := range os.Args {
				if a == "-test" && i+1 < len(os.Args) {
					testFilter = os.Args[i+1]
				}
			}
		}
	}
	return
}

func main() {
	// Parse environment
	testFilter, verbose := parseArgs()

	// Create configuration
	config := NewEnvironmentConfig()

	// Create test runner
	runner := NewTestRunner(config)

	if verbose {
		fmt.Println("Verbose mode enabled")
	}

	if testFilter != "" {
		fmt.Printf("Test filter: %s\n", testFilter)
	}

	// Run tests
	exitCode := runner.Run()

	os.Exit(exitCode)
}
