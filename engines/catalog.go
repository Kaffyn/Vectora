package engines

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// Catálogo embarcado contém o JSON com todas as builds disponíveis.
//
//go:embed catalog.json
var catalogData []byte

var (
	// CachedCatalog é o catálogo parseado uma única vez na inicialização.
	CachedCatalog *Catalog
	ErrCatalogNotLoaded = fmt.Errorf("catalog not loaded")
)

// LoadCatalog carrega o catálogo embarcado na memória.
// Deve ser chamado uma única vez na inicialização do daemon.
func LoadCatalog() error {
	catalog := &Catalog{}
	if err := json.Unmarshal(catalogData, catalog); err != nil {
		return fmt.Errorf("failed to parse catalog: %w", err)
	}
	CachedCatalog = catalog
	return nil
}

// GetCatalog retorna o catálogo global (já carregado).
func GetCatalog() (*Catalog, error) {
	if CachedCatalog == nil {
		return nil, ErrCatalogNotLoaded
	}
	return CachedCatalog, nil
}

// FindBuild procura por um build específico no catálogo.
func (c *Catalog) FindBuild(buildID string) (*Build, error) {
	for i := range c.Builds {
		if c.Builds[i].ID == buildID {
			return &c.Builds[i], nil
		}
	}
	return nil, fmt.Errorf("build %q not found in catalog", buildID)
}

// FindBestBuild retorna o build mais adequado para o hardware fornecido.
// Estratégia:
//   1. Match exato: OS + Arch + GPU
//   2. Match: OS + Arch (sem GPU)
//   3. CPU fallback: Qualquer CPU-only para o OS
func (c *Catalog) FindBestBuild(hw *Hardware) (*Build, error) {
	if len(c.Builds) == 0 {
		return nil, fmt.Errorf("catalog has no builds")
	}

	// Passo 1: Match exato com GPU
	for i := range c.Builds {
		b := &c.Builds[i]
		if b.OS == hw.OS &&
			b.Architecture == hw.Architecture &&
			b.GPU == hw.GPUType &&
			b.GPUVersion == hw.GPUVersion {
			return b, nil
		}
	}

	// Passo 2: Match OS + Arch + GPU (ignorar versão exata do GPU)
	for i := range c.Builds {
		b := &c.Builds[i]
		if b.OS == hw.OS &&
			b.Architecture == hw.Architecture &&
			b.GPU == hw.GPUType {
			return b, nil
		}
	}

	// Passo 3: Match OS + Arch (sem GPU, CPU-only)
	for i := range c.Builds {
		b := &c.Builds[i]
		if b.OS == hw.OS &&
			b.Architecture == hw.Architecture &&
			b.GPU == "none" {
			return b, nil
		}
	}

	// Passo 4: Fallback: Qualquer CPU-only para o OS
	for i := range c.Builds {
		b := &c.Builds[i]
		if b.OS == hw.OS && b.GPU == "none" {
			return b, nil
		}
	}

	return nil, fmt.Errorf("no suitable build found for %s-%s (gpu: %s)", hw.OS, hw.Architecture, hw.GPUType)
}

// ListBuilds retorna todos os builds do catálogo que atendem critérios opcionais.
func (c *Catalog) ListBuilds(osFilter, gpuFilter string) []Build {
	var result []Build
	for i := range c.Builds {
		b := &c.Builds[i]
		if osFilter != "" && b.OS != osFilter {
			continue
		}
		if gpuFilter != "" && b.GPU != gpuFilter {
			continue
		}
		result = append(result, *b)
	}
	return result
}

// GetBuildByID é um helper para obter um build específico pelo ID.
func GetBuildByID(buildID string) (*Build, error) {
	catalog, err := GetCatalog()
	if err != nil {
		return nil, err
	}
	return catalog.FindBuild(buildID)
}

// RecommendedBuild retorna a build mais adequada para o hardware atual.
func RecommendedBuild(hw *Hardware) (*Build, error) {
	catalog, err := GetCatalog()
	if err != nil {
		return nil, err
	}
	return catalog.FindBestBuild(hw)
}
