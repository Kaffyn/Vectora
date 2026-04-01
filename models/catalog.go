package models

import (
	"encoding/json"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed catalog.json
var catalogData []byte

// LoadCatalog carrega o catálogo embarcado de modelos
func LoadCatalog() (*Catalog, error) {
	var cat Catalog
	if err := json.Unmarshal(catalogData, &cat); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %w", err)
	}
	return &cat, nil
}

// FindModel procura um modelo por ID no catálogo
func FindModel(catalog *Catalog, id string) (*Model, error) {
	for i, model := range catalog.Models {
		if model.ID == id {
			return &catalog.Models[i], nil
		}
	}
	return nil, fmt.Errorf("model '%s' not found in catalog", id)
}

// SearchModels busca modelos por query de texto (nome/descrição)
func SearchModels(catalog *Catalog, query string) []Model {
	query = strings.ToLower(query)
	var results []Model
	for _, model := range catalog.Models {
		if strings.Contains(strings.ToLower(model.Name), query) ||
			strings.Contains(strings.ToLower(model.Description), query) ||
			strings.Contains(strings.ToLower(model.ID), query) {
			results = append(results, model)
		}
	}
	return results
}

// GetModelsByCapability retorna modelos que possuem uma capacidade específica
func GetModelsByCapability(catalog *Catalog, capability string) []Model {
	var results []Model
	for _, model := range catalog.Models {
		for _, cap := range model.Capabilities {
			if cap == capability {
				results = append(results, model)
				break
			}
		}
	}
	return results
}
