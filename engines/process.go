package engines

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// ProcessManager gerencia a vida útil de um processo llama.cpp.
type ProcessManager struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr *bufio.Reader
	PID    int
}

// StartProcess inicia um novo processo llama.cpp e retorna um ProcessManager.
//
// Args esperados:
//   - llama_path: caminho absoluto para o binário llama-cli
//   - model_path: caminho para o arquivo GGUF do modelo
//   - ctx_tokens: número de tokens de contexto
//   - n_threads: número de threads (CPU)
//   - gpuLayers: número de layers para GPU (0 = CPU-only)
func StartProcess(
	ctx context.Context,
	llamaPath string,
	modelPath string,
	ctxTokens int,
	nThreads int,
	gpuLayers int,
) (*ProcessManager, error) {
	// Montar comando
	args := []string{
		"--model", modelPath,
		"--ctx-size", fmt.Sprintf("%d", ctxTokens),
		"--threads", fmt.Sprintf("%d", nThreads),
		"--n-gpu-layers", fmt.Sprintf("%d", gpuLayers),
		"--log-disable", // Desabilitar logs padrão do llama.cpp
		"-p", "init", // Prompt inicial
	}

	cmd := exec.CommandContext(ctx, llamaPath, args...)

	// Configurar pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Iniciar processo
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	pm := &ProcessManager{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		stderr: bufio.NewReader(stderr),
		PID:    cmd.Process.Pid,
	}

	// Esperar pela mensagem de inicialização
	// (timeout de 30 segundos)
	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := pm.waitForReady(initCtx); err != nil {
		pm.Terminate(context.Background())
		return nil, fmt.Errorf("process failed to initialize: %w", err)
	}

	return pm, nil
}

// waitForReady aguarda a mensagem de inicialização do llama.cpp.
func (pm *ProcessManager) waitForReady(ctx context.Context) error {
	// Ler stderr até encontrar "init" ou "ready" message
	// (implementação simplificada)

	done := make(chan error, 1)
	go func() {
		for {
			line, err := pm.stderr.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					done <- nil
				} else {
					done <- err
				}
				return
			}

			// Log de debug
			if line != "" {
				fmt.Fprintf(os.Stderr, "[llama] %s", line)
			}

			// Procurar por sinais de pronto
			if contains(line, "init", "ready", "loaded") {
				done <- nil
				return
			}
		}
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Complete envia um prompt e retorna a resposta do llama.cpp.
// Usa JSON para comunicação (newline-delimited).
func (pm *ProcessManager) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	// Construir request JSON
	req := map[string]interface{}{
		"method":     "complete",
		"prompt":     prompt,
		"max_tokens": maxTokens,
	}

	data, _ := json.Marshal(req)

	// Enviar para stdin (newline-delimited)
	if _, err := pm.stdin.Write(append(data, '\n')); err != nil {
		return "", fmt.Errorf("failed to write to llama process: %w", err)
	}

	// Ler resposta JSON
	responseLine, err := pm.stdout.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from llama process: %w", err)
	}

	// Parse resposta
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(responseLine), &resp); err != nil {
		return "", fmt.Errorf("failed to parse llama response: %w", err)
	}

	// Extrair texto
	if text, ok := resp["text"].(string); ok {
		return text, nil
	}

	return "", fmt.Errorf("no text in response: %v", resp)
}

// Embedding calcula embedding para um texto.
func (pm *ProcessManager) Embedding(ctx context.Context, text string) ([]float32, error) {
	req := map[string]interface{}{
		"method": "embedding",
		"text":   text,
	}

	data, _ := json.Marshal(req)

	if _, err := pm.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write to llama process: %w", err)
	}

	responseLine, err := pm.stdout.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read from llama process: %w", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(responseLine), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	// Extrair embedding (array de floats)
	if embData, ok := resp["embedding"].([]interface{}); ok {
		result := make([]float32, len(embData))
		for i, v := range embData {
			if f, ok := v.(float64); ok {
				result[i] = float32(f)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("no embedding in response")
}

// IsRunning verifica se o processo ainda está ativo.
func (pm *ProcessManager) IsRunning() bool {
	if pm.cmd == nil || pm.cmd.ProcessState == nil {
		return false
	}
	return !pm.cmd.ProcessState.Exited()
}

// Terminate tenta terminar o processo gracefully.
func (pm *ProcessManager) Terminate(ctx context.Context) error {
	if pm.cmd == nil || pm.cmd.Process == nil {
		return nil
	}

	// SIGTERM (ou CTRL_BREAK_EVENT no Windows)
	if err := pm.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to signal process: %w", err)
	}

	// Esperar por até 5 segundos
	done := make(chan error, 1)
	go func() {
		done <- pm.cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		// SIGKILL se SIGTERM não funcionou
		if err := pm.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		pm.cmd.Wait()
		return fmt.Errorf("process killed after timeout")

	case err := <-done:
		if pm.stdin != nil {
			pm.stdin.Close()
		}
		return err
	}
}

// Wait aguarda o término do processo.
func (pm *ProcessManager) Wait() error {
	if pm.cmd == nil {
		return nil
	}
	return pm.cmd.Wait()
}

// GetPID retorna o PID do processo.
func (pm *ProcessManager) GetPID() int {
	return pm.PID
}

// contains verifica se qualquer string está em uma linha.
func contains(line string, strs ...string) bool {
	for _, s := range strs {
		if contains_str(line, s) {
			return true
		}
	}
	return false
}

func contains_str(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
