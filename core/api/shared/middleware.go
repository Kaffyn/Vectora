package shared

import (
	"vectora/core/engine"
	"vectora/core/policies"
)

// CoreDeps injeta as dependencias nativas limpas para todos os Transportes (JSON-RPC, gRPC)
type CoreDeps struct {
	Engine *engine.Engine
	Policy *policies.PolicyEngine
}
