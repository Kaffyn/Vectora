package main

import (
	"os"
	"os/exec"
)

func main() {
	// Wrapper que encaminha para o daemon com o subcomando 'chat'
	daemonPath := "vectora.exe"

	// Procurar pelo daemon no mesmo diretório ou no PATH
	daemonExe, err := exec.LookPath(daemonPath)
	if err != nil {
		// Se não encontrar, tentar no diretório atual
		if _, err := os.Stat(daemonPath); err != nil {
			os.Stderr.WriteString("Erro: vectora daemon não encontrado. Certifique-se que vectora.exe está instalado.\n")
			os.Exit(1)
		}
		daemonExe = daemonPath
	}

	// Passar todos os argumentos para o daemon com o subcomando 'chat'
	args := []string{"chat"}
	args = append(args, os.Args[1:]...)

	cmd := exec.Command(daemonExe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
