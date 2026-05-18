from langchain_cohere import CohereEmbeddings, CohereRerank
from langchain_core.documents import Document

from vectora.config.settings import settings


class RetrievalService:
    def __init__(self):
        api_key = settings.get_cohere_api_key()
        self.embeddings = CohereEmbeddings(
            cohere_api_key=api_key, model="embed-multilingual-v3.0"
        )
        self.reranker = CohereRerank(
            cohere_api_key=api_key, model="rerank-multilingual-v3.0"
        )

    async def get_relevant_context(
        self, query: str, collection: str, limit: int = 5
    ) -> list[Document]:
        """Pipeline RAG atômico: Busca, Rerank e Filtragem por score."""
        # 1. Busca Vetorial (LanceDB)
        raw_docs = await self._vector_search(query, collection, limit * 2)

        # 2. Reranking Cohere
        reranked = self.reranker.compress_documents(raw_docs, query)

        # 3. Filtro de Score (Deterministico)
        # Removemos qualquer coisa abaixo de 0.3 de relevância antes de enviar pro LLM
        return [doc for doc in reranked if getattr(doc, "relevance_score", 1.0) > 0.3]

    async def _vector_search(self, query: str, collection: str, limit: int) -> None:
        # Lógica original do vector_search aqui, mas retornando apenas Documents
        ...
