package manager

import (
	"context"
	"errors"
	"fmt"
)

// contextKey é um tipo privado para evitar colisões no context.Context
type contextKey string

const tenantKey contextKey = "tenant"

// ContextWithTenant injeta o objeto Tenant no context.
func ContextWithTenant(ctx context.Context, tenant *Tenant) context.Context {
	return context.WithValue(ctx, tenantKey, tenant)
}

// TenantFromContext extrai o Tenant ativo do context.
func TenantFromContext(ctx context.Context) *Tenant {
	t := ctx.Value(tenantKey)
	if t == nil {
		return nil
	}
	return t.(*Tenant)
}

// GuardianValidate valida uma operação contra as regras do Guardian do tenant.
func GuardianValidate(ctx context.Context, filePath string) error {
	tenant := TenantFromContext(ctx)
	if tenant == nil {
		return errors.New("no tenant found in context")
	}

	// O Guardian.ValidatePath já verifica se está dentro do TrustFolder (Root)
	_, err := tenant.Guardian.ValidatePath(filePath)
	if err != nil {
		return fmt.Errorf("guardian_denied: %w", err)
	}

	return nil
}
