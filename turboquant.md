# 🔧 TurboQuant Integration: Storage Layer Compression (Dream Phase Addendum)

## Context: Why Compress Storage?

In the MVP, embeddings are stored as `float32` (4 bytes per dimension). For a 1536-dimensional vector (OpenAI/Gemini), this results in **~6KB per chunk**. In a workspace with 100k chunks, that's **~600MB of vectors alone** — excluding metadata, indices, and Chromem-go overhead.

**TurboQuant** (Google Research, 2025) demonstrates that it is possible to compress vector representations by up to **90%** with minimal semantic precision loss, using techniques such as:

- **Non-uniform quantization** (not just int8, but adaptive codebooks)
- **Sparse compression** (removing redundant dimensions via PCA/lightweight Autoencoder)
- **Differential encoding** (storing only deltas between similar vectors)

## Integration Objective

Apply TurboQuant **not just in LLMs via llama.cpp**, but also:

1. **In Chromem-go (Vector DB):** Compress embeddings before persisting to disk/RAM.
2. **In BBolt (KV Store):** Compress metadata, chat history, and context caches.
3. **In IPC/MCP:** Reduce transfer payload between Core and UIs/external agents.

**Expected Result:** Workspaces 5-10x smaller, faster queries (less I/O), and the ability to run RAG on modest hardware (8GB RAM) without sacrificing quality.

---

## Implementation Architecture

### 1. Abstract Compression Layer (`core/storage/compress/`)

Create a generic interface for vector and metadata compression:

```go
// core/storage/compress/compressor.go
package compress

type VectorCompressor interface {
    // Compress reduces a float32 vector to a compact representation
    Compress(vec []float32) ([]byte, error)
    // Decompress restores the original vector (or an approximation)
    Decompress(data []byte) ([]float32, error)
    // Similarity calculates direct similarity between compressed vectors (optional, for performance)
    Similarity(a, b []byte) (float32, error)
    // Stats returns compression metrics (ratio, estimated loss)
    Stats() CompressionStats
}

type MetadataCompressor interface {
    CompressKV(key []byte, value []byte) ([]byte, error)
    DecompressKV(data []byte) (key, value []byte, err error)
}
```

### 2. TurboQuant Implementation for Chromem-go

**Option A: Wrapper Adapter (Initially Recommended)**
Do not fork chromem-go immediately. Create an adapter that intercepts write/read operations:

```go
// core/storage/chromem_quantized.go
type QuantizedVectorStore struct {
    base     *chromem.DB
    compressor compress.VectorCompressor
    mu       sync.RWMutex
}

func (q *QuantizedVectorStore) Add(ctx context.Context, collection, id, content string, vector []float32, metadata map[string]any) error {
    // 1. Compress vector before storing
    compressed, err := q.compressor.Compress(vector)
    if err != nil { return err }

    // 2. Store compression metadata alongside
    metaWithStats := map[string]any{
        "_compressed": true,
        "_original_dims": len(vector),
        "_compression_ratio": float32(len(compressed)) / float32(len(vector)*4),
    }
    for k, v := range metadata { metaWithStats[k] = v }

    // 3. Delegate to base chromem-go (which now stores []byte instead of []float32)
    return q.base.Add(ctx, collection, id, content, compressed, metaWithStats)
}

func (q *QuantizedVectorStore) Query(ctx context.Context, collection string, queryVector []float32, topK int) ([]chromem.SearchResult, error) {
    // 1. Compress query vector
    qCompressed, _ := q.compressor.Compress(queryVector)

    // 2. Search in base (which returns compressed vectors)
    results, err := q.base.Query(ctx, collection, qCompressed, topK)

    // 3. Optional: decompress results for compatibility with external APIs
    for i := range results {
        if vec, ok := results[i].Vector.([]byte); ok {
            results[i].Vector, _ = q.compressor.Decompress(vec)
        }
    }
    return results, err
}
```

**Option B: Chromem-go Fork (Advanced Phase)**
If wrapper performance is insufficient, fork chromem-go to:

- Change the internal type from `[]float32` to `[]byte` with native compression support.
- Implement **similarity search directly in compressed space** (avoiding decompression during each comparison).
- Add quantized indices (IVF-PQ style) for even faster searching.

### 3. Supported Compression Algorithms

Implement multiple backends for `VectorCompressor`, selectable via config:

| Algorithm            | Type                    | Estimated Ratio | Precision Loss | Recommended Use             |
| -------------------- | ----------------------- | --------------- | -------------- | --------------------------- |
| `float32` (baseline) | None                    | 1.0x            | 0%             | Debug, maximum precision    |
| `int8` (linear)      | Simple quantization     | 4.0x            | ~1-2% MRR@10   | General workspaces          |
| `turboquant-v1`      | Non-uniform + sparse    | 8-10x           | ~3-5% MRR@10   | Production, modest hardware |
| `turboquant-v2`      | + differential encoding | 12-15x          | ~5-8% MRR@10   | Edge devices, mobile        |
| `lossless-zstd`      | Lossless compression    | 2-3x            | 0%             | Metadata, critical history  |

**Config per Workspace:**

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

### 4. BBolt Integration (KV Store)

For metadata and history, use lightweight compression with **zstd** or **snappy**:

```go
// core/storage/kv_compressed.go
type CompressedKVStore struct {
    base *bbolt.DB
    compressor compress.MetadataCompressor
}

func (c *CompressedKVStore) Set(ctx context.Context, bucket, key string, value []byte) error {
    compressed, err := c.compressor.CompressKV([]byte(key), value)
    if err != nil { return err }
    return c.base.Set(ctx, bucket, key, compressed) // original key, compressed value
}
```

**Benefit:** A 100MB chat history can drop to ~20-30MB without information loss.

### 5. IPC/MCP Impact

When large vectors or results travel via IPC/MCP, compress before serializing:

```go
// core/api/ipc/compression.go
func CompressResponse(resp any) ([]byte, error) {
    // Detect if the response contains large vectors
    if hasLargeVectors(resp) {
        compressed := turboquant.CompressVectors(resp.Vectors)
        resp.Vectors = compressed // replace with compact version
        resp._compressed = true // flag for the client to decompress if necessary
    }
    return json.Marshal(resp)
}
```

**Client (VS Code Extension, Claude Code):** If it receives `_compressed: true`, it decompresses locally before use.

---

## Trade-offs and Mitigations

| Trade-off                       | Mitigation                                                                             |
| ------------------------------- | -------------------------------------------------------------------------------------- |
| **Semantic precision loss**     | Workspace-specific config; allow "re-embed" without compression for critical chunks    |
| **CPU overhead in compression** | Asynchronous background compression; cache frequent vectors in non-compressed RAM      |
| **Outward tool compatibility**  | Query API always returns decompressed vectors by default; compression is transparent   |
| **Debug complexity**            | Detailed ratio/loss logs; `vectora workspace stats --compression` command for auditing |

---

## Implementation Roadmap

### Phase Dream-1: Foundation (Q2 2026)

- [ ] `VectorCompressor` interface + simple `int8` backend
- [ ] Wrapper adapter for Chromem-go (Option A)
- [ ] Workspace-specific config via `vectora config set workspace.<id>.compression`

### Phase Dream-2: TurboQuant Core (Q3 2026)

- [ ] Implementation of the `turboquant-v1` algorithm (based on arXiv:2502.02617)
- [ ] Direct similarity search in compressed space (avoid decompression during query)
- [ ] Precision vs. ratio benchmarks on real datasets (MTEB)

### Phase Dream-3: Advanced Optimizations (Q4 2026)

- [ ] Chromem-go fork with native support for compressed vectors
- [ ] Quantized indices (IVF-PQ) for 10x faster search
- [ ] Differential compression between similar chunks of the same file

### Phase Dream-4: Edge Optimization (2027)

- [ ] Support for `turboquant-v2` for mobile/embedded
- [ ] Selective compression: only old chunks, keep recent ones in high precision
- [ ] Auto-tuning: adjust ratio dynamically based on available RAM

---

## User Commands (CLI)

```bash
# View compression stats for a workspace
vectora workspace stats --id my-project --compression

# Change compression algorithm
vectora config set workspace.my-project.compression.algorithm turboquant-v1

# Re-embed a workspace with new compression (background job)
vectora workspace re-embed --id my-project --compression turboquant-v1

# Compare precision before/after compression
vectora workspace eval --id my-project --metric mrr@10 --compression-compare
```

---

## Technical References

- [TurboQuant: Redefining AI Efficiency with Extreme Compression](https://research.google/blog/turboquant-redefining-ai-efficiency-with-extreme-compression/)
- [TurboQuant Paper (arXiv 2502.02617)](https://research.google/blog/turboquant-redefining-ai-efficiency-with-extreme-compression/)
- [TurboQuant Update (arXiv 2504.19874)](https://research.google/blog/turboquant-redefining-ai-efficiency-with-extreme-compression/)
- [Chromem-go Repository](https://github.com/philippgille/chromem-go)
- [BBolt Documentation](https://github.com/etcd-io/bbolt)

---

> [!IMPORTANT]
> **User Transparency:** Vectora will always inform when compression is active and the estimated loss of precision. Critical workspaces (e.g., production code) can opt-out of compression without functional penalty — only at the cost of higher disk/RAM usage.
