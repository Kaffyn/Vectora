package policies

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Guardian encapsula todas as regras de segurança
type Guardian struct {
	TrustFolder            string
	BlockedExts            map[string]bool
	BlockedFiles           map[string]bool
	ExcludedDirs           map[string]bool
	SecretRegexes          []*regexp.Regexp
	ModifyIgnorePatterns   []string
	ModifyUnignorePatterns []string
}

func NewGuardian(trustFolder string) *Guardian {
	g := &Guardian{
		TrustFolder: trustFolder,
		BlockedExts: map[string]bool{
			".db": true, ".sqlite": true, ".exe": true, ".dll": true,
			".key": true, ".pem": true, ".env": true, ".log": true,
		},
		BlockedFiles: map[string]bool{
			".env": true, "secrets.yml": true, "id_rsa": true,
		},
		ExcludedDirs: map[string]bool{
			"node_modules": true, ".git": true, "vendor": true,
			"dist": true, "build": true,
		},
	}

	// Compila regex de segredos
	g.SecretRegexes = []*regexp.Regexp{
		regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
		regexp.MustCompile(`sk-[a-zA-Z0-9]{48}`),
	}

	// Tenta carregar .vectoraignore (erros são ignorados silenciosamente caso não exista)
	ignorePath := filepath.Join(trustFolder, ".vectoraignore")
	if data, err := os.ReadFile(ignorePath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "!") {
				g.ModifyUnignorePatterns = append(g.ModifyUnignorePatterns, strings.TrimPrefix(line, "!"))
			} else {
				g.ModifyIgnorePatterns = append(g.ModifyIgnorePatterns, line)
			}
		}
	}

	return g
}

// IsPathSafe verifica escopo e symlinks
func (g *Guardian) IsPathSafe(targetPath string) bool {
	_, err := g.ValidatePath(targetPath)
	return err == nil
}

// ValidatePath verifica se o caminho está dentro do trust folder e retorna o caminho absoluto limpo.
func (g *Guardian) ValidatePath(targetPath string) (string, error) {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}

	// Resolve symlinks para evitar bypass
	realPath, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		// Se não existe, ainda podemos validar o path pretendido
		realPath = absTarget
	}

	absTrust, _ := filepath.Abs(g.TrustFolder)

	// Verifica se o path real está dentro do trust folder
	if !strings.HasPrefix(realPath, absTrust+string(filepath.Separator)) && realPath != absTrust {
		return "", fmt.Errorf("path_security_breach: path '%s' is outside the trust folder '%s'", realPath, absTrust)
	}

	return realPath, nil
}

// IsProtected verifica extensões e nomes de arquivo bloqueados
func (g *Guardian) IsProtected(path string) bool {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	if g.BlockedFiles[base] {
		return true
	}
	if g.BlockedExts[ext] {
		return true
	}
	return false
}

// IsModificationBlocked verifica se o arquivo bate com as regras do .vectoraignore
func (g *Guardian) IsModificationBlocked(path string) bool {
	if len(g.ModifyIgnorePatterns) == 0 && len(g.ModifyUnignorePatterns) == 0 {
		return false
	}

	// Usaremos caminhos relativos ao TrustFolder, se possível
	relPath := path
	if rel, err := filepath.Rel(g.TrustFolder, path); err == nil {
		relPath = rel
	}

	// Normaliza separadores no Windows para barra normal para funcionar bem no filepath.Match e com nomes do repositório
	relPath = filepath.ToSlash(relPath)
	filename := filepath.Base(relPath)

	isIgnored := false

	for _, pattern := range g.ModifyIgnorePatterns {
		// Checa tanto a string completa quanto apenas o arquivo local
		if m, _ := filepath.Match(pattern, relPath); m {
			isIgnored = true
		} else if m, _ := filepath.Match(pattern, filename); m {
			isIgnored = true
		} else if strings.Contains(pattern, "/") {
			// Se o padrão tem barra, ele pode querer dar match no início do path
			if strings.HasPrefix(relPath, pattern) {
				isIgnored = true
			} else {
				m, _ := filepath.Match(pattern+"/*", relPath)
				if m {
					isIgnored = true
				}
			}
		}
	}

	if isIgnored {
		// Checa as exceções com "!"
		for _, pattern := range g.ModifyUnignorePatterns {
			if m, _ := filepath.Match(pattern, relPath); m {
				isIgnored = false // Arquivo salvo pela regra de safe-list
			} else if m, _ := filepath.Match(pattern, filename); m {
				isIgnored = false
			} else if strings.Contains(pattern, "/") {
				if strings.HasPrefix(relPath, pattern) {
					isIgnored = false
				} else {
					m, _ := filepath.Match(pattern+"/*", relPath)
					if m {
						isIgnored = false
					}
				}
			}
		}
	}

	return isIgnored
}

// IsExcludedDir verifica se um diretório deve ser ignorado na indexação
func (g *Guardian) IsExcludedDir(name string) bool {
	if name != "." && name != ".." && strings.HasPrefix(name, ".") {
		return true
	}
	return g.ExcludedDirs[name]
}

// SanitizeOutput mascara segredos no output das tools
func (g *Guardian) SanitizeOutput(content string) string {
	for _, re := range g.SecretRegexes {
		content = re.ReplaceAllString(content, "[REDACTED_SECRET]")
	}
	return content
}
