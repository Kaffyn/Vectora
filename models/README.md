# 🧠 Model Storage: Conhecimento Local

Este diretório gerencia o armazenamento de modelos de linguagem e índices vetoriais. Em vez de uma pasta única, o **Vectora** utiliza uma hierarquia organizada para suportar múltiplas versões de Engines e Revisões de Modelos.

## 📂 Organização por Namespaces

### 1. `gguf/` (LLMs Locais)
Armazena os pesos dos modelos Qwen 2.5 Coder otimizados para inferência local.
- **Estrutura**: `gguf/{model-name}/{quantization}.gguf`
- **Níveis Oficiais**:
    - **0.5B**: ~600MB de RAM. Foco em RAG puro.
    - **3B**: ~2.5GB de RAM. Padrão para coding assist.
    - **7B**: ~5.5GB de RAM. Máxima profundidade técnica.

### 2. `indices/` (Bases Vetoriais chromem-go)
Contém as bases de dados pré-calculadas para documentação de engines e frameworks.
- **Estrutura**: `indices/{engine-name}/{version}/{revision}/`
- **Conteúdo**: Arquivos de persistência e metadados do chromem-go.
- **Exemplo**: `indices/godot-4.2/v1/` contém o mapeamento completo da API XML da Godot 4.2.

## 🔄 Gestão de Dados
Os dados neste diretório são configurados pelo **Instalador** ou pelo comando `vectora sync`. O sistema utiliza hashes para garantir a integridade dos pesos do modelo e a paridade entre a API XML indexada e os chunks vetoriais.

## 📜 Regras de Uso
- **Não editar manualmente**: Os índices chromem-go são geridos internamente pelo core.
- **Persistence Affinity**: Os modelos GGUF e persistência do chromem-go são otimizados para garantir que o sistema operacional gerencie o cache de disco de forma eficiente, protegendo a RAM principal.
