# Implementação: Kaffyn Account (Identity & SSO)

Este documento detalha o plano de desenvolvimento para a **Kaffyn Account**, a espinha dorsal de identidade e segurança de todo o ecossistema Kaffyn/Vectora. O objetivo é fornecer um SSO robusto, seguro e compatível com OIDC para unificar a experiência entre o Vectora Desktop e o Vectora Web.

## User Review Required

> [!IMPORTANT]
> **Gestão de Segredos**: Propomos o uso do **HashiCorp Vault** para o "Credential Vault". Isso exige uma infraestrutura de gerenciamento de chaves (KMS) robusta. Devemos considerar uma implementação customizada em Go usando AES-GCM como fallback para ambientes de desenvolvimento?
> **Provedor de Identidade**: Decidiremos entre usar o **Casbin** (para RBAC customizado) ou um framework completo como **Zitadel/Keycloak**. A recomendação atual é construir o core em Go Puro + Casbin para total controle sobre o fluxo de roteamento de modelos.

## Proposed Changes

### [Component Name] cmd/account

#### [NEW] [main.go](file:///c:/Users/bruno/Desktop/Vectora/account/cmd/main.go)

Ponto de entrada do serviço de identidade. Inicialização do Gin/Echo server e conexão com Postgres/Redis.

### [Component Name] internal/auth

#### [NEW] [oidc.go](file:///c:/Users/bruno/Desktop/Vectora/account/internal/auth/oidc.go)

Implementação do provedor OpenID Connect (OIDC). Gerenciamento de sub-protocolos OAuth 2.1.

#### [NEW] [rbac.go](file:///c:/Users/bruno/Desktop/Vectora/account/internal/auth/rbac.go)

Lógica de controle de acesso baseada em papéis (RBAC) com suporte a JSONB para permissões específicas de produtos.

### [Component Name] internal/vault

#### [NEW] [encryption.go](file:///c:/Users/bruno/Desktop/Vectora/account/internal/vault/encryption.go)

Camada de abstração para o Credential Vault. Integração com HashiCorp Vault ou fallback AES-GCM.

## Open Questions

> [!WARNING]
> **Sincronização Offline**: Como o Vectora Desktop funciona offline, o que acontece quando o usuário altera uma permissão de workspace ou rotaciona uma chave no Kaffyn Account enquanto o Desktop está desconectado? Precisamos de uma **Fila de Sincronização Local** no Daemon do Vectora.

## Verification Plan

### Automated Tests

- Testes de integração simulando o fluxo de "Authorization Code Grant" entre o Vectora Web e a Kaffyn Account.
- Testes de estresse no Redis para garantir que a rotação de tokens JWT suporte picos de tráfego.

### Manual Verification

- Validar o login via SSO no ambiente de desenvolvimento.
- Verificar se as chaves de API cifradas no Postgres não podem ser lidas sem a chave mestra do Vault.
