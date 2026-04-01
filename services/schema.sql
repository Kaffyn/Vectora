-- ============================================================================
-- Vectora Index Service - PostgreSQL Schema
-- ============================================================================
-- Executar em Supabase PostgreSQL com pgvector extension

-- Enable pgvector extension (ja está habilitado em Supabase por padrão)
CREATE EXTENSION IF NOT EXISTS vector;

-- ============================================================================
-- WORKSPACES - Multi-tenancy
-- ============================================================================
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT workspaces_name_owner_unique UNIQUE(owner_id, name)
);

CREATE INDEX idx_workspaces_owner_id ON workspaces(owner_id);

-- ============================================================================
-- INDEXES - Coleções de documentos (um por workspace)
-- ============================================================================
CREATE TABLE IF NOT EXISTS indexes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    document_count INT DEFAULT 0,
    size_bytes BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT indexes_workspace_name_unique UNIQUE(workspace_id, name)
);

CREATE INDEX idx_indexes_workspace_id ON indexes(workspace_id);

-- ============================================================================
-- DOCUMENTS - Metadados de documentos
-- ============================================================================
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    index_id UUID NOT NULL REFERENCES indexes(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    filename VARCHAR(255),
    content_type VARCHAR(50),
    file_size BIGINT,
    uploaded_by VARCHAR(255),
    storage_path VARCHAR(512),  -- Caminho em Supabase Storage (opcional)
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_documents_index_id ON documents(index_id);
CREATE INDEX idx_documents_workspace_id ON documents(workspace_id);
CREATE INDEX idx_documents_created_at ON documents(created_at DESC);

-- ============================================================================
-- CHUNKS - Segmentos com embeddings vetoriais
-- ============================================================================
CREATE TABLE IF NOT EXISTS chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    index_id UUID NOT NULL REFERENCES indexes(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    chunk_number INT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),  -- OpenAI/Gemini: 1536 dimensões
    metadata JSONB,          -- {filename, page, offset, source, etc}
    created_at TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT chunks_workspace_fk FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- Índices para performance
CREATE INDEX idx_chunks_document_id ON chunks(document_id);
CREATE INDEX idx_chunks_index_id ON chunks(index_id);
CREATE INDEX idx_chunks_workspace_id ON chunks(workspace_id);

-- Índice IVFFLAT para busca vetorial (pgvector)
-- IVFFLAT é mais escalável que HNSW para grandes datasets
CREATE INDEX IF NOT EXISTS idx_chunks_embedding_ivfflat ON chunks
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Índice HNSW alternativo (comentado - testar performance)
-- CREATE INDEX IF NOT EXISTS idx_chunks_embedding_hnsw ON chunks
-- USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);

-- ============================================================================
-- INDEXING_STATUS - Rastreamento de progresso de indexação
-- ============================================================================
CREATE TABLE IF NOT EXISTS indexing_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    index_id UUID NOT NULL REFERENCES indexes(id),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    status VARCHAR(50) DEFAULT 'pending',  -- pending, processing, completed, error
    progress_percent INT DEFAULT 0,
    chunks_processed INT DEFAULT 0,
    total_chunks INT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT indexing_status_status_check CHECK (status IN ('pending', 'processing', 'completed', 'error'))
);

CREATE INDEX idx_indexing_status_document_id ON indexing_status(document_id);
CREATE INDEX idx_indexing_status_workspace_id ON indexing_status(workspace_id);
CREATE INDEX idx_indexing_status_status ON indexing_status(status);

-- ============================================================================
-- AUDIT_LOG - Auditoria de operações (opcional, para produção)
-- ============================================================================
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,  -- upload, search, delete, etc
    resource_type VARCHAR(50),    -- document, index, workspace
    resource_id UUID,
    user_id VARCHAR(255),
    details JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_audit_log_workspace_id ON audit_log(workspace_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);

-- ============================================================================
-- VIEW: Informações agregadas de índices
-- ============================================================================
CREATE OR REPLACE VIEW index_stats AS
SELECT
    i.id,
    i.workspace_id,
    i.name,
    COUNT(DISTINCT d.id) as document_count,
    COUNT(DISTINCT c.id) as chunk_count,
    SUM(d.file_size) as total_size_bytes,
    MAX(d.created_at) as last_document_added,
    i.created_at,
    i.updated_at
FROM indexes i
LEFT JOIN documents d ON i.id = d.index_id
LEFT JOIN chunks c ON i.id = c.index_id
GROUP BY i.id, i.workspace_id, i.name, i.created_at, i.updated_at;

-- ============================================================================
-- TRIGGERS: Auto-update updated_at
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_workspaces_updated_at BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_indexes_updated_at BEFORE UPDATE ON indexes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_indexing_status_updated_at BEFORE UPDATE ON indexing_status
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- RLS (Row-Level Security) - Optional para produção
-- ============================================================================
-- ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE indexes ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE chunks ENABLE ROW LEVEL SECURITY;
--
-- CREATE POLICY workspace_isolation ON workspaces
--     FOR ALL USING (owner_id = current_user_id());

-- ============================================================================
-- PERFORMANCE TIPS
-- ============================================================================
-- 1. Aumentar work_mem para melhor performance de índices:
--    SET work_mem = '256MB';
--
-- 2. Vacuumar periodicamente (Supabase faz automaticamente)
--    VACUUM ANALYZE;
--
-- 3. Monitorar índice IVFFLAT:
--    SELECT * FROM pg_stat_user_indexes WHERE relname LIKE 'idx_chunks_%';
--
-- 4. Se performance degradar, reindexar:
--    REINDEX INDEX idx_chunks_embedding_ivfflat;
