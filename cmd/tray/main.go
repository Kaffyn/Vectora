package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/getlantern/systray"
)

var (
	engineCmd *exec.Cmd
	webCmd    *exec.Cmd
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Vectora")
	systray.SetTooltip("Vectora - AI Engine")

	// Usaremos um botão vazio por não termos ícone agora
	// Para uso real, deve se carregar um byte slice .ico
	systray.SetIcon(getDummyIcon())

	mEngineStart := systray.AddMenuItem("Start RAG Engine", "Starts the Go Backend")
	mEngineStop := systray.AddMenuItem("Stop RAG Engine", "Stops the Go Backend")
	systray.AddSeparator()

	mWebStart := systray.AddMenuItem("Start Web UI", "Starts the Next.js Frontend")
	mWebStop := systray.AddMenuItem("Stop Web UI", "Stops the Next.js Frontend")
	systray.AddSeparator()

	mACP := systray.AddMenuItem("Autonomy Level (ACP)", "Set AI agent autonomy level")
	mLevelAsk := mACP.AddSubMenuItem("Ask Everything", "Confirm every AI action")
	mLevelGuarded := mACP.AddSubMenuItem("Guarded (Safe)", "Auto-approve reads, confirm writes")
	mLevelYOLO := mACP.AddSubMenuItem("YOLO (Full Auto)", "IA does everything without asking")
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit Vectora", "Exits the manager and all services")

	mLevelGuarded.Check() // Default

	mEngineStop.Disable()
	mWebStop.Disable()

	rootPath, _ := os.Getwd()

	go func() {
		for {
			select {
			case <-mEngineStart.ClickedCh:
				mEngineStart.Disable()
				if err := startEngine(rootPath); err != nil {
					fmt.Printf("Engine start error: %v\n", err)
					mEngineStart.Enable()
				} else {
					mEngineStop.Enable()
				}
			case <-mEngineStop.ClickedCh:
				mEngineStop.Disable()
				stopEngine()
				mEngineStart.Enable()

			case <-mWebStart.ClickedCh:
				mWebStart.Disable()
				if err := startWeb(rootPath); err != nil {
					fmt.Printf("Web start error: %v\n", err)
					mWebStart.Enable()
				} else {
					mWebStop.Enable()
				}
			case <-mWebStop.ClickedCh:
				mWebStop.Disable()
				stopWeb()
				mWebStart.Enable()

			case <-mLevelAsk.ClickedCh:
				mLevelAsk.Check()
				mLevelGuarded.Uncheck()
				mLevelYOLO.Uncheck()
				setACPLevel("ask-any")
			case <-mLevelGuarded.ClickedCh:
				mLevelAsk.Uncheck()
				mLevelGuarded.Check()
				mLevelYOLO.Uncheck()
				setACPLevel("guarded")
			case <-mLevelYOLO.ClickedCh:
				mLevelAsk.Uncheck()
				mLevelGuarded.Uncheck()
				mLevelYOLO.Check()
				setACPLevel("yolo")

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func setACPLevel(level string) {
	// Chamada para o Core Dashboard / Settings
	fmt.Printf("🛡️ Tray -> Core: Alterando nível ACP para [%s]\n", level)
	// http.Post("http://localhost:8080/api/settings", ... (Ajustar quando endpoint Settings suportar ACP))
}

func onExit() {
	// Cleanup all child processes cleanly
	stopWeb()
	stopEngine()
}

func startEngine(rootPath string) error {
	exePath := filepath.Join(rootPath, "vectora-cli.exe")
	engineCmd = exec.Command(exePath, "start")
	
	// Esconder o console preta (Somente Windows)
	engineCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return engineCmd.Start()
}

func stopEngine() {
	if engineCmd != nil && engineCmd.Process != nil {
		engineCmd.Process.Kill()
		engineCmd.Wait()
		engineCmd = nil
	}
}

func startWeb(rootPath string) error {
	webPath := filepath.Join(rootPath, "src", "web")
	
	webCmd = exec.Command("bun", "dev")
	webCmd.Dir = webPath
	
	webCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return webCmd.Start()
}

func stopWeb() {
	if webCmd != nil && webCmd.Process != nil {
		webCmd.Process.Kill()
		webCmd.Wait()
		webCmd = nil
	}
}

func getDummyIcon() []byte {
	// 1x1 transparent ICO just to avoid panic
	return []byte{
		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00,
		0x18, 0x00, 0x30, 0x00, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}
