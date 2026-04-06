package models

import (
	"fmt"

	engines "github.com/Kaffyn/Vectora/engines"
)

// DetectHardware detecta e retorna as capacidades de hardware do sistema
func DetectHardware() (*Hardware, error) {
	// Chamar a detecção existente de engines
	engineHW, err := engines.DetectHardware()
	if err != nil {
		return nil, fmt.Errorf("hardware detection failed: %w", err)
	}

	// Mapear tipos de engines.Hardware para models.Hardware
	hw := &Hardware{
		OS:          engineHW.OS,
		Architecture: engineHW.Architecture,
		CPUFeatures: engineHW.CPUFeatures,
		CoreCount:   engineHW.CoreCount,
		RAM:         engineHW.RAM,
		GPUType:     engineHW.GPUType,
		GPUVersion:  engineHW.GPUVersion,
	}

	return hw, nil
}
