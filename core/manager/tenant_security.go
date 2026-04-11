package manager

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// TenantContextKey é a chave utilizada para armazenar Tenant no context
type TenantContextKey string

const tenantContextKey TenantContextKey = "tenant"

// TenantSecurityInterceptor valida acesso a recursos dentro de limites do workspace
// Previne path traversal e acesso cross-tenant
type TenantSecurityInterceptor struct {
	tenantID   string
	tenantRoot string
}

// NewTenantSecurityInterceptor cria um novo interceptor de segurança
func NewTenantSecurityInterceptor(tenantID, tenantRoot string) *TenantSecurityInterceptor {
	return &TenantSecurityInterceptor{
		tenantID:   tenantID,
		tenantRoot: tenantRoot,
	}
}

// ValidateFilePath verifica se um caminho está dentro do workspace do tenant
// Retorna o caminho absoluto se válido, erro se tentar sair do workspace
func (tsi *TenantSecurityInterceptor) ValidateFilePath(reqPath string) (string, error) {
	if reqPath == "" {
		return "", errors.New("path não pode estar vazio")
	}

	// Resolver path absoluto
	absPath, err := filepath.Abs(reqPath)
	if err != nil {
		return "", fmt.Errorf("caminho inválido: %w", err)
	}

	absTenantRoot, err := filepath.Abs(tsi.tenantRoot)
	if err != nil {
		return "", fmt.Errorf("erro ao resolver tenant root: %w", err)
	}

	// Normalizar paths para comparação
	absPath = filepath.ToSlash(absPath)
	absTenantRoot = filepath.ToSlash(absTenantRoot)

	// Garantir que tenant root termina com /
	if !strings.HasSuffix(absTenantRoot, "/") {
		absTenantRoot += "/"
	}

	// Verificar se path está dentro do workspace
	if !strings.HasPrefix(absPath, absTenantRoot) && absPath != strings.TrimSuffix(absTenantRoot, "/") {
		return "", fmt.Errorf("tentativa de acesso fora do workspace: %s (limite: %s)", absPath, tsi.tenantRoot)
	}

	return filepath.FromSlash(absPath), nil
}

// ContextWithTenant injeta um tenant no context
// Usado para passar informações do tenant através de chamadas de função
func ContextWithTenant(ctx context.Context, tenant *Tenant) context.Context {
	return context.WithValue(ctx, tenantContextKey, tenant)
}

// TenantFromContext extrai um tenant do context
// Retorna nil se nenhum tenant estiver no context
func TenantFromContext(ctx context.Context) *Tenant {
	t := ctx.Value(tenantContextKey)
	if t == nil {
		return nil
	}
	return t.(*Tenant)
}

// ValidateTenantContext verifica se um context contém um tenant válido
// Retorna erro se tenant estiver ausente ou inválido
func ValidateTenantContext(ctx context.Context) error {
	tenant := TenantFromContext(ctx)
	if tenant == nil {
		return errors.New("nenhum tenant no context")
	}
	return nil
}

// GuardianValidate valida uma operação de arquivo contra as regras Guardian do tenant
// Consulta o Guardian do tenant para verificar se a operação é permitida
func GuardianValidate(ctx context.Context, filePath string) error {
	tenant := TenantFromContext(ctx)
	if tenant == nil {
		return errors.New("tenant não encontrado no context")
	}

	if tenant.Guardian == nil {
		return errors.New("Guardian não inicializado para este tenant")
	}

	// Validar que o path está dentro do workspace
	if !tenant.Guardian.IsPathSafe(filePath) {
		return fmt.Errorf("acesso negado pelo Guardian - path fora do workspace ou não seguro: %s", filePath)
	}

	return nil
}

// EnforcePathBoundary verifica se um path está dentro do limite do workspace
// Função auxiliar para uso em handlers que precisam validar paths
func EnforcePathBoundary(ctx context.Context, filePath string) (string, error) {
	tenant := TenantFromContext(ctx)
	if tenant == nil {
		return "", errors.New("tenant não encontrado no context")
	}

	// Usar Guardian para validar
	validPath, err := tenant.Guardian.ValidatePath(filePath)
	if err != nil {
		return "", err
	}

	return validPath, nil
}
