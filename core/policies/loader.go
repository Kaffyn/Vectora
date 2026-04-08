package policies

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

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

		// Note: Alguns YAMLs podem ter múltiplas regras se usarmos separadores ou lermos como slice.
		// Para simplificar o blueprint, lemos como múltiplos documentos ou um único objeto.
		// Aqui assumimos que cada arquivo pode ter múltiplas regras.
		
		// Tentativa de ler como lista
		var rule PolicyRule
		if err := yaml.Unmarshal(data, &rule); err == nil && rule.PolicyID != "" {
			l.Rules = append(l.Rules, rule)
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
