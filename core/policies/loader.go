package policies

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed rules/*.yaml
var RulesFS embed.FS

// PolicyRule representa uma regra individual carregada do YAML.
type PolicyRule struct {
	PolicyID    string      `yaml:"policy_id"`
	Rule        string      `yaml:"rule"`
	Enforcement interface{} `yaml:"enforcement"`
}

// Loader gerencia o carregamento das políticas embarcadas.
type Loader struct {
	Rules []PolicyRule
}

func NewLoader() *Loader {
	return &Loader{
		Rules: []PolicyRule{},
	}
}

// LoadAll carrega todas as regras da pasta rules/ para a memória.
func (l *Loader) LoadAll() error {
	err := fs.WalkDir(RulesFS, "rules", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		data, err := RulesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		// Use a NewDecoder to support multiple documents in the same file (separated by ---)
		// This is required to satisfy yaml-lint while allowing multiple policies per file.
		decoder := yaml.NewDecoder(strings.NewReader(string(data)))
		for {
			var rule PolicyRule
			if err := decoder.Decode(&rule); err != nil {
				if err == io.EOF {
					break
				}
				// Skip empty documents or decoding errors for specific rules
				continue
			}
			if rule.PolicyID != "" {
				l.Rules = append(l.Rules, rule)
			}
		}

		return nil
	})

	return err
}

// GetGuardian inicia um Guardian com as regras carregadas (lógica simplificada para MVP).
func (l *Loader) GetGuardian(trustFolder string) *Guardian {
	g := NewGuardian(trustFolder)
	// Futuro: injetar regras dinâmicas do YAML no Guardian
	return g
}
