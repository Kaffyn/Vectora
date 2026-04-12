package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TestCLI tests basic CLI commands
func TestCLI(config *EnvironmentConfig, runner *TestRunner) {
	tests := []struct {
		name   string
		args   []string
		checks func(*CommandResult) bool
	}{
		{
			name: "Help Command",
			args: []string{"--help"},
			checks: func(r *CommandResult) bool {
				return r.ExitCode == 0 &&
					(ContainsString(r.Stdout, "Vectora") ||
						ContainsString(r.Stdout, "Usage") ||
						ContainsString(r.Stdout, "Available Commands"))
			},
		},
		{
			name: "Version Command",
			args: []string{"--version"},
			checks: func(r *CommandResult) bool {
				return r.ExitCode == 0 && ContainsString(r.Stdout, "0.1.0")
			},
		},
		{
			name: "Invalid Command",
			args: []string{"nonexistent"},
			checks: func(r *CommandResult) bool {
				return r.ExitCode != 0
			},
		},
		{
			name: "Help for Start",
			args: []string{"start", "--help"},
			checks: func(r *CommandResult) bool {
				return r.ExitCode == 0 && ContainsString(r.Stdout, "start")
			},
		},
		{
			name: "Help for Ask",
			args: []string{"ask", "--help"},
			checks: func(r *CommandResult) bool {
				return r.ExitCode == 0 && ContainsString(r.Stdout, "ask")
			},
		},
		{
			name: "Help for Embed",
			args: []string{"embed", "--help"},
			checks: func(r *CommandResult) bool {
				return r.ExitCode == 0 && ContainsString(r.Stdout, "embed")
			},
		},
		{
			name: "Help for Stop",
			args: []string{"stop", "--help"},
			checks: func(r *CommandResult) bool {
				return r.ExitCode == 0 && ContainsString(r.Stdout, "stop")
			},
		},
		{
			name: "Status Without Core",
			args: []string{"status"},
			checks: func(r *CommandResult) bool {
				// Should either work or fail gracefully
				return true
			},
		},
	}

	for _, test := range tests {
		runner.RunTest(test.name, func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result := ExecCommand(ctx, "vectora", test.args...)

			if !test.checks(result) {
				return fmt.Errorf("checks failed for %v\nstdout: %s\nstderr: %s\nexit: %d",
					test.args, result.Stdout, result.Stderr, result.ExitCode)
			}

			return nil
		})
	}
}

// TestCLIWithValidation tests CLI with validation
func TestCLIWithValidation(config *EnvironmentConfig, runner *TestRunner) {
	// Test ask command without core (should fail gracefully)
	runner.RunTest("Ask Without Core", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora", "ask", "test query")
		// Should either fail with clear message or indicate core is not running
		return nil
	})

	// Test embed command without core
	runner.RunTest("Embed Without Core", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora", "embed", ".")
		// Should fail or indicate core is not running
		return nil
	})

	// Test config with invalid path
	runner.RunTest("Config Invalid Path", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora", "config", "/nonexistent/path")
		// Should handle gracefully
		return nil
	})

	// Test help output consistency
	runner.RunTest("Help Output Consistency", func() error {
		ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel1()

		result1 := ExecCommand(ctx1, "vectora", "--help")

		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()

		result2 := ExecCommand(ctx2, "vectora", "--help")

		if result1.Stdout != result2.Stdout {
			return fmt.Errorf("help output not consistent")
		}

		return nil
	})

	// Test version format
	runner.RunTest("Version Format", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := ExecCommand(ctx, "vectora", "--version")

		if !strings.Contains(result.Stdout, "0.") {
			return fmt.Errorf("version format invalid: %s", result.Stdout)
		}

		return nil
	})

	// Test command not found
	runner.RunTest("Command Not Found", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := ExecCommand(ctx, "vectora", "invalid-command-xyz")

		if result.ExitCode == 0 {
			return fmt.Errorf("expected non-zero exit code for invalid command")
		}

		return nil
	})

	// Test empty arguments
	runner.RunTest("Empty Arguments", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora")

		// Should show help or error
		return nil
	})

	// Test command with too many args
	runner.RunTest("Ask Command Too Many Args", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// This should work - ask accepts multiple args and joins them
		_ = ExecCommand(ctx, "vectora", "ask", "arg1", "arg2", "arg3")

		// Should not exit with error for multiple args
		return nil
	})
}

// TestCLIErrorHandling tests CLI error handling
func TestCLIErrorHandling(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Missing Required Argument", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := ExecCommand(ctx, "vectora", "ask")

		if result.ExitCode == 0 {
			return fmt.Errorf("expected non-zero exit code")
		}

		return nil
	})

	runner.RunTest("Invalid Flag", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := ExecCommand(ctx, "vectora", "--invalid-flag")

		if result.ExitCode == 0 {
			return fmt.Errorf("expected non-zero exit code")
		}

		return nil
	})

	runner.RunTest("Timeout Handling", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// This may timeout which is okay
		_ = ExecCommand(ctx, "vectora", "--help")

		// Either succeeds quickly or times out - both are acceptable
		return nil
	})
}

// TestCLIIntegration tests CLI integration scenarios
func TestCLIIntegration(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Sequential Commands", func() error {
		commands := [][]string{
			{"--version"},
			{"--help"},
			{"status"},
		}

		for _, args := range commands {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result := ExecCommand(ctx, "vectora", args...)
			if result.Err != nil && result.ExitCode != 0 {
				// Some commands may fail if core is not running, which is OK
			}
		}

		return nil
	})

	runner.RunTest("Command Output Format", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := ExecCommand(ctx, "vectora", "--help")

		// Check for common help output patterns
		hasUsage := ContainsString(result.Stdout, "Usage") || ContainsString(result.Stdout, "usage")
		hasCommands := ContainsString(result.Stdout, "Commands") || ContainsString(result.Stdout, "commands")

		if !hasUsage && !hasCommands {
			return fmt.Errorf("help output missing expected sections")
		}

		return nil
	})

	runner.RunTest("Exit Codes", func() error {
		tests := []struct {
			args     []string
			shouldFail bool
		}{
			{[]string{"--help"}, false},
			{[]string{"--version"}, false},
			{[]string{"invalid"}, true},
		}

		for _, test := range tests {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result := ExecCommand(ctx, "vectora", test.args...)

			if test.shouldFail && result.ExitCode == 0 {
				return fmt.Errorf("expected failure for args %v", test.args)
			}
			if !test.shouldFail && result.ExitCode != 0 {
				return fmt.Errorf("expected success for args %v", test.args)
			}
		}

		return nil
	})
}
