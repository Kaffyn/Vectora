# Checkpointers e Snapshots

Um checkpointer disponibiliza uma camada de persistência de dados para o grafo,
por isso precisamos de algum serviço para salvar esses dados (base de dados,
arquivo ou outro tipo de serviço).

A cada **super-step**, o checkpointer salva um Snapshot do estado naquele ponto
do tempo.

Isso nos permite fazer _Time Travel_, _Human-in-the-loop_, ter histórico de
conversas, tolerância a falhas, fazer fork de conversas e executar várias outras
ações que desejarmos relacionadas aos chats passados e futuros.

Um checkpointer é identificado por dois atributos importantes, o `thread_id` e o
`checkpoint_id`. Em uma conversa identificada por um único `thread_id`, são
criados vários checkpoints, cada um com o seu próprio ID.

Exemplo:

![Exemplo de super-steps no LangGraph](./assets/012/checkpointer001.jpg)

A cada execução do grafo, novos checkpoints são salvos na base de dados e é
possível usar apenas a config inicial do grafo com o mesmo `thread_id` para
obter todo o histórico de conversas. Porém, também é possível usar a combinação
de `thread_id` com `checkpoint_id` para usar algum checkpoint específico no
histórico.

As funções do grafo `get_state` e `get_state_history` podem ser usadas para
obter snapshots do histórico para uma Thread. `get_state` retorna o último
snapshot com todas as mensagens, já `get_state_history` retorna uma lista
contendo todos os snapshots disponíveis.

Já existem alguns checkpointers prontos para o LangGraph, como `InMemorySaver`,
`SqliteSaver` e `PostgresSaver`. Para versões assíncronas temos
`AsyncSqliteSaver` e `AsyncPostgresSaver` (`InMemorySaver` também pode ser usado
em código assíncrono).

---

## Instalação dos checkpointers

O LangGraph disponibiliza os pacotes para os checkpointers mais usados, como
`memory` (não recomendado para produção), `sqlite` e `postgresql`.

Para SQLite e PostgreSQL, precisamos instalar os pacotes do que queremos usar:

```bash
# PostgreSQL
# Se você ver algum erro sobre esses pacotes, basta instalar:
# psycopg[binary], psycopg ou psycopg[c]
uv add "psycopg[binary]" # no meu caso, só precisei de psycopg[binary]
uv add langgraph-checkpoint-postgres

# SQLite
uv add langgraph-checkpoint-sqlite
uv add aiosqlite # se for usar AsyncSqliteSaver
```

Você pode usar `sqlite` em produção, mas a recomendação é usar PostgreSQL para
soluções mais escaláveis e assíncronas.

---

## Docker

Se você for subir um container com o Docker para o PostgreSQL, já vou deixar
isso pronto para você:

```yml
services:
  psql:
    container_name: psql
    hostname: psql
    image: postgres:18
    restart: "no"
    environment:
      POSTGRES_USER: langgraph # ⚠️ IMPORTANTE: Usuário
      POSTGRES_PASSWORD: defina_sua_senha # ⚠️ IMPORTANTE: senha
      POSTGRES_DB: langgraph # ⚠️ IMPORTANTE: base de dados
      PGDATA: /var/lib/postgresql/18/docker
    volumes:
      # ⚠️ IMPORTANTE: Estou em src/examples/ex010/.
      # ⚠️ IMPORTANTE: Todas as pastas _local_data precisam estar no .gitignore.
      # ℹ️ INFO: Isso vai gerar a pasta src/examples/ex010/_local_data no projeto
      # 🆘 docker-compose -f src/examples/ex010/compose.yml up
      - ./_local_data/postgresql:/var/lib/postgresql
    ports:
      - 5432:5432
```

Veja que deixei vários recadinhos para você no `compose.yml`. Não adicione sua
senha ali, prefira usar o `.env` ou outro local seguro.

---

## Na aula 11

Na aula 11 (em vídeo), vou te mostrar como configurar checkpointers de uma forma
desacoplada, para que você possa conseguir usar sua conexão de base de dados
para outras partes da sua aplicação. Além disso, vamos usar o conceito de
Lifespan da aula anterior para garantirmos que nossas conexões serão fechadas
mesmo que ocorra um erro no meio do grafo.

O código está em:

- [../src/examples/ex012](../src/examples/ex012)

---
