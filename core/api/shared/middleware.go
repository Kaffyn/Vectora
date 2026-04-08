package shared

import (
	"github.com/Kaffyn/Vectora/core/engine"
	"github.com/Kaffyn/Vectora/core/policies"
)

// CoreDeps injeta as dependencias nativas limpas para todos os Transportes (JSON-RPC, gRPC)
type CoreDeps struct {
	Engine *engine.Engine
	Policy *policies.PolicyEngine
}
