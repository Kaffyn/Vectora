package policies

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Guardian encapsula todas as regras de segurança
type Guardian struct {
	TrustFolder   string
	BlockedExts   map[string]bool
	BlockedFiles  map[string]bool
	ExcludedDirs  map[string]bool
	SecretRegexes []*regexp.Regexp
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

	return g
}

// IsPathSafe verifica escopo e symlinks
func (g *Guardian) IsPathSafe(targetPath string) bool {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// Resolve symlinks para evitar bypass
	realPath, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		// Se não existe, ainda podemos validar o path pretendido
		realPath = absTarget
	}

	absTrust, _ := filepath.Abs(g.TrustFolder)

	// Verifica se o path real está dentro do trust folder
	return strings.HasPrefix(realPath, absTrust+string(filepath.Separator)) || realPath == absTrust
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

// IsExcludedDir verifica se um diretório deve ser ignorado na indexação
func (g *Guardian) IsExcludedDir(name string) bool {
	return g.ExcludedDirs[name]
}

// SanitizeOutput mascara segredos no output das tools
func (g *Guardian) SanitizeOutput(content string) string {
	for _, re := range g.SecretRegexes {
		content = re.ReplaceAllString(content, "[REDACTED_SECRET]")
	}
	return content
}
