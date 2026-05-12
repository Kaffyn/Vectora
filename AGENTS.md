# Diretrizes para Agentes de IA (AGENTS.md)

Este documento estabelece as regras, padrões e fluxos de trabalho que devem ser rigorosamente seguidos por qualquer Agente de IA (como eu) ao atuar no desenvolvimento do projeto **Vectora**.

## 1. Regras de Formatação de Markdown

É estritamente obrigatório inserir parágrafos de texto entre títulos e subtítulos. Nunca deixe dois cabeçalhos (ex: um `##` seguido imediatamente de um `###`) sem um texto introdutório ou explicativo entre eles. Toda seção deve ter contexto.

## 2. Fluxo de Trabalho e Versionamento

O versionamento é essencial. Após a conclusão de cada tarefa lógica, arquivo modificado ou funcionalidade implementada, o Agente deve **sempre** executar os comandos de versionamento no terminal:

```bash
git add .
git commit -m "<tipo>: <mensagem clara e descritiva>"
```

### 2.1. Padrão de Commits (Conventional Commits)

Os commits devem obrigatoriamente seguir a especificação do _Conventional Commits_:

- `feat:` Para novas funcionalidades.
- `fix:` Para correção de bugs.
- `docs:` Para alterações exclusivas em documentação (README, lang.md, etc).
- `refactor:` Para refatoração de código sem alteração de comportamento.
- `chore:` Para atualizações de dependências (`uv.lock`, `.pre-commit-config.yaml`) e tarefas de build.
- `test:` Para criação ou alteração de testes.

## 3. Padrões de Arquitetura e Código (Design Patterns)

O Vectora é um software avançado e exige um alto padrão de engenharia. Os seguintes padrões devem ser aplicados no código:

### 3.1. Tipagem e Validação Fortes

Todo código Python deve fazer uso massivo de _Type Hints_ (tipagem estática) utilizando os padrões do Python 3.14.5+. O uso do `pydantic` é mandatório para validação de contratos de dados, _schemas_ de API e modelos internos.

### 3.2. Modularidade e Injeção de Dependências

O código deve ser altamente modular, aplicando princípios do SOLID. Para recursos como Bancos de Vetores, deve-se usar o padrão de **Camada de Abstração (Repository Pattern)**. A aplicação não deve depender diretamente da implementação do LanceDB ou Qdrant em seus _controllers_, mas sim de uma interface comum (ex: via LangChain VectorStore) definida dinamicamente.

### 3.3. Programação Assíncrona

Como o Vectora faz uso do **FastAPI**, **LangGraph** e integrações externas (APIs de LLM), toda a base de código I/O-bound (banco de dados, rede) deve ser assíncrona (`async/await`). Funções bloqueantes não devem ser usadas na _Main Thread_.

### 3.4. Estado e Grafos (LangGraph)

Qualquer fluxo de tomada de decisão do agente ou _pipeline_ de longo prazo deve ser modelado no **LangGraph** utilizando grafos de estado. Deve-se manter os _Nodes_ (nós) do grafo puros e independentes, lendo e escrevendo estritamente no _State_ (estado) passado pelo orquestrador.

### 3.5. Tratamento de Erros

Exceções devem ser tratadas localmente sempre que fizer sentido, retornando mensagens claras. No contexto da API, utilizar _Exception Handlers_ do FastAPI. Para Agentes, garantir que ferramentas (Tools) tenham blocos de _try/except_ para que falhas não derrubem o grafo inteiro, mas retornem o erro como observação ao LLM.

## 4. Gerenciamento de Dependências

O gerenciador oficial do projeto é o `uv`. Nenhuma dependência deve ser instalada via `pip install` direto sem refletir no `pyproject.toml`. O projeto utiliza o conceito de grupos/extras (ex: `[qdrant]`) para a instalação "Dual-Mode".

## 5. Qualidade e Pre-commits

Qualquer código submetido deverá passar nos _hooks_ do `pre-commit` já configurados (Ruff, Isort, Mypy, Prettier). O agente deve se certificar de formatar o código previamente ou entender que o commit automático ajustará pequenos desvios.
