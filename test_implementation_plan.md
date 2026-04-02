# Plano de Implementação de Testes (E2E) Vectora: Zyris Engine RAG

Este documento serve como o Single Source of Truth para o novo conjunto de testes de integração contínua (End-to-End) dentro do binário original do Vectora. Abandonamos mocks em favor da verificação empírica do pipeline com dados de produção.

## 1. Topologia da Bateria de Testes (Flag `--tests`)

O `cmd/vectora` já possui a ramificação para rodar testes internos (via flag `--tests`). Modificaremos esse fluxo para suportar a ingestão e o questionamento 100% verídicos usando a Gemini API.

**O Fluxo Perfeito do Teste Integrado será:**

1. **Ativação Segura:** Inicializar o OS Manager e varrer a configuração `~/.Vectora/.env` em busca da chave do Gemini de segurança.
2. **Mount Real (Arquitetura Efêmera):** Levantar um BBolt DataStore e um Chromem-Go isolados da sessão do usuário real (num temp dir e collection dummy `ws_zyris_test`).
3. **Ingestão Vetorial:**
   - Varredura de todo o banco de arquivos situado em `./data` (Onde repousa os `.xml` do _Zyris Engine_).
   - Cortamos o texto (Chunking fixo de ~512 tokens).
   - Instanciamos o provedor Gemini via Langchaingo (Provider) e solicitamos os `Embeddings`.
   - Alimentamos o Chromem com os chunks assinados pelo Vector Math Gemini.
4. **Resolução de RAG:**
   - Disparamos a requisição ao `internal/core.Pipeline` passando `workspace.query` com perguntas como: _"Como instanciar um Singleton no Zyris Engine baseado no contrato XML?"_.
   - A resposta deve ocorrer sem falhas e seu output exibido no terminal e verificado se a fonte `Sources` bateu com os XMLs corretos.
5. **Teste de Tooling (Segurança):**
   - Na mesma bateria, tentaremos instigar a LLM para salvar essa resposta escrevendo no disco (`write_file`), e testaremos a Bridge do Git checando se o arquivo `.bak` foi alocado pela rotina.

## 2. Tarefas e Fases de Codificação

### Fase 1: Extrator Real de Dados

- Criar a rotina `vectorizeTestData` para apontar no file-system `/data`, ler todos e apenas os arquivos com a extensão `.xml` e `.txt` associados ao Zyris, realizar o split e embedar usando o provider instanciado do Gemini.

### Fase 2: O Cliente IPC Integrado

- Ao invés de invocar a Pipeline diretamente, queremos que o Teste se comporte como Frontend!
- Então o teste vai rodar `go ipcServer.Start()` no backend e invocar internamente um Socket UDP/Pipe Client simulando Wails, submetendo:
  ```json
  {
    "type": "request",
    "method": "workspace.query",
    "payload": {
      "workspace_id": "zyris_test",
      "query": "Me dê detalhes da entity Player."
    }
  }
  ```

### Fase 3: Instrumentação do `main.go`

- A função atual `runSystemIntegrityTests()` será renomeada (ou movida para pacote apartado test logic) onde executaremos: `VerifyKVStore`, `VerifyVectorStore`, `VerifyIPCAndRag` e `Shutdown()`.

## User Review Required

> [!IMPORTANT]
>
> 1. Como os testes agora englobam faturamento vivo de Tokens no Gemini (Embeddings e Completion), está ciente que rodá-los comente saldo do projeto, correto?
> 2. Você deseja que eu prepare um Mock de SafetyNet para caso falhe a rede e devolva o erro HTTP 400+, parando a build graciosamente se cair o Gemini no teste?
