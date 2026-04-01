package engines

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extractor gerencia extração de arquivos comprimidos.
type Extractor struct {
	targetDir string
}

// NewExtractor cria um novo extractor para um diretório alvo.
func NewExtractor(targetDir string) *Extractor {
	return &Extractor{targetDir: targetDir}
}

// ExtractZIP extrai apenas arquivos específicos de um ZIP.
// Se fileFilter é vazio, extrai tudo. Caso contrário, extrai apenas arquivos
// cujos caminhos contenham os padrões em fileFilter.
func (e *Extractor) ExtractZIP(zipPath string, fileFilter []string) error {
	if err := os.MkdirAll(e.targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		// Pular diretórios
		if file.FileInfo().IsDir() {
			continue
		}

		// Aplicar filter se fornecido
		if len(fileFilter) > 0 {
			if !shouldExtract(file.Name, fileFilter) {
				continue
			}
		}

		if err := e.extractFile(file); err != nil {
			return err
		}
	}

	return nil
}

// ExtractTarGZ extrai um arquivo .tar.gz.
// Útil para builds do Linux.
func (e *Extractor) ExtractTarGZ(tarGzPath string) error {
	if err := os.MkdirAll(e.targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	f, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	// Para simplicidade, usar tar via command line no Linux
	// No Windows, converter para ZIP primeiro ou usar CGO
	return fmt.Errorf("tar.gz extraction not yet implemented (use system tar command)")
}

// extractFile extrai um arquivo individual do ZIP.
func (e *Extractor) extractFile(file *zip.File) error {
	// Usar apenas o basename do arquivo (ignora paths dentro do ZIP)
	baseName := filepath.Base(file.Name)

	destPath := filepath.Join(e.targetDir, baseName)

	// Ler do ZIP
	inFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer inFile.Close()

	// Criar arquivo de destino
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer outFile.Close()

	// Copiar conteúdo
	if _, err := io.Copy(outFile, inFile); err != nil {
		outFile.Close()
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	// Fechar antes de chmod
	outFile.Close()

	// Definir permissões executáveis se necessário
	if isExecutable(file.Name) {
		if err := os.Chmod(destPath, 0755); err != nil {
			return fmt.Errorf("failed to chmod file: %w", err)
		}
	}

	return nil
}

// shouldExtract verifica se um arquivo deve ser extraído baseado no filter.
func shouldExtract(filePath string, filter []string) bool {
	for _, pattern := range filter {
		// Match: substring contido no path
		if strings.Contains(strings.ToLower(filePath), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isExecutable detecta se um arquivo deve ser executável.
func isExecutable(filePath string) bool {
	name := strings.ToLower(filepath.Base(filePath))

	// Extensões executáveis no Windows
	windowsExecs := []string{".exe", ".dll", ".so"}
	for _, ext := range windowsExecs {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}

	// Arquivos sem extensão são frequentemente executáveis no Unix
	if !strings.Contains(filepath.Base(filePath), ".") {
		return true
	}

	// Qualquer arquivo em diretório "bin" provavelmente é executável
	if strings.Contains(filePath, "bin/") || strings.Contains(filePath, "bin\\") {
		return true
	}

	return false
}

// ListFilesInZip lista os arquivos dentro de um ZIP (debug).
func ListFilesInZip(zipPath string) ([]string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	var files []string
	for _, file := range reader.File {
		files = append(files, file.Name)
	}
	return files, nil
}

// ValidateZIP verifica se um arquivo ZIP é válido.
func ValidateZIP(zipPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("invalid zip file: %w", err)
	}
	defer reader.Close()

	// Tentar ler um arquivo para validar integridade
	if len(reader.File) == 0 {
		return fmt.Errorf("zip file is empty")
	}

	// Validar primeiro arquivo
	f := reader.File[0]
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to read zip content: %w", err)
	}
	defer rc.Close()

	// Ler alguns bytes
	buf := make([]byte, 1024)
	if _, err := rc.Read(buf); err != nil && err != io.EOF {
		return fmt.Errorf("corruption detected in zip: %w", err)
	}

	return nil
}
