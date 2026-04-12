# 🔧 TurboQuant Integration: Storage Layer Compression (Dream Phase Addendum)

## Contexto: Por Que Comprimir o Storage?

No MVP, os embeddings são armazenados como `float32` (4 bytes por dimensão). Para um vetor de 1536 dimensões (OpenAI/Gemini), isso resulta em **~6KB por chunk**. Em um workspace com 100k chunks, são **~600MB apenas de vetores** — sem contar metadata, índices e overhead do Chromem-go.

**TurboQuant** (Google Research, 2025) demonstra que é possível comprimir representações vetoriais em até **90%** com perda mínima de precisão semântica, usando técnicas como:

- **Quantização não-uniforme** (não apenas int8, mas códigos de livro adaptativos)
- **Compressão esparsa** (remover dimensões redundantes via PCA/Autoencoder leve)
- **Encoding diferencial** (armazenar apenas deltas entre vetores similares)

## Objetivo da Integração

Aplicar TurboQuant **não apenas nos LLMs via llama.cpp**, mas também:

1. **No Chromem-go (Vector DB):** Comprimir embeddings antes de persistir em disco/RAM.
2. **No BBolt (KV Store):** Comprimir metadata, histórico de chat e caches de contexto.
3. **No IPC/MCP:** Reduzir payload de transferência entre Core e UIs/agentes externos.

**Resultado Esperado:** Workspaces 5-10x menores, consultas mais rápidas (menos I/O), e capacidade de rodar RAG em hardware modesto (8GB RAM) sem sacrificar qualidade.

---

## Arquitetura de Implementação

### 1. Camada de Compressão Abstrata (`core/storage/compress/`)

Criar uma interface genérica para compressão de vetores e metadata:

```go
// core/storage/compress/compressor.go
package compress

type VectorCompressor interface {
    // Compress reduz um vetor float32 para representação compacta
    Compress(vec []float32) ([]byte, error)
    // Decompress restaura o vetor original (ou aproximação)
    Decompress(data []byte) ([]float32, error)
    // Similarity calcula similaridade direta entre vetores comprimidos (opcional, para performance)
    Similarity(a, b []byte) (float32, error)
    // Stats retorna métricas de compressão (ratio, perda estimada)
    Stats() CompressionStats
}

type MetadataCompressor interface {
    CompressKV(key []byte, value []byte) ([]byte, error)
    DecompressKV(data []byte) (key, value []byte, err error)
}
```

### 2. Implementação TurboQuant para Chromem-go

**Opção A: Wrapper Adapter (Recomendado Inicialmente)**
Não forkar o chromem-go imediatamente. Criar um adapter que intercepta operações de escrita/leitura:

```go
// core/storage/chromem_quantized.go
type QuantizedVectorStore struct {
    base     *chromem.DB
    compressor compress.VectorCompressor
    mu       sync.RWMutex
}

func (q *QuantizedVectorStore) Add(ctx context.Context, collection, id, content string, vector []float32, metadata map[string]any) error {
    // 1. Comprimir vetor antes de armazenar
    compressed, err := q.compressor.Compress(vector)
    if err != nil { return err }

    // 2. Armazenar metadata de compressão junto
    metaWithStats := map[string]any{
        "_compressed": true,
        "_original_dims": len(vector),
        "_compression_ratio": float32(len(compressed)) / float32(len(vector)*4),
    }
    for k, v := range metadata { metaWithStats[k] = v }

    // 3. Delegar ao chromem-go base (que agora armazena []byte em vez de []float32)
    return q.base.Add(ctx, collection, id, content, compressed, metaWithStats)
}

func (q *QuantizedVectorStore) Query(ctx context.Context, collection string, queryVector []float32, topK int) ([]chromem.SearchResult, error) {
    // 1. Comprimir query vector
    qCompressed, _ := q.compressor.Compress(queryVector)

    // 2. Buscar no base (que retorna vetores comprimidos)
    results, err := q.base.Query(ctx, collection, qCompressed, topK)

    // 3. Opcional: decomprimir resultados para compatibilidade com APIs externas
    for i := range results {
        if vec, ok := results[i].Vector.([]byte); ok {
            results[i].Vector, _ = q.compressor.Decompress(vec)
        }
    }
    return results, err
}
```

**Opção B: Fork do Chromem-go (Fase Avançada)**
Se a performance do wrapper for insuficiente, forkar o chromem-go para:

- Alterar o tipo interno de `[]float32` para `[]byte` com suporte nativo a compressão.
- Implementar **similarity search direto em espaço comprimido** (evitando decompressão em cada comparação).
- Adicionar índices quantizados (IVF-PQ style) para busca ainda mais rápida.

### 3. Algoritmos de Compressão Suportados

Implementar múltiplos backends para o `VectorCompressor`, selecionáveis por config:

| Algoritmo            | Tipo                   | Ratio Estimado | Perda de Precisão | Uso Recomendado             |
| -------------------- | ---------------------- | -------------- | ----------------- | --------------------------- |
| `float32` (baseline) | Nenhum                 | 1.0x           | 0%                | Debug, máxima precisão      |
| `int8` (linear)      | Quantização simples    | 4.0x           | ~1-2% MRR@10      | Workspaces gerais           |
| `turboquant-v1`      | Não-uniform + sparse   | 8-10x          | ~3-5% MRR@10      | Produção, hardware modesto  |
| `turboquant-v2`      | + encoding diferencial | 12-15x         | ~5-8% MRR@10      | Edge devices, mobile        |
| `lossless-zstd`      | Compressão sem perda   | 2-3x           | 0%                | Metadata, histórico crítico |

**Configuração por Workspace:**

```json
{
  "workspace_id": "my-project",
  "embedding_compression": {
    "algorithm": "turboquant-v1",
    "target_ratio": 10.0,
    "max_precision_loss": 0.05
  }
}
```

### 4. Integração com BBolt (KV Store)

Para metadata e histórico, usar compressão leve com **zstd** ou **snappy**:

```go
// core/storage/kv_compressed.go
type CompressedKVStore struct {
    base *bbolt.DB
    compressor compress.MetadataCompressor
}

func (c *CompressedKVStore) Set(ctx context.Context, bucket, key string, value []byte) error {
    compressed, err := c.compressor.CompressKV([]byte(key), value)
    if err != nil { return err }
    return c.base.Set(ctx, bucket, key, compressed) // key original, value comprimido
}
```

**Benefício:** Histórico de chat de 100MB pode cair para ~20-30MB sem perda de informação.

### 5. Impacto no IPC/MCP

Quando vetores ou resultados grandes trafegam via IPC/MCP, comprimir antes de serializar:

```go
// core/api/ipc/compression.go
func CompressResponse(resp any) ([]byte, error) {
    // Detectar se a resposta contém vetores grandes
    if hasLargeVectors(resp) {
        compressed := turboquant.CompressVectors(resp.Vectors)
        resp.Vectors = compressed // substituir por versão compacta
        resp._compressed = true // flag para o cliente descomprimir se necessário
    }
    return json.Marshal(resp)
}
```

**Cliente (VS Code Extension, Claude Code):** Se receber `_compressed: true`, descomprime localmente antes de usar.

---

## Trade-offs e Mitigações

| Trade-off                                    | Mitigação                                                                                      |
| -------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| **Perda de precisão semântica**              | Configuração por workspace; permitir "re-embed" sem compressão para chunks críticos            |
| **Overhead de CPU na compressão**            | Compressão assíncrona em background; cache de vetores frequentes em RAM não-comprimida         |
| **Compatibilidade com ferramentas externas** | API de query sempre retorna vetores descomprimidos por padrão; compressão é transparente       |
| **Complexidade de debug**                    | Logs detalhados de ratio/perda; comando `vectora workspace stats --compression` para auditoria |

---

## Roadmap de Implementação

### Phase Dream-1: Foundation (Q2 2026)

- [ ] Interface `VectorCompressor` + backend `int8` simples
- [ ] Wrapper adapter para Chromem-go (Opção A)
- [ ] Configuração por workspace via `vectora config set workspace.<id>.compression`

### Phase Dream-2: TurboQuant Core (Q3 2026)

- [ ] Implementação do algoritmo `turboquant-v1` (baseado no paper arXiv:2502.02617)
- [ ] Similarity search direto em espaço comprimido (evitar decompressão em query)
- [ ] Benchmarks de precisão vs. ratio em datasets reais (MTEB)

### Phase Dream-3: Otimizações Avançadas (Q4 2026)

- [ ] Fork do Chromem-go with suporte nativo a vetores comprimidos
- [ ] Índices quantizados (IVF-PQ) para busca 10x mais rápida
- [ ] Compressão diferencial entre chunks similares do mesmo arquivo

### Phase Dream-4: Edge Optimization (2027)

- [ ] Suporte a `turboquant-v2` para mobile/embedded
- [ ] Compressão seletiva: apenas chunks antigos, manter recentes em alta precisão
- [ ] Auto-tuning: ajustar ratio dinamicamente baseado em RAM disponível

---

## Comandos de Usuário (CLI)

```bash
# Ver stats de compressão de um workspace
vectora workspace stats --id my-project --compression

# Alterar algoritmo de compressão
vectora config set workspace.my-project.compression.algorithm turboquant-v1

# Re-embed um workspace com nova compressão (background job)
vectora workspace re-embed --id my-project --compression turboquant-v1

# Comparar precisão antes/depois da compressão
vectora workspace eval --id my-project --metric mrr@10 --compression-compare
```

---

## Referências Técnicas

- [TurboQuant: Redefining AI Efficiency with Extreme Compression](https://research.google/blog/turboquant-redefining-ai-efficiency-with-extreme-compression/)
- [TurboQuant Paper (arXiv 2502.02617)](https://arxiv.org/abs/2502.02617)
- [TurboQuant Update (arXiv 2504.19874)](https://arxiv.org/abs/2504.19874)
- [Chromem-go Repository](https://github.com/philippgille/chromem-go)
- [BBolt Documentation](https://github.com/etcd-io/bbolt)

---

> [!IMPORTANT]
> **Transparência ao Usuário:** O Vectora sempre informará quando a compressão estiver ativa e qual a perda estimada de precisão. Workspaces críticos (ex: código de produção) podem optar por desativar compressão sem penalidade funcional — apenas com maior uso de disco/RAM.
