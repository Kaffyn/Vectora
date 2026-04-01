package models

import "strings"

// SearchByTag retorna modelos que possuem uma tag específica
func SearchByTag(catalog *Catalog, tag string) []Model {
	var results []Model
	for _, model := range catalog.Models {
		for _, t := range model.Tags {
			if t == tag {
				results = append(results, model)
				break
			}
		}
	}
	return results
}

// SearchByCapability retorna modelos com uma capacidade específica
func SearchByCapability(catalog *Catalog, capability string) []Model {
	return GetModelsByCapability(catalog, capability)
}

// SearchByName retorna modelos cuja nome ou ID contém a query
func SearchByName(catalog *Catalog, query string) []Model {
	query = strings.ToLower(query)
	var results []Model
	for _, model := range catalog.Models {
		if strings.Contains(strings.ToLower(model.Name), query) ||
			strings.Contains(strings.ToLower(model.ID), query) {
			results = append(results, model)
		}
	}
	return results
}

// FilterByRAMRequirement filtra modelos que cabem na RAM disponível
func FilterByRAMRequirement(catalog *Catalog, availableRAM float64) []Model {
	var results []Model
	for _, model := range catalog.Models {
		if model.RequiredRAMGB <= availableRAM {
			results = append(results, model)
		}
	}
	return results
}

// FilterByVRAMRequirement filtra modelos que cabem na VRAM disponível
func FilterByVRAMRequirement(catalog *Catalog, availableVRAM float64) []Model {
	var results []Model
	for _, model := range catalog.Models {
		if model.RequiredVRAMGB <= availableVRAM {
			results = append(results, model)
		}
	}
	return results
}

// GetSmallestModel retorna o modelo com menor requirement de RAM
func GetSmallestModel(catalog *Catalog) *Model {
	if len(catalog.Models) == 0 {
		return nil
	}
	smallest := &catalog.Models[0]
	for i := 1; i < len(catalog.Models); i++ {
		if catalog.Models[i].RequiredRAMGB < smallest.RequiredRAMGB {
			smallest = &catalog.Models[i]
		}
	}
	return smallest
}
