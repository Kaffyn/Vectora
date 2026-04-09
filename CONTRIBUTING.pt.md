# Contribuindo para o Vectora

> [!TIP]
> Read this file in another language.
> English [CONTRIBUTING.md] | Portuguese [CONTRIBUTING.pt.md]

Este documento descreve os padrões, filosofia e regras técnicas para contribuir com o Vectora. Nosso objetivo é construir um orquestrador de IA confiável, de alto desempenho e seguro.

## Filosofia do Projeto

O Vectora é construído sobre três pilares fundamentais:

1. **Confiabilidade**: O agente nunca deve deixar o sistema em um estado inconsistente. A integração com o **Git do usuário** (via IDE) deve ser utilizada para garantir o rastreamento de alterações.
2. **Segurança**: A adesão ao modelo de **Trust Folder** (Diretório de Confiança) é inegociável. As ferramentas nunca devem escapar do escopo definido.
3. **Limpeza**: Valorizamos código Go idiomático, logging estruturado e documentação profissional.

**A Política de "Sem Emojis":**

Para manter uma estética profissional e limpa, não utilizamos emojis na documentação, comentários de código ou mensagens de commit. Toda comunicação técnica deve ser clara e concisa.

## Stack Técnica e Padrões

- **Linguagem**: Go 1.23+ (utilizando iteradores e recursos modernos da biblioteca padrão).
- **Logging**: Utilize `log/slog` para todo o logging estruturado. Evite `fmt.Printf` para tudo que não seja saída direta no CLI.
- **Erros**: Sempre realize o wrapping de erros com contexto usando `fmt.Errorf("...: %w", err)`. Utilize `errors.Is` e `errors.As` para verificações.
- **Concorrência**: Utilize goroutines e canais (channels) com cautela. Siga os princípios de CSP (Communicating Sequential Processes).

## Estrutura do Repositório

O Vectora é organizado como um monorepo. O foco atual do MVP é 100% no **core** da engine.

- `cmd/vectora/`: Ponto de entrada principal utilizando Cobra CLI.
- `core/acp/`: Implementação do Agent Client Protocol (JSON-RPC 2.0).
- `core/storage/`: Camada de abstração para `bbolt` e `chromem-go`.
- `core/tools/`: O arsenal agêntico (Arquivo, Terminal, Conhecimento).
- `core/ipc/`: Implementação do transporte via Named Pipes para Windows.
- `pkg/gemini/`: Cliente otimizado para as APIs do Gemini Pro e Embedding.

## Fluxo de Trabalho para Contribuição

**1. Preparação:**

- Certifique-se de ter o Go 1.23 instalado.
- Configure seu ambiente para executar o core (o uso de named pipes requer Windows).

**2. Branching e Commits:**

- Utilize nomes de branch descritivos: `feat/nome-da-feature`, `fix/nome-do-bug`, `docs/nome-da-pagina`.
- Mantenha os commits atômicos e bem documentados.
- **Não utilize emojis em mensagens de commit.**

**3. Padrões e Testes:**

- Execute `go fmt ./...` antes de submeter.
- Escreva testes unitários para novas lógicas nos pacotes `core/`.
- Garanta que todos os métodos ACP estejam documentados e testados com mocks de JSON-RPC.

**4. Pull Requests:**

- Forneça uma descrição clara das mudanças e o que elas resolvem.
- Vincule a quaisquer issues relevantes ou discussões arquiteturais.

---

_Parte da organização open source Kaffyn._
