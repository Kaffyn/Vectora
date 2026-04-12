package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"time"

	"github.com/Kaffyn/Vectora/core/infra"
)

const CMD_TIMEOUT = 30 * time.Second

type ShellTool struct {
	TrustFolder string
}

func (t *ShellTool) Name() string        { return "run_shell_command" }
func (t *ShellTool) Description() string { return "Executes a shell command on the host system." }
func (t *ShellTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`)
}

func (t *ShellTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	cmdCtx, cancel := context.WithTimeout(ctx, CMD_TIMEOUT)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cmdCtx, "powershell", "-c", params.Command)
	} else {
		cmd = exec.CommandContext(cmdCtx, "bash", "-c", params.Command)
	}
	cmd.Dir = t.TrustFolder
	infra.HideWindow(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return &ToolResult{Output: "Command timed out", IsError: true}, nil
		}
		output += "\nError: " + stderr.String()
	}

	return &ToolResult{Output: output, IsError: err != nil}, nil
}
