package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"time"
)

const CMD_TIMEOUT = 30 * time.Second

type TerminalRunTool struct {
	TrustFolder string
}

func (t *TerminalRunTool) Name() string        { return "terminal_run" }
func (t *TerminalRunTool) Description() string { return "Executa um comando shell no diretório do projeto." }
func (t *TerminalRunTool) Schema() string {
	return `{"type": "object", "properties": {"command": {"type": "string"}}, "required": ["command"]}`
}

func (t *TerminalRunTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct{ Command string `json:"command"` }
	json.Unmarshal(args, &params)

	cmdCtx, cancel := context.WithTimeout(ctx, CMD_TIMEOUT)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cmdCtx, "cmd", "/C", params.Command)
	} else {
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", params.Command)
	}
	cmd.Dir = t.TrustFolder

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
