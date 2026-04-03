# Plano de Implementação: Vectora Index (Catalog & Distribution)

Este plano detalha a infraestrutura e o workflow do **Vectora Index**, o marketplace curado de datasets vetoriais para o ecossistema Vectora.

---

## 1. Visão Geral

O Vectora Index permite que usuários baixem bases de conhecimento pré-indexadas (ex: Godot 4.3 Docs, Physics Papers) e publiquem seus próprios datasets.

- **Servidor:** `index-server/` (Go, net/http).
- **Cliente:** `internal/index/client.go` (embutido no Daemon).
- **Segurança:** Curadoria manual Kaffyn e assinatura digital de payloads.

---

## 2. Arquitetura do Servidor (`index-server/`)

O servidor de índice é um serviço HTTP portável e independente do Daemon.

### 2.1 Catálogo (JSON Registry)

O registro central é um arquivo JSON assinado contendo os metadados de cada dataset:
- `id`: Unique identifier (ex: `kaffyn/godot-4.3`).
- `version`: Versionamento semântico.
- `md5`: Para verificação de integridade pós-download.
- `readme_url`: Para exibição de preview no Web UI.

### 2.2 Endpoint de Download

Os datasets são servidos em formatos binários otimizados para `chromem-go` (snapshots do banco vetorial).

---

## 3. Workflow de Download no Daemon

1. **Browse:** O Web UI solicita `index.browse` via IPC. O Daemon busca o JSON do Index Server.
2. **Download:** O usuário escolhe um dataset. O Daemon inicia o streaming para a pasta temporária do workspace.
3. **Verificação de Assinatura (RN-INDEX-01):** O Daemon valida o hash MD5/SHA256 contra o registro oficial.
4. **Hidratação:** O banco vetorial é montado e registrado como um novo Workspace local imediatamente.

---

## 4. Publicação e Curadoria

- **Submissão:** Usuários podem subir `internal/db/chroma` compactados para o Index Server.
- **Review:** A equipe Kaffyn revisa a qualidade da indexação e o conteúdo antes de tornar o dataset "público".

---

## 5. Regras de Negócio (Index)

- **RN-INDEX-01:** Nenhuma extração de dataset pode ocorrer se o hash de integridade falhar.
- **RN-INDEX-02:** O Index Server deve suportar downloads parciais (Resume) para conexões instáveis.
- **RN-INDEX-03:** A comunicação Index -> Daemon deve ser feita via HTTPS obrigatoriamente.

---

## 6. Próximos Passos (Workflow)

1.  [ ] **Implementar o Registry Manager:** Script para gerar e assinar o registro JSON.
2.  [ ] **UI do Browser:** Criar a galeria no Next.js (`internal/app`) para exibição dos datasets com cards ricos.

[Fim do Plano do Index - Revisão 2026.04.03]
