package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/Kaffyn/Vectora/assets"
	"github.com/Kaffyn/Vectora/internal/i18n"
	vecos "github.com/Kaffyn/Vectora/internal/os"
	"github.com/jeandeaual/go-locale"
)

// Constantes de versionamento - atualizadas durante build
const (
	VERSION = "1.0.0"
)

var (
	// VersionHash é injetado durante build via ldflags
	VersionHash = "development"
)

// VersionInfo representa a versão instalada
type VersionInfo struct {
	Version string
	Hash    string
}

var installPath string
var w fyne.Window
var selectedModel string
var selectedBackend string
var hardwareInfo map[string]interface{}
var recommendedModel map[string]interface{}

// Funções auxiliares para integração com MPM
func getMPMPath(installDir string) string {
	if filepath.VolumeName(installDir) != "" {
		return filepath.Join(installDir, "mpm.exe")
	}
	return filepath.Join(installDir, "mpm")
}

func callMPM(args ...string) (string, error) {
	var execPath string

	// Tentar encontrar mpm.exe no path de instalação
	if installPath != "" {
		mpmPath := getMPMPath(installPath)
		if _, err := os.Stat(mpmPath); err == nil {
			execPath = mpmPath
		}
	}

	// Fallback: procurar no PATH
	if execPath == "" {
		exe, _ := os.Executable()
		execPath = getMPMPath(filepath.Dir(exe))
	}

	cmd := exec.Command(execPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run mpm: %w", err)
	}

	return out.String(), nil
}

func detectHardware(installDir string) (map[string]interface{}, error) {
	output, err := callMPM("detect", "--json")
	if err != nil {
		return nil, err
	}

	// Trim whitespace and handle empty output
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("no output from hardware detection")
	}

	var hw map[string]interface{}
	if err := json.Unmarshal([]byte(output), &hw); err != nil {
		// Return nil with error instead of failing
		return nil, fmt.Errorf("failed to parse hardware info: %w", err)
	}

	return hw, nil
}

func recommendModel(installDir string) (map[string]interface{}, error) {
	output, err := callMPM("recommend", "--json")
	if err != nil {
		return nil, err
	}

	// Trim whitespace and handle empty output
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("no output from model recommendation")
	}

	var model map[string]interface{}
	if err := json.Unmarshal([]byte(output), &model); err != nil {
		return nil, fmt.Errorf("failed to parse model info: %w", err)
	}

	return model, nil
}

func listModels(installDir string) ([]map[string]interface{}, error) {
	// Tenta ler do catalog.json em vários caminhos possíveis
	var catalogPaths []string

	// 1. Relativo ao executável (../models/catalog.json)
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		catalogPaths = append(catalogPaths, filepath.Join(exeDir, "..", "models", "catalog.json"))
		catalogPaths = append(catalogPaths, filepath.Join(exeDir, "models", "catalog.json"))
	}

	// 2. No diretório de instalação
	if installDir != "" {
		catalogPaths = append(catalogPaths, filepath.Join(installDir, "models", "catalog.json"))
	}

	// 3. No diretório atual
	catalogPaths = append(catalogPaths, "./models/catalog.json")
	catalogPaths = append(catalogPaths, "models/catalog.json")

	// Tenta cada caminho
	var data []byte
	for _, catalogPath := range catalogPaths {
		d, err := os.ReadFile(catalogPath)
		if err == nil {
			data = d
			break
		}
	}

	// Se não conseguiu ler de nenhum caminho, fallback para mpm list
	if data == nil {
		return listModelsFromMPM()
	}

	type CatalogFormat struct {
		Models []map[string]interface{} `json:"models"`
	}

	var catalog CatalogFormat
	if err := json.Unmarshal(data, &catalog); err != nil {
		// Se falhar a parsing, fallback para mpm list
		return listModelsFromMPM()
	}

	return catalog.Models, nil
}

func listModelsFromMPM() ([]map[string]interface{}, error) {
	output, err := callMPM("list", "--json")
	if err != nil {
		return nil, err
	}

	// Trim whitespace and handle empty output
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("no output from models list")
	}

	var models []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &models); err != nil {
		return nil, fmt.Errorf("failed to parse models list: %w", err)
	}

	return models, nil
}

func installModel(installDir string, modelID string) (string, error) {
	return callMPM("install", "--model", modelID)
}

// ======== Funções de Versionamento ========

// ReadInstalledVersion lê version.txt da instalação existente
func ReadInstalledVersion(installDir string) (*VersionInfo, error) {
	versionFile := filepath.Join(installDir, "version.txt")

	data, err := os.ReadFile(versionFile)
	if err != nil {
		return nil, fmt.Errorf("arquivo de versão não encontrado: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("formato de arquivo de versão inválido")
	}

	info := &VersionInfo{}

	// Parse VERSION= e HASH=
	for _, line := range lines {
		if strings.HasPrefix(line, "VERSION=") {
			info.Version = strings.TrimPrefix(line, "VERSION=")
		} else if strings.HasPrefix(line, "HASH=") {
			info.Hash = strings.TrimPrefix(line, "HASH=")
		}
	}

	return info, nil
}

// WriteVersionFile cria/atualiza version.txt na instalação
func WriteVersionFile(installDir string) error {
	versionFile := filepath.Join(installDir, "version.txt")

	content := fmt.Sprintf("VERSION=%s\nHASH=%s\n", VERSION, VersionHash)
	return os.WriteFile(versionFile, []byte(content), 0644)
}

// CompareVersions detecta se atualização é necessária
// Retorna: (needsUpdate bool, reason string)
func CompareVersions(installed, current *VersionInfo) (bool, string) {
	// Versão mais nova sempre requer update
	if installed.Version != current.Version {
		return true, fmt.Sprintf("Nova versão disponível: %s → %s", installed.Version, current.Version)
	}

	// Mesma versão mas hash diferente = binários mudaram
	if installed.Hash != current.Hash {
		return true, fmt.Sprintf("Atualização de segurança/correção disponível (hash diferente)")
	}

	return false, ""
}

// UpdateInstallation atualiza apenas os executáveis, preservando dados
func UpdateInstallation(installDir string) error {
	// Extrair novos binários sobre os existentes
	assets := getInstallerAssets()
	binariesToUpdate := []string{
		"vectora.exe",
		"vectora-cli.exe",
		"lpm.exe",
		"mpm.exe",
	}

	for _, binName := range binariesToUpdate {
		if binData, exists := assets[binName]; exists && len(binData) > 0 {
			target := filepath.Join(installDir, binName)
			// Substituir arquivo mantendo permissões
			if err := os.WriteFile(target, binData, 0755); err != nil {
				return fmt.Errorf("falha ao atualizar %s: %w", binName, err)
			}
		}
	}

	// Atualizar version.txt
	if err := WriteVersionFile(installDir); err != nil {
		return fmt.Errorf("falha ao atualizar version.txt: %w", err)
	}

	return nil
}

// ======== Fim de Funções de Versionamento ========

func createLayout(titleText string, content fyne.CanvasObject, backFunc func(), nextFunc func(), nextLabel string) *fyne.Container {
	title := canvas.NewText(titleText, color.RGBA{R: 75, G: 58, B: 240, A: 255})
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}

	footerBg := canvas.NewRectangle(color.RGBA{R: 11, G: 10, B: 16, A: 255})

	btnBack := widget.NewButton("Voltar", backFunc)
	if backFunc == nil {
		btnBack.Disable()
	}

	btnNext := widget.NewButton(nextLabel, nextFunc)
	if nextFunc == nil {
		btnNext.Disable()
	}

	ecosystemLbl := canvas.NewText("Kaffyn Ecosystem", color.RGBA{R: 200, G: 200, B: 200, A: 255})
	ecosystemLbl.TextSize = 13

	footerContent := container.NewHBox(
		btnBack,
		layout.NewSpacer(),
		container.NewCenter(ecosystemLbl),
		layout.NewSpacer(),
		btnNext,
	)

	return container.NewBorder(
		title,
		container.NewStack(footerBg, container.NewPadded(footerContent)),
		nil, nil,
		content,
	)
}

// Wrapper para NewManager que reutiliza o padrão
func newSystemManager() (vecos.OSManager, error) {
	return vecos.NewManager()
}

// Detect if running in CLI mode (with Cobra subcommands/flags)
func isCLIMode() bool {
	if len(os.Args) < 2 {
		return false
	}

	// Check if first arg is a known Cobra subcommand
	subcommands := map[string]bool{
		"install":    true,
		"uninstall":  true,
		"help":       true,
		"-h":         true,
		"--help":     true,
		"--version":  true,
		"-v":         true,
	}

	return subcommands[os.Args[1]]
}

func runGUIMode() {
	// GUI mode uses AppData for installation, so no admin escalation needed
	// Vectora now installs to AppData (C:\Users\{username}\AppData\Local\Vectora)
	// which is user-writable and doesn't require admin privileges

	isUninstallMode := false
	if len(os.Args) >= 2 && os.Args[1] == "--uninstall" {
		isUninstallMode = true
	}

	a := app.New()
	a.Settings().SetTheme(&zyrisTheme{})

	w = a.NewWindow("Setup Vectora")
	w.Resize(fyne.NewSize(650, 550))
	w.CenterOnScreen()
	w.SetOnClosed(func() { os.Exit(0) })
	w.SetIcon(fyne.NewStaticResource("vectora-setup", assets.VectoraSetupIconData))

	// Re-get manager for logic
	systemManager, err := vecos.NewManager()
	if err != nil {
		w.SetContent(widget.NewLabel("FALHA: Erro ao carregar System Manager."))
		w.ShowAndRun()
		return
	}

	// Função auxiliar para SystemManager global
	newSysManager := func() (vecos.OSManager, error) {
		return vecos.NewManager()
	}
	_ = newSysManager // Evitar unused warning

	userLoc, err := locale.GetLanguage()
	if err == nil {
		if strings.HasPrefix(userLoc, "pt") {
			i18n.SetLanguage("pt")
		} else if strings.HasPrefix(userLoc, "es") {
			i18n.SetLanguage("es")
		} else if strings.HasPrefix(userLoc, "fr") {
			i18n.SetLanguage("fr")
		} else {
			i18n.SetLanguage("en")
		}
	} else {
		i18n.SetLanguage("en")
	}

	var showStepLang, showStepPath, showStepInstall func()
	var showStepConfigMode, showStepConfigGemini, showStepConfigQwen, showStepFinish func()
	var showStepDetectHardware, showStepRecommendModel, showStepSelectBackend, showStepChooseModel, showStepInstallModel func()
	var showUninstallProgress, showUpdateProgress, showAlreadyInstalled func(string)

	showUninstallProgress = func(existingPath string) {
		progress := widget.NewProgressBar()
		content := container.NewVBox(
			// Destroying Local Vectora Instance and Windows links...,
			progress,
		)

		wrapper := createLayout("Uninstalling Vectora", content, nil, nil, "Finalizar")
		w.SetContent(wrapper)

		go func() {
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 60)
				progress.SetValue(float64(i) / 20.0)
			}

			systemManager.UnregisterApp(existingPath)
			daemonBin := "vectora"
			if filepath.VolumeName(existingPath) != "" {
				daemonBin = "vectora.exe"
			}
			os.Remove(filepath.Join(existingPath, daemonBin))

			progress.SetValue(1.0)

			finalContent := container.NewVBox(widget.NewLabel("Desinstalação e Limpeza Concluídas com Sucesso."))
			w.SetContent(createLayout("Uninstalled", finalContent, nil, func() { w.Close() }, "Exit"))
		}()
	}

	showUpdateProgress = func(updatePath string) {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		statusLbl := widget.NewLabel("Atualizando executáveis do Vectora...")
		content := container.NewVBox(statusLbl, progress)
		wrapper := createLayout("Atualizando Vectora", content, nil, nil, "Finalizar")
		w.SetContent(wrapper)

		go func() {
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 30)
				progress.SetValue(float64(i) / 20.0)
			}

			// Executar apenas atualização (preserva dados)
			if err := UpdateInstallation(updatePath); err != nil {
				dialog.ShowError(err, w)
				return
			}

			progress.SetValue(1.0)
			statusLbl.SetText("Atualização Concluída com Sucesso! ✓")
			time.Sleep(1000 * time.Millisecond)

			doneContent := container.NewVBox(
				widget.NewLabel("Vectora foi atualizado com sucesso."),
				widget.NewLabel(""),
				widget.NewLabel("Todos os seus dados foram preservados:"),
				widget.NewLabel("• Configurações"),
				widget.NewLabel("• Banco de dados"),
				widget.NewLabel("• Índices Chroma"),
			)
			w.SetContent(createLayout("Atualização Concluída", doneContent, nil, func() { w.Close() }, "Fechar"))
		}()
	}

	showAlreadyInstalled = func(existingPath string) {
		// Ler versão instalada
		installedVersion, err := ReadInstalledVersion(existingPath)
		if err != nil {
			installedVersion = nil // Versão antiga, sem version.txt
		}

		// Criar VersionInfo da versão atual (setup em execução)
		currentVersion := &VersionInfo{
			Version: VERSION,
			Hash:    VersionHash,
		}

		// Verificar se atualização é necessária
		needsUpdate := false
		updateReason := ""
		if installedVersion != nil {
			needsUpdate, updateReason = CompareVersions(installedVersion, currentVersion)
		} else {
			// Arquivo version.txt não existe (instalação muito antiga)
			needsUpdate = true
			updateReason = "Instalação detectada, versão desconhecida"
		}

		// Preparar mensagem
		var mainLabel string
		if needsUpdate {
			mainLabel = fmt.Sprintf("O Vectora já está instalado.\n\n⚠️ %s", updateReason)
		} else {
			mainLabel = "O Vectora já está instalado no seu sistema. O que deseja fazer?"
		}

		lbl := widget.NewLabel(mainLabel)

		// NOVO: Botão de Atualizar (aparece apenas se necessário)
		// Atualizar = apenas novos executáveis, dados preservados
		btnUpdate := widget.NewButton("⬆️ Atualizar Vectora", func() {
			installPath = existingPath
			showUpdateProgress(existingPath)
		})
		btnUpdate.Importance = widget.HighImportance

		// Botão para Reinstalar (desinstalação completa + instalação limpa)
		btnReinstall := widget.NewButton("↻ Reinstalar Vectora", func() {
			installPath = existingPath
			showUninstallProgress(existingPath)
		})
		btnReinstall.Importance = widget.HighImportance

		// Botão para Gerenciar Modelos
		btnManage := widget.NewButton("⚙️ Gerenciar Modelos e LlamaCpp", func() {
			installPath = existingPath
			showStepChooseModel()
		})

		// Botão para Desinstalar
		btnUninstall := widget.NewButton("🗑️ Desfazer Instalação (Wipe)", func() {
			showUninstallProgress(existingPath)
		})
		btnUninstall.Importance = widget.LowImportance

		// Grid com as opções
		optionsGrid := container.NewGridWithColumns(1)

		// Adicionar botão de update primeiro se necessário
		if needsUpdate {
			optionsGrid.Add(btnUpdate)
		}

		optionsGrid.Add(btnReinstall)
		optionsGrid.Add(btnManage)
		optionsGrid.Add(btnUninstall)

		content := container.NewVBox(
			lbl,
			widget.NewLabel(""),
			optionsGrid,
		)

		w.SetContent(createLayout("Vectora Detectado", content, nil, func() { w.Close() }, "Fechar"))
	}

	showStepLang = func() {
		languages := []struct {
			name  string
			code  string
			emoji string
		}{
			{"English", "en", "🇬🇧"},
			{"Português", "pt", "🇧🇷"},
			{"Español", "es", "🇪🇸"},
			{"Français", "fr", "🇫🇷"},
		}

		currentLang := i18n.GetCurrentLang()

		// Criar spacers pequenos (mínimo espaçamento)
		minSpacer := canvas.NewRectangle(color.Transparent)
		minSpacer.SetMinSize(fyne.NewSize(0, 4))

		content := container.NewVBox(
			widget.NewLabel(i18n.T("inst_welcome")),
		)

		spacer1 := canvas.NewRectangle(color.Transparent)
		spacer1.SetMinSize(fyne.NewSize(0, 4))
		content.Add(spacer1)

		content.Add(widget.NewLabel("Este assistente ajudará você a configurar o ambiente de IA\nno seu sistema local."))

		spacer2 := canvas.NewRectangle(color.Transparent)
		spacer2.SetMinSize(fyne.NewSize(0, 8))
		content.Add(spacer2)

		content.Add(widget.NewLabel(i18n.T("inst_select_lang")))

		// Criar grid de botões de idioma
		langGrid := container.NewGridWithColumns(2)
		for _, lang := range languages {
			langCode := lang.code
			btn := widget.NewButton(fmt.Sprintf("%s %s", lang.emoji, lang.name), func() {
				i18n.SetLanguage(langCode)
				showStepLang()
			})
			if langCode == currentLang {
				btn.Importance = widget.HighImportance
			}
			langGrid.Add(btn)
		}

		content.Add(langGrid)
		w.SetContent(createLayout("Vectora Setup", content, nil, showStepPath, i18n.T("inst_btn_next")+" >"))
	}

	showStepPath = func() {
		defaultPath, _ := systemManager.GetInstallDir() // Changes to Program Files on Windows
		if installPath == "" {
			installPath = defaultPath
		}

		pathEntry := widget.NewEntry()
		pathEntry.SetText(installPath)

		btnFolder := widget.NewButton(i18n.T("inst_btn_browse"), func() {
			dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
				if lu != nil {
					pathEntry.SetText(lu.Path())
					installPath = lu.Path()
				}
			}, w)
		})

		content := container.NewVBox(
			widget.NewLabel("Selecione a pasta onde os executáveis serão instalados:"),
			container.NewBorder(nil, nil, nil, btnFolder, pathEntry),
			widget.NewLabel("O instalador criará automaticamente o diretório caso\nele não exista."),
		)

		nextCmd := func() {
			installPath = pathEntry.Text
			showStepInstall()
		}
		w.SetContent(createLayout("Local de Instalação", content, showStepLang, nextCmd, "Avançar >"))
	}

	showStepInstall = func() {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		content := container.NewVBox(
			widget.NewLabel("Desempacotando binários do ecossistema e escrevendo manifestos..."),
			progress,
		)

		wrapper := createLayout("Instalando...", content, nil, nil, "Finalizar")
		w.SetContent(wrapper)

		go func() {
			for i := 0; i <= 20; i++ {
				time.Sleep(time.Millisecond * 30)
				progress.SetValue(float64(i) / 20.0)
			}

			// Official Scaffolding hierarchy in Program Files (Requires Admin)
			if err := os.MkdirAll(installPath, 0755); err != nil {
				dialog.ShowError(fmt.Errorf("PERMISSION DENIED: Run the installer as Administrator or request IT access.\n(%v)", err), w)
				return
			}
			os.MkdirAll(filepath.Join(installPath, "data", "chroma"), 0755)
			os.MkdirAll(filepath.Join(installPath, "logs"), 0755)
			os.MkdirAll(filepath.Join(installPath, "backups"), 0755)

			assets := getInstallerAssets()

			// Extract all binaries
			binariesToExtract := []string{
				"vectora.exe",
				"vectora-cli.exe",
				"lpm.exe",
				"mpm.exe",
			}

			for _, binName := range binariesToExtract {
				if binData, exists := assets[binName]; exists && len(binData) > 0 {
					target := filepath.Join(installPath, binName)
					if err := os.WriteFile(target, binData, 0755); err != nil {
						fmt.Printf("[WARNING] Failed to extract %s: %v\n", binName, err)
					}
				}
			}
			fmt.Println("--- Engine Extraction and Setup Completed ---")

			srcSelf, _ := os.Executable()
			uninstallerPath := filepath.Join(installPath, "vectora-uninstaller.exe")
			copySysFile(srcSelf, uninstallerPath)

			systemManager.RegisterApp(installPath)

			// Escrever version.txt (APENAS em instalações novas)
			if err := WriteVersionFile(installPath); err != nil {
				fmt.Printf("[WARNING] Falha ao escrever version.txt: %v\n", err)
				// Não interromper instalação se falhar
			}

			progress.SetValue(1.0)

			doneContent := container.NewVBox(widget.NewLabel(i18n.T("inst_install_done")), widget.NewLabel(""), widget.NewLabel(i18n.T("inst_configure_engine")))
			w.SetContent(createLayout("Instalação Concluída", doneContent, nil, showStepConfigMode, "Configurar IA >"))
		}()
	}

	showStepConfigMode = func() {
		radio := widget.NewRadioGroup([]string{
			"Usar Gemini API (Apenas Chave / Leve RAM)",
			"Usar Qwen3 (100% Privado / Download Pesado)",
		}, func(s string) {})
		radio.SetSelected("Usar Qwen3 (100% Privado / Download Pesado)")

		content := container.NewVBox(widget.NewLabel(i18n.T("inst_select_backend")), radio)

		nextCmd := func() {
			if radio.Selected == "Usar Gemini API (Apenas Chave / Leve RAM)" {
				showStepConfigGemini()
			} else {
				showStepConfigQwen()
			}
		}
		w.SetContent(createLayout("Motor Cognitivo", content, nil, nextCmd, "Avançar >"))
	}

	showStepConfigGemini = func() {
		lbl := widget.NewLabel("Insira sua chave de API do Google AI Studio:")
		entry := widget.NewPasswordEntry()
		entry.SetPlaceHolder("AIzaSy...")

		content := container.NewVBox(lbl, entry, widget.NewLabel("Ela será arquivada e criptografada em segurança na sua máquina local."))

		nextCmd := func() {
			showStepFinish()
		}
		w.SetContent(createLayout("Validar Access Token", content, showStepConfigMode, nextCmd, "Avançar >"))
	}

	// Detectar hardware e recomendarmodelo
	showStepDetectHardware = func() {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		statusLbl := widget.NewLabel("Detectando especificações do sistema...")
		content := container.NewVBox(statusLbl, progress)

		wrapper := createLayout("Detectando Hardware", content, nil, nil, "Próximo")
		w.SetContent(wrapper)

		go func() {
			// Simular progresso
			for i := 0; i < 10; i++ {
				time.Sleep(100 * time.Millisecond)
				progress.SetValue(float64(i) / 10.0)
			}

			hw, err := detectHardware(installPath)
			if err != nil {
				// Detecção falhou, prosseguir sem hardware info
				statusLbl.SetText("Não foi possível detectar o hardware. Continuando com detecção manual...")
				hardwareInfo = nil // Deixar nulo para indicar que falhou
				progress.SetValue(1.0)
				time.Sleep(1500 * time.Millisecond)
				// Pular recomendação e ir direto para seleção
				showStepChooseModel()
				return
			}

			hardwareInfo = hw
			progress.SetValue(1.0)
			time.Sleep(500 * time.Millisecond)

			// Passar para próximo step automaticamente
			showStepRecommendModel()
		}()
	}

	// Recomendar modelo
	showStepRecommendModel = func() {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		statusLbl := widget.NewLabel("Recomendando modelo para seu hardware...")
		content := container.NewVBox(statusLbl, progress)

		wrapper := createLayout("Recomendação de Modelo", content, nil, nil, "Próximo")
		w.SetContent(wrapper)

		go func() {
			// Simular progresso
			for i := 0; i < 10; i++ {
				time.Sleep(100 * time.Millisecond)
				progress.SetValue(float64(i) / 10.0)
			}

			model, err := recommendModel(installPath)
			if err != nil {
				statusLbl.SetText(fmt.Sprintf("Erro ao recomendar modelo: %v", err))
				return
			}

			recommendedModel = model
			selectedModel = model["id"].(string)
			progress.SetValue(1.0)
			time.Sleep(500 * time.Millisecond)

			// Passar para próximo step
			showStepSelectBackend()
		}()
	}

	// Selecionar backend de GPU/CPU para llamacpp
	showStepSelectBackend = func() {
		gpuType := ""
		if hardwareInfo != nil {
			if gpu, ok := hardwareInfo["GPUType"].(string); ok {
				gpuType = gpu
			}
		}

		// Opções baseadas no GPU detectado
		type BackendOption struct {
			id    string
			label string
		}
		var backendOptions []BackendOption

		backendOptions = append(backendOptions, BackendOption{"cpu", "💻 CPU"})

		if gpuType == "cuda" {
			backendOptions = append(backendOptions, BackendOption{"cuda", "🎮 NVIDIA CUDA"})
			backendOptions = append(backendOptions, BackendOption{"vulkan", "🌐 Vulkan"})
		} else if gpuType == "metal" {
			backendOptions = append(backendOptions, BackendOption{"metal", "🍎 Metal"})
			backendOptions = append(backendOptions, BackendOption{"vulkan", "🌐 Vulkan"})
		} else if gpuType == "vulkan" || (gpuType != "none" && gpuType != "") {
			backendOptions = append(backendOptions, BackendOption{"vulkan", "🌐 Vulkan"})
		}

		selectedBackend = "cpu"

		// Grid de botões para backends
		backendGrid := container.NewGridWithColumns(2)
		for _, opt := range backendOptions {
			backend := opt.id
			btn := widget.NewButton(opt.label, func() {
				selectedBackend = backend
			})
			if backend == selectedBackend {
				btn.Importance = widget.HighImportance
			}
			backendGrid.Add(btn)
		}

		content := container.NewVBox(
			widget.NewLabel("Selecione o backend para llama.cpp:"),
			widget.NewLabel(""),
			backendGrid,
			widget.NewLabel(""),
			widget.NewLabel(fmt.Sprintf("🎮 GPU: %s", gpuType)),
		)

		nextCmd := func() {
			showStepChooseModel()
		}

		w.SetContent(createLayout("Backend de Processamento", content, showStepRecommendModel, nextCmd, "Próximo"))
	}

	// Permitir escolher modelos com multi-seleção
	showStepChooseModel = func() {
		var allModels []map[string]interface{}
		models, err := listModels(installPath)
		if err != nil {
			lbl := widget.NewLabel(fmt.Sprintf("Erro ao listar modelos: %v", err))
			w.SetContent(createLayout("Erro", lbl, showStepDetectHardware, nil, ""))
			return
		}

		// Filtrar apenas modelos Qwen3 e Qwen2.5-Coder (remover Vision e outros)
		for _, m := range models {
			id := m["id"].(string)
			if strings.Contains(id, "qwen3") || strings.Contains(id, "qwen2.5-coder") {
				allModels = append(allModels, m)
			}
		}

		// Separar modelos por tipo (Text/Coder juntos, Embedding separado)
		var normalModels, embeddingModels []map[string]interface{}
		for _, m := range allModels {
			id := m["id"].(string)
			if strings.Contains(id, "-embed") {
				embeddingModels = append(embeddingModels, m)
			} else {
				// Inclui tanto Qwen3 quanto Qwen2.5-Coder na mesma lista
				normalModels = append(normalModels, m)
			}
		}

		// Mapas para rastrear seleções
		selectedModels := make(map[string]bool)

		// Usar modelos recomendados se detecção de hardware foi bem-sucedida
		if hardwareInfo != nil {
			// Por padrão: usar modelo recomendado para o hardware + embedding compatível
			if recommendedModel != nil && recommendedModel["id"] != nil {
				recommendedID := recommendedModel["id"].(string)
				selectedModels[recommendedID] = true
			}

			// Selecionar embedding compatível com o hardware
			// Se RAM >= 12GB: usar embedding-4b (melhor qualidade)
			// Caso contrário: usar embedding-0.6b (mais leve)
			var ramGB float64
			switch v := hardwareInfo["RAM"].(type) {
			case float64:
				ramGB = v / (1024 * 1024 * 1024)
			case int64:
				ramGB = float64(v) / (1024 * 1024 * 1024)
			case int:
				ramGB = float64(v) / (1024 * 1024 * 1024)
			}

			if ramGB >= 12 {
				selectedModels["qwen3-embedding-4b"] = true
			} else {
				selectedModels["qwen3-embedding-0.6b"] = true
			}
		} else {
			// Se detecção de hardware falhou, selecionar modelos padrão sensatos
			// Qwen 3 4B é um bom middle ground para maioria dos sistemas
			selectedModels["qwen3-4b"] = true
			// Embedding 0.6B é mais leve e funciona bem
			selectedModels["qwen3-embedding-0.6b"] = true
		}

		// Helper para criar checkbox
		createCheckbox := func(modelID, modelName string, models map[string]bool) *widget.Check {
			id := modelID // Capture in closure
			check := widget.NewCheck(modelName, func(checked bool) {
				if checked {
					models[id] = true
				} else {
					delete(models, id)
				}
			})
			check.Checked = models[modelID]
			return check
		}

		// Grid de modelos (sem scroll, direto na tela)
		modelsContent := container.NewVBox()

		if len(normalModels) > 0 {
			modelsContent.Add(widget.NewLabel("📦 Modelos Padrão:"))
			normalGrid := container.NewGridWithColumns(2)
			for _, m := range normalModels {
				id := m["id"].(string)
				name := m["name"].(string)
				normalGrid.Add(createCheckbox(id, name, selectedModels))
			}
			modelsContent.Add(normalGrid)
		}

		if len(embeddingModels) > 0 {
			modelsContent.Add(widget.NewLabel("📚 Modelos Embedding:"))
			embGrid := container.NewGridWithColumns(2)
			for _, m := range embeddingModels {
				id := m["id"].(string)
				name := m["name"].(string)
				embGrid.Add(createCheckbox(id, name, selectedModels))
			}
			modelsContent.Add(embGrid)
		}

		content := container.NewVBox(
			widget.NewLabel(i18n.T("inst_select_models")),
			modelsContent,
		)

		nextCmd := func() {
			// Converter mapa de seleção para lista
			var selected []string
			for id := range selectedModels {
				if id != "" {
					selected = append(selected, id)
				}
			}

			if len(selected) == 0 {
				dialog.ShowError(fmt.Errorf("selecione pelo menos um modelo"), w)
				return
			}

			// Validar que há um modelo base se há embedding
			hasBase := false
			hasEmbedding := false
			for _, id := range selected {
				if strings.Contains(id, "-embed") {
					hasEmbedding = true
				} else {
					hasBase = true
				}
			}

			if hasEmbedding && !hasBase {
				dialog.ShowError(fmt.Errorf("ao usar modelos Embedding, você deve selecionar também um modelo base (padrão)"), w)
				return
			}

			// Guardar seleção e instalar
			selectedModel = strings.Join(selected, ",")
			showStepInstallModel()
		}

		// Se hardware detection falhou, voltar ao config mode (seleção de AI engine)
		// Se hardware detection funcionou, voltar ao backend selection
		var backCmd func()
		if hardwareInfo == nil {
			backCmd = showStepConfigMode
		} else {
			backCmd = showStepSelectBackend
		}

		w.SetContent(createLayout("Escolher Modelos", content, backCmd, nextCmd, "Instalar >"))
	}

	// Instalar modelos
	showStepInstallModel = func() {
		progress := widget.NewProgressBar()
		progress.SetValue(0.0)

		models := strings.Split(selectedModel, ",")
		statusLbl := widget.NewLabel(fmt.Sprintf("Preparando instalação de %d modelo(s)...", len(models)))
		content := container.NewVBox(statusLbl, progress)

		wrapper := createLayout("Instalando Modelos", content, nil, nil, "Concluir")
		w.SetContent(wrapper)

		go func() {
			// Simular progresso de preparação
			for i := 0; i < 5; i++ {
				time.Sleep(100 * time.Millisecond)
				progress.SetValue(float64(i) / 10.0)
			}

			successCount := 0
			failureCount := 0
			progressStep := 0.6 / float64(len(models))

			for idx, modelID := range models {
				modelID = strings.TrimSpace(modelID)
				statusLbl.SetText(fmt.Sprintf("Instalando modelo %d/%d: %s...", idx+1, len(models), modelID))
				_, err := installModel(installPath, modelID)

				if err != nil {
					failureCount++
					statusLbl.SetText(fmt.Sprintf("⚠️  Erro em %s: %v\n(será instalado na primeira inicialização)", modelID, err))
				} else {
					successCount++
					statusLbl.SetText(fmt.Sprintf("✓ %s instalado com sucesso!", modelID))
				}

				progress.SetValue(0.1 + float64(idx+1)*progressStep)
				time.Sleep(300 * time.Millisecond)
			}

			// Resumo final
			if failureCount == 0 {
				statusLbl.SetText(fmt.Sprintf("✓ Todos os %d modelo(s) foram instalados com sucesso!", successCount))
			} else {
				statusLbl.SetText(fmt.Sprintf("✓ %d modelo(s) instalado(s), %d com aviso (serão instalados na primeira execução)", successCount, failureCount))
			}

			progress.SetValue(1.0)
			time.Sleep(500 * time.Millisecond)

			// Passar para finish
			showStepFinish()
		}()
	}

	showStepConfigQwen = func() {
		// Iniciar fluxo de Qwen3
		showStepDetectHardware()
	}

	showStepFinish = func() {
		doneContent := container.NewVBox(widget.NewLabel("O Vectora foi instalado com sucesso."))
		w.SetContent(createLayout("Sucesso", doneContent, nil, func() { w.Close() }, "Encerrar"))
	}

	if isUninstallMode {
		existingPath := systemManager.IsInstalled()
		if existingPath != "" {
			showUninstallProgress(existingPath)
		} else {
			execPath, _ := os.Executable()
			showUninstallProgress(filepath.Dir(execPath))
		}
	} else if existingPath := systemManager.IsInstalled(); existingPath != "" {
		showAlreadyInstalled(existingPath)
	} else {
		showStepLang()
	}

	w.ShowAndRun()
}

// Main function - delegates to CLI or GUI mode
func main() {
	// Check CLI mode BEFORE hiding console
	if isCLIMode() {
		runCLIMode()
	} else {
		// Hide console window on Windows ONLY for GUI mode
		if runtime.GOOS == "windows" {
			getConsoleWindow := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")
			showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")

			hwnd, _, _ := getConsoleWindow.Call()
			if hwnd != 0 {
				showWindow.Call(hwnd, 0) // SW_HIDE = 0
			}
		}
		runGUIMode()
	}
}

func copySysFile(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer out.Close()
	io.Copy(out, in)
}
