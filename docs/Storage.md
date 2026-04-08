# Blueprint: Persistência e Storage (BBolt + Chromem)

**Status:** Fase 4 - Implementação Concluída  
**Módulo:** `core/storage/`  
**Dependencies:** `go.etcd.io/bbolt`, `github.com/philippgille/chromem-go`, `golang.org/x/text` (para token counting básico ou integração com tiktoken)

## 1. Estrutura de Tipos e Contratos (`types.go`)

Definimos as structs que serão serializadas para JSON no BBolt e usadas como interface com o Chromem.

```go
package storage

import (
	"time"
)

// WorkspaceMeta armazena o estado de alto nível do projeto indexado.
// Serializado no bucket 'config' sob a chave 'workspace_state'.
type WorkspaceMeta struct {
	Path          string    `json:"path"`
	LastIndexedAt time.Time `json:"last_indexed_at"`
	ModelProvider string    `json:"model_provider"` // ex: "gemini-embedding-2"
	TotalFiles    int       `json:"total_files"`
	TotalChunks   int       `json:"total_chunks"`
	Status        string    `json:"status"` // "idle", "indexing", "error"
}

// FileIndexEntry mapeia um arquivo físico para seus vetores no Chromem.
// Chave no bucket 'file_index': path relativo do arquivo.
type FileIndexEntry struct {
	AbsolutePath string   `json:"absolute_path"`
	ContentHash  string   `json:"content_hash"` // SHA-256
	ChunkIDs     []string `json:"chunk_ids"`    // IDs dos documentos no Chromem
	UpdatedAt    time.Time `json:"updated_at"`
	SizeBytes    int64    `json:"size_bytes"`
}

// ChatMessage representa uma interação atômica.
type ChatMessage struct {
	Role      string    `json:"role"`      // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // Ex: tool_calls
}

// ChunkConfig define as regras de quebra de texto.
type ChunkConfig struct {
	MaxTokens int `json:"max_tokens"`
	Overlap   int `json:"overlap"`
}

var DefaultChunkConfig = ChunkConfig{
	MaxTokens: 800, // Conservador para embeddings precisos
	Overlap:   100, // "Costura quente" para contexto
}
```

## 2. Motor BBolt: Metadados Atômicos (`bbolt_store.go`)

Implementação robusta com transações curtas e criação automática de buckets.

```go
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.etcd.io/bbolt"
)

const (
	BucketConfig      = "config"
	BucketFileIndex   = "file_index"
	BucketChatHistory = "chat_history"
	KeyWorkspaceState = "workspace_state"
)

type BBoltStore struct {
	db *bbolt.DB
	mu sync.RWMutex // Proteção extra para leituras concorrentes em memória se necessário
}

func NewBBoltStore(dbPath string) (*BBoltStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout:      1 * time.Second, // Fail fast se estiver locked
		NoGrowSync:   false,
		FreelistType: bbolt.FreelistArrayType, // Melhor performance para writes sequenciais
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open bbolt: %w", err)
	}

	// Inicialização Idempotente dos Buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{BucketConfig, BucketFileIndex, BucketChatHistory}
		for _, b := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(b)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &BBoltStore{db: db}, nil
}

// SaveWorkspaceMeta persiste o estado global do workspace.
func (s *BBoltStore) SaveWorkspaceMeta(meta WorkspaceMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketConfig))
		return b.Put([]byte(KeyWorkspaceState), data)
	})
}

// GetWorkspaceMeta recupera o estado. Retorna erro se não existir.
func (s *BBoltStore) GetWorkspaceMeta() (*WorkspaceMeta, error) {
	var meta WorkspaceMeta
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketConfig))
		data := b.Get([]byte(KeyWorkspaceState))
		if data == nil {
			return fmt.Errorf("workspace state not found")
		}
		return json.Unmarshal(data, &meta)
	})
	return &meta, err
}

// UpsertFileIndex atualiza ou cria o índice de um arquivo.
func (s *BBoltStore) UpsertFileIndex(relPath string, entry FileIndexEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketFileIndex))
		return b.Put([]byte(relPath), data)
	})
}

// GetFileIndex recupera o índice de um arquivo específico.
func (s *BBoltStore) GetFileIndex(relPath string) (*FileIndexEntry, error) {
	var entry FileIndexEntry
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketFileIndex))
		data := b.Get([]byte(relPath))
		if data == nil {
			return fmt.Errorf("index not found for %s", relPath)
		}
		return json.Unmarshal(data, &entry)
	})
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// DeleteFileIndex remove o índice de um arquivo (usado na re-indexação).
func (s *BBoltStore) DeleteFileIndex(relPath string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketFileIndex))
		return b.Delete([]byte(relPath))
	})
}

// Close fecha o banco de dados de forma segura.
func (s *BBoltStore) Close() error {
	return s.db.Close()
}
```

## 3. Motor Vetorial: Chromem-go (`chromem_store.go`)

Gerencia coleções isoladas por workspace e operações de vetor.

```go
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/philippgille/chromem-go"
)

type ChromemStore struct {
	db *chromem.DB
}

func NewChromemStore(dataDir string) (*ChromemStore, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}

	// PersistentDB salva automaticamente em disco a cada alteração
	db, err := chromem.NewPersistentDB(dataDir, false) // false = sem compressão snappy (mais rápido CPU)
	if err != nil {
		return nil, fmt.Errorf("failed to init chromem: %w", err)
	}

	return &ChromemStore{db: db}, nil
}

// GetOrCreateCollection garante que a coleção do workspace exista.
// Usamos o ID do workspace como nome da coleção para isolamento total.
func (s *ChromemStore) GetOrCreateCollection(workspaceID string) (*chromem.Collection, error) {
	col := s.db.GetCollection(workspaceID)
	if col != nil {
		return col, nil
	}

	// Cria nova coleção com distância Cosseno (padrão ouro para embeddings de texto)
	col, err := s.db.CreateCollection(workspaceID, map[string]string{
		"type":    "codebase",
		"version": "1.0",
	}, chromem.NewCosineDistance())

	if err != nil {
		return nil, err
	}
	return col, nil
}

// AddVectors insere novos chunks no índice vetorial.
func (s *ChromemStore) AddVectors(ctx context.Context, workspaceID string, docs []chromem.Document) error {
	col, err := s.GetOrCreateCollection(workspaceID)
	if err != nil {
		return err
	}

	// AddDocuments é atômico para o batch
	return col.AddDocuments(ctx, docs, "")
}

// DeleteVectors remove vetores específicos por ID.
func (s *ChromemStore) DeleteVectors(ctx context.Context, workspaceID string, ids []string) error {
	col, err := s.GetOrCreateCollection(workspaceID)
	if err != nil {
		return err
	}
	return col.DeleteDocuments(ctx, ids)
}

// QueryVectors busca os vizinhos mais próximos de um embedding.
func (s *ChromemStore) QueryVectors(ctx context.Context, workspaceID string, embedding []float32, limit int) ([]chromem.SearchResult, error) {
	col, err := s.GetOrCreateCollection(workspaceID)
	if err != nil {
		return nil, err
	}

	// QueryEmbedding espera o vetor já calculado
	results, err := col.QueryEmbedding(ctx, embedding, limit, nil, nil)
	if err != nil {
		return nil, err
	}
	return results, nil
}
```

## 4. Lógica de Chunking Inteligente (`chunker.go`)

Implementa a estratégia de "Costura Quente" (Overlap) para preservar contexto semântico.

```go
package storage

import (
	"strings"
	"unicode/utf8"
)

// SimpleChunker divide o texto em blocos sobrepostos baseados em caracteres/tokens aproximados.
// Nota: Para MVP, usamos contagem de runes como proxy de tokens.
// Em produção, integraríamos com tiktoken para precisão absoluta.
func SimpleChunker(content string, config ChunkConfig) []string {
	if len(content) == 0 {
		return []string{}
	}

	var chunks []string
	runes := []rune(content)
	totalRunes := len(runes)

	// Ajuste simples: 1 token ~= 4 caracteres em média para código inglês/latino
	maxChars := config.MaxTokens * 4
	overlapChars := config.Overlap * 4

	if maxChars >= totalRunes {
		return []string{content}
	}

	start := 0
	for start < totalRunes {
		end := start + maxChars
		if end > totalRunes {
			end = totalRunes
		}

		// Tenta quebrar em linha nova para não cortar código no meio
		chunkEnd := end
		if end < totalRunes {
			// Procura a próxima newline após o limite para quebrar limpo
			nextNewline := strings.IndexRune(string(runes[end:]), '\n')
			if nextNewline != -1 && nextNewline < (maxChars/2) { // Não procure muito longe
				chunkEnd = end + nextNewline + 1
			}
		}

		chunk := string(runes[start:chunkEnd])
		chunks = append(chunks, chunk)

		// Move o start para trás do overlap para a "costura"
		start = chunkEnd - overlapChars
		if start < 0 {
			start = 0
		}

		// Segurança contra loop infinito se o overlap for maior que o chunk
		if start >= chunkEnd {
			break
		}
	}

	return chunks
}
```

## 5. Engine Unificado e Gerenciamento de Ciclo de Vida (`engine.go`)

Orquestra o BBolt e o Chromem, garantindo que ambos sejam fechados corretamente.

```go
package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type Engine struct {
	Meta       *BBoltStore
	Vec        *ChromemStore
	RootPath   string
	WorkspaceID string
}

// NewEngine inicializa o storage para um workspace específico.
// workspaceHash deve ser único para o path do projeto (ex: SHA256 do path absoluto).
func NewEngine(baseRoot string, workspaceHash string) (*Engine, error) {
	workspaceDir := filepath.Join(baseRoot, workspaceHash)

	metaStore, err := NewBBoltStore(filepath.Join(workspaceDir, "metadata.db"))
	if err != nil {
		return nil, fmt.Errorf("init bbolt: %w", err)
	}

	vecStore, err := NewChromemStore(filepath.Join(workspaceDir, "vector_db"))
	if err != nil {
		metaStore.Close()
		return nil, fmt.Errorf("init chromem: %w", err)
	}

	return &Engine{
		Meta:        metaStore,
		Vec:         vecStore,
		RootPath:    workspaceDir,
		WorkspaceID: workspaceHash,
	}, nil
}

// CalculateHash gera SHA-256 de um conteúdo.
func CalculateHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// IndexFile é a operação atômica de alta nível:
// 1. Calcula hash novo.
// 2. Compara com BBolt.
// 3. Se diferente, deleta vetores antigos no Chromem.
// 4. Chunka novo conteúdo.
// 5. Insere vetores novos no Chromem.
// 6. Atualiza BBolt.
func (e *Engine) IndexFile(ctx context.Context, relPath string, content string) error {
	newHash := CalculateHash(content)

	// 1. Verifica se precisa re-indexar
	existing, err := e.Meta.GetFileIndex(relPath)
	if err == nil && existing.ContentHash == newHash {
		return nil // Nada mudou, skip rápido
	}

	// 2. Limpeza de vetores antigos se existirem
	if existing != nil && len(existing.ChunkIDs) > 0 {
		if err := e.Vec.DeleteVectors(ctx, e.WorkspaceID, existing.ChunkIDs); err != nil {
			return fmt.Errorf("delete old vectors: %w", err)
		}
	}

	// 3. Chunking
	chunks := SimpleChunker(content, DefaultChunkConfig)
	if len(chunks) == 0 {
		return nil
	}

	// 4. Preparar documentos para Chromem
	docs := make([]chromem.Document, len(chunks))
	newChunkIDs := make([]string, len(chunks))

	for i, chunk := range chunks {
		// ID único: hash_do_arquivo:indice
		chunkID := fmt.Sprintf("%s:%d", newHash, i)
		newChunkIDs[i] = chunkID

		docs[i] = chromem.Document{
			ID:        chunkID,
			Content:   chunk,
			Metadata: map[string]string{
				"source": relPath,
				"hash":   newHash,
				"index":  fmt.Sprintf("%d", i),
			},
		}
	}

	// 5. Inserir no Chromem
	if err := e.Vec.AddVectors(ctx, e.WorkspaceID, docs); err != nil {
		return fmt.Errorf("add new vectors: %w", err)
	}

	// 6. Atualizar Metadata no BBolt
	entry := FileIndexEntry{
		AbsolutePath: filepath.Join(e.RootPath, relPath), // Ou path original do projeto
		ContentHash:  newHash,
		ChunkIDs:     newChunkIDs,
		SizeBytes:    int64(len(content)),
		UpdatedAt:    time.Now(),
	}

	if err := e.Meta.UpsertFileIndex(relPath, entry); err != nil {
		// Rollback parcial: deveríamos deletar os vetores inseridos se falhar aqui
		e.Vec.DeleteVectors(ctx, e.WorkspaceID, newChunkIDs)
		return fmt.Errorf("update meta: %w", err)
	}

	return nil
}

func (e *Engine) Close() error {
	return e.Meta.Close()
}
```

---

### Notas de Implementação

1.  **Dependência de Tempo:** Adicione `"time"` aos imports do `bbolt_store.go` e `types.go`.
2.  **Contexto:** Todas as operações de I/O pesado ou rede (se houver no futuro) usam `context.Context` para permitir cancelamento.
3.  **Segurança:** O `IndexFile` faz um rollback manual simples se falhar ao atualizar o metadata. Em um sistema distribuído, usaríamos transações de duas fases, mas para local-first, essa abordagem é suficiente e rápida.
