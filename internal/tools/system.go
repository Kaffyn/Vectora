package tools

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
)

type ShellTool struct{}

func (t *ShellTool) Name() string        { return "run_shell_command" }
func (t *ShellTool) Description() string { return "Injeta e executa um script/comando diretamente no Kernel Shell do sistema hospedeiro." }
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

	// Warning de Streaming:
	// Num cenário futuro a 'StdoutPipe' rodará integrada ao daemon.ipc.Broadcast para live-steaming
	// Por hora estamos aguardando a finalização integral via CombinedOutput().
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ToolResult{IsError: true, Output: "Shell Falhou!\nOutput: " + string(out) + "\nMotivo: " + err.Error()}, nil
	}

	return ToolResult{Output: string(out)}, nil
}
