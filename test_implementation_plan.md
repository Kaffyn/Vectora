# Plano de Implementação: Testes de Integridade (End-to-End)

Este plano descreve a estratégia de testes sistêmicos para garantir a confiabilidade do Vectora, focando na flag `--tests` no binário principal.

---

## 1. Visão Geral (Audit de Integridade)

O Vectora rejeita testes superficiais. O objetivo desta suite é simular o uso real do sistema sob condições adversas para validar o RAG, a Gestão de Memória e as Ferramentas.

- **Fisofia:** Padrão 300% (Happy Path, Negative, Edge Case).
- **Ambiente:** Isola o bbolt e chromem-go em uma pasta temporária `/test_data` para evitar poluição.

---

## 2. A Suite "Zyris Engine" (O Teste Real)

Para validar o RAG local e as Ferramentas, o sistema usará o **Zyris Engine** (uma base de arquivos estática complexa) como benchmark:

1. **Setup:** Criar um workspace temporário, ingerir arquivos `.xml` e `.txt` complexos (cerca de 500 chunks).
2. **Indexing:** Iniciar a indexação via IPC e aguardar o evento `workspace.indexed`.
3. **Query (RAG Test):** Fazer 5 perguntas difíceis sobre a arquitetura Zyris (ex: "Como funciona o buffer de áudio na ZyrisRAG?").
4. **Tool Test:** Ordenar que o Agente ACP leia o arquivo, edite uma nota e reporte o snapshot do GitBridge.
5. **Undo Test:** Verificar se o `undo` restaurou o arquivo original byte-a-byte.

---

## 3. Testes de Stress e Concorrência

- **IPC Latency:** Disparar 100 requisições simultâneas ao socket e validar a fila de resposta.
- **Race conditions:** Rodar `go test -race` em todas as rotas do `internal/ipc`.
- **Memory Leak:** Monitorar se o Daemon libera a memória após fechar 10 workspaces pesados.

---

## 4. O Comitê de Auditoria (Flag `--tests`)

A implementação no `main.go` deve seguir:

```go
func runSystemIntegrityTests() {
    // 1. Iniciar Daemon em modo isolado
    // 2. Rodar suite Core RAG
    // 3. Rodar suite Tool Snapshots
    // 4. Rodar suite IPC Protocol
    // 5. Encerrar e limpar /test_data
}
```

---

## 5. Regras de Aceite dos Testes

- **RN-TEST-01:** Nenhum teste pode falhar por problemas de rede (devem usar mocks ou modo offline puro).
- **RN-TEST-02:** O tempo total de auditoria não deve passar de 3 minutos.
- **RN-TEST-03:** Em caso de falha, o log detalhado deve ser salvo em `tests/logs/latest_fail.txt`.

---

## 6. Próximos Passos (Workflow de QA)

1.  [ ] **Implementar Mocks:** Criar mocks para o Gemini API quando a rede estiver indisponível.
2.  [ ] **Assets de Teste:** Inserir os arquivos de benchmark na pasta `tests/data/zyris`.
3.  [ ] **CI/CD Sync:** Integrar a flag `--tests` no workflow do Github Actions.

[Fim do Plano de Testes - Revisão 2026.04.03]
