A criação de uma **Kaffyn Account** (SSO Centralizado) é uma decisão estratégica brilhante que resolve problemas de usabilidade, segurança e escalabilidade para todo o ecossistema.

Abaixo está a definição detalhada das responsabilidades dessa conta, estruturada para atender tanto ao usuário final quanto à arquitetura técnica da Kaffyn.

---

### 🏛️ Arquitetura da Kaffyn Account (SSO Centralizado)

A conta não é apenas um "login", mas um **Hub de Identidade e Gestão de Credenciais**. Ela atua como uma camada de abstração entre os serviços da Kaffyn e os provedores externos (Google, Qwen, Claude, etc.).

#### 1. Gerenciamento Centralizado de Credenciais (Credential Vault)

- **Responsabilidade:** Armazenar de forma segura as chaves de API, tokens OAuth e segredos dos provedores externos (Qwen, Gemini, Google, Claude, AWS, etc.).
- **Como funciona:**
  - O usuário autentica-se **uma vez** com a Kaffyn Account.
  - A Kaffyn Account faz o "handshake" com o provedor externo (ex: Google OAuth2) e obtém o token.
  - O token é criptografado e armazenado no **Vault da Kaffyn**, nunca exposto ao código do Vectora ou outros apps.
  - Quando o Vectora precisa chamar a API do Google, ele solicita à Kaffyn Account um token temporário e revogável.
- **Benefício:** O usuário não precisa gerenciar 50 chaves de API diferentes em arquivos `.env` ou configurações locais. Se uma chave expirar, ela é renovada centralmente sem intervenção do usuário.

#### 2. Orquestração de Roteamento de Modelos (Model Router)

- **Responsabilidade:** Decidir qual LLM usar baseado nas permissões e créditos vinculados à conta.
- **Como funciona:**
  - A conta define quais modelos estão disponíveis (ex: "Plano Pro tem acesso ao Qwen3-80B", "Plano Free só tem Qwen3-7B").
  - O Vectora App (local) consulta a conta para saber se pode usar o modelo local ou se deve rotear para a nuvem da Kaffyn (se configurado).
  - No **Vectora Web**, a conta decide automaticamente o roteamento (ex: "Usar Qwen via API interna da Kaffyn" vs "Usar Gemini direto do usuário").

#### 3. Sistema de RBAC Granular por Produto (Product-Specific Access)

- **Responsabilidade:** Garantir o princípio de **Menor Privilégio** entre os produtos da Kaffyn.
- **Lógica de Acesso:**
  - A conta possui permissões globais (ex: "Pode ler meu Google Drive").
  - Cada produto (Vectora, Vectora Web, futuro "Kaffyn Notes", etc.) declara suas necessidades na hora da conexão.
  - **Exemplo Prático:**
    - **Vectora:** Solicita acesso `read:docs` (para indexar PDFs). ✅ Permitido.
    - **Vectora Web:** Solicita acesso `read:docs` + `write:chat_history`. ✅ Permitido.
    - **Futuro App "Kaffyn Calendar":** Solicita acesso `read:calendar`. ❌ **Negado** (o usuário nunca autorizou isso para este app específico, mesmo estando logado).
- **Implementação Técnica:**
  - Uso de **OAuth Scopes** específicos por aplicação.
  - Um painel de controle onde o usuário vê: _"Quais apps têm acesso aos seus dados?"_.

#### 4. Sincronização de Estado e Workspaces (Cross-Device Sync)

- **Responsabilidade:** Manter a consistência entre o **Vectora App (Desktop)** e o **Vectora Web**.
- **Como funciona:**
  - Metadados dos workspaces (nomes, configurações, status de indexação) são sincronizados via conta.
  - **Diferença Crítica:** Os _dados brutos_ (vetores, embeddings) permanecem locais no Desktop (privacidade total) ou na VPS da Kaffyn (Web), mas a **estrutura lógica** (qual workspace está ativo, histórico de chat resumido) é sincronizada.
  - Isso permite começar um chat no desktop e continuar no navegador sem perder o contexto da sessão.

#### 5. Gestão de Assinaturas e Limites (Billing & Quotas)

- **Responsabilidade:** Controlar limites de uso baseados no plano da conta.
- **Funcionalidades:**
  - Monitorar cota de processamento (tokens/queries) para o Vectora Web.
  - Gerenciar licenças de modelos premium (ex: Qwen3-Coder-Next 80B).
  - Bloquear acesso a recursos avançados se a assinatura expirar.

---

### 🔒 Fluxo de Autenticação (Exemplo de Uso)

1.  **Login:** Usuário acessa o **Vectora App** ou **Vectora Web**.
2.  **Redirecionamento:** Redirecionado para `auth.kaffyn.com` (SSO).
3.  **Autenticação:** Login único (Email/Google/Microsoft).
4.  **Consentimento:**
    - _Prompt:_ "O Vectora deseja acessar seu Google Drive para indexar documentos."
    - _Usuário:_ "Permitir".
    - _Token:_ Gerado e vinculado à conta.
5.  **Uso:**
    - O Vectora App (Go) pede ao daemon local: "Preciso de docs do Drive".
    - Daemon consulta a Kaffyn Account (via IPC seguro ou token JWT).
    - Token é usado para buscar dados.
6.  **Isolamento:**
    - Se o usuário abrir o **futuro app "Kaffyn Finance"**, ele verá um prompt diferente: "Este app quer acessar sua conta bancária".
    - Mesmo logado na mesma conta, o Financeiro **não terá** acesso automático aos dados do Drive que foram dados ao Vectora.

---

### 📋 Resumo das Responsabilidades da Kaffyn Account

| Área              | Responsabilidade Principal                                                 | Benefício para o Usuário                                            |
| :---------------- | :------------------------------------------------------------------------- | :------------------------------------------------------------------ |
| **Segurança**     | Armazenamento criptografado de todas as Chaves/API Keys externas.          | Não precisa salvar chaves no PC; renovação automática.              |
| **Privacidade**   | Controle granular de escopo (RBAC) por produto.                            | Saber exatamente quem tem acesso a quê; evitar vazamentos cruzados. |
| **UX**            | Single Sign-On (SSO) para todos os produtos Kaffyn.                        | Login único; experiência fluida entre Desktop e Web.                |
| **Gestão**        | Centralização de credenciais de múltiplos provedores (Qwen, Google, etc.). | Foco no trabalho, não na configuração de APIs.                      |
| **Sincronização** | Sincronização de metadados e estado entre dispositivos.                    | Continuidade de fluxo de trabalho (Desktop ↔ Web).                  |
| **Comercial**     | Gestão de planos, limites e faturamento unificado.                         | Uma única fatura para todos os produtos; upgrade fácil.             |

Esta estrutura transforma a **Kaffyn Account** no **núcleo de confiança** do ecossistema, permitindo que o Vectora mantenha sua promessa de privacidade local enquanto oferece a conveniência de um serviço conectado e moderno.
