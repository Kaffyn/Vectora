# Iniciar Vectora

## 🎯 Opção 1: CLI Interativa (Recomendado para Teste)

```bash
# Setup
cp .env.example .env
nano .env  # Adicionar GOOGLE_API_KEY

# Iniciar
uv run python -m src.main
```

**O que acontece:**

- Terminal interativo simples
- Você digita mensagens no prompt
- Bot responde com Markdown
- Digite 'q' ou 'quit' para sair
- Usa Rich para formatação bonita

**Exemplo:**

```
Você:
Ola, como voce funciona?

RESPOSTA (gemini-2.0-flash):
Ola! Eu sou Vectora, um agente de IA...

---

Você:
_
```

---

## 🎨 Opção 2: TUI Textual (Interface Gráfica no Terminal)

```bash
# Setup
cp .env.example .env
nano .env

# Iniciar TUI
uv run python -m src.chat
```

**O que acontece:**

- Interface TUI (Terminal User Interface) no estilo Textual
- Header com título "⚡ Vectora Chat"
- Log de mensagens formatadas
- Input box para digitar
- Mais sofisticado visualmente

**Keybinds:**

- TAB: Focar entre widgets
- CTRL+C: Sair

---

## 🐳 Opção 3: Docker Compose (Completo com Banco de Dados)

```bash
# Build e iniciar
docker compose up -d

# Testar API
curl http://localhost:8000/health
```

---

## ⚙️ Requisitos

- `.env` configurado com API key (GOOGLE_API_KEY, OPENAI_API_KEY, etc)
- Python 3.13+
- uv instalado
- Para Docker: Docker Engine + Compose

---

## 🐛 Troubleshooting

**"GOOGLE_API_KEY not set"**

```bash
# Copiar template e editar
cp .env.example .env
export GOOGLE_API_KEY="seu-api-key"
```

**"Module not found"**

```bash
# Certifique-se de estar no diretório raiz
cd /path/to/vectora
uv sync  # Reinstalar dependências
```

**"Connection refused" (Docker)**

```bash
# Aguardar inicialização (30-60s)
docker compose logs -f vectora

# Ou verificar
docker compose ps
```
