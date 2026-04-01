package tools

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
)

type ShellTool struct{}

func (t *ShellTool) Name() string        { return "run_shell_command" }
func (t *ShellTool) Description() string { return "Injects and executes a script/command directly into the host system's Kernel Shell." }
func (t *ShellTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`)
}
func (t *ShellTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	cmdStr, _ := args["command"].(string)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-c", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", cmdStr)
	}

	// Streaming Warning:
	// In the future, 'StdoutPipe' will run integrated with daemon.ipc.Broadcast for live-streaming.
	// For now, we are awaiting full completion via CombinedOutput().
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ToolResult{IsError: true, Output: "Shell Failed!\nOutput: " + string(out) + "\nReason: " + err.Error()}, nil
	}

	return ToolResult{Output: string(out)}, nil
}
