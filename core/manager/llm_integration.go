package manager

import (
	"context"
	"fmt"
)

// IntegrateLLMThrottling integra rate limiting por tenant no router de LLM
// Esta função deve ser chamada ao registrar handlers de LLM
// Garante que cada tenant respeita sua cota de chamadas simultâneas
func (rp *ResourcePool) IntegrateLLMThrottling(tenant *Tenant, callback func(ctx context.Context) error) error {
	// Validar tenant
	if tenant == nil {
		return fmt.Errorf("tenant é nil")
	}

	// Adquirir slot de LLM para o tenant
	ctx, cancel := context.WithTimeout(context.Background(), 30*60*1000) // 30 minutos timeout
	defer cancel()

	if err := rp.AcquireLLMSlot(tenant.ID, ctx); err != nil {
		return fmt.Errorf("falha ao adquirir slot de LLM para tenant %s: %w", tenant.ID, err)
	}
	defer rp.ReleaseLLMSlot(tenant.ID)

	// Executar callback com contexto do tenant
	return callback(ContextWithTenant(context.Background(), tenant))
}

// RespectTenantQuota é um wrapper para qualquer operação que consuma recursos LLM
// Valida o contexto, adquire slot e executa a operação
func RespectTenantQuota(ctx context.Context, rp *ResourcePool, operation func(context.Context) error) error {
	// Extrair tenant do contexto
	tenant := TenantFromContext(ctx)
	if tenant == nil {
		return fmt.Errorf("nenhum tenant no context")
	}

	// Adquirir slot
	if err := rp.AcquireLLMSlot(tenant.ID, ctx); err != nil {
		return fmt.Errorf("rate limited - tenant ja tem chamadas simultâneas no máximo")
	}
	defer rp.ReleaseLLMSlot(tenant.ID)

	// Executar operação
	return operation(ctx)
}
