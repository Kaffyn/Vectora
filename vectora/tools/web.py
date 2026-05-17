"""Web tools: busca e extração de conteúdo da internet via Tavily."""

import json
import logging

from langchain.tools import tool
from settings import settings
from tavily import TavilyClient

logger = logging.getLogger(__name__)


@tool
def web_search(query: str) -> str:
    """Busca a web por informações atuais usando Tavily (otimizado para agentes).

    Tavily retorna resultados estruturados, prontos para RAG, com conteúdo já extraído.

    Args:
        query: String da consulta de busca

    Returns:
        JSON com resultados estruturados (url, title, content) prontos para embedding
    """
    if not settings.enable_web_search:
        logger.warning("web_search tool called but disabled")
        return "Web search is disabled. Enable ENABLE_WEB_SEARCH=true to use this tool."

    if not settings.tavily_api_key:
        logger.error("TAVILY_API_KEY not configured")
        return json.dumps(
            {
                "status": "error",
                "error": "TAVILY_API_KEY not configured. Set TAVILY_API_KEY environment variable.",
            }
        )

    logger.info("web_search tool called", extra={"query": query})

    try:
        client = TavilyClient(api_key=settings.tavily_api_key)
        response = client.search(query=query, search_depth="advanced", max_results=5)

        logger.info(
            "web_search completed",
            extra={"query": query, "num_results": len(response.get("results", []))},
        )

        return json.dumps(response["results"])
    except Exception:
        logger.exception("web_search failed", extra={"query": query})
        return json.dumps(
            {
                "status": "error",
                "error": "Web search failed. Please try again.",
            }
        )


@tool
def fetch_url(url: str) -> str:
    """Busca e extrai conteúdo de texto de uma URL específica usando Tavily.

    Args:
        url: URL para buscar (deve começar com http:// ou https://)

    Returns:
        Conteúdo de texto extraído da página
    """
    if not url.startswith(("http://", "https://")):
        logger.warning("fetch_url called with invalid URL", extra={"url": url})
        return "Error: URL must start with http:// or https://"

    if not settings.tavily_api_key:
        logger.error("TAVILY_API_KEY not configured")
        return "Error: TAVILY_API_KEY not configured. Cannot fetch URL."

    logger.info("fetch_url tool called", extra={"url": url})

    try:
        client = TavilyClient(api_key=settings.tavily_api_key)
        response = client.extract(urls=[url])

        results = response.get("results", [])
        if not results:
            logger.warning("fetch_url returned no content", extra={"url": url})
            return f"No content found at {url}"

        content = results[0].get("raw_content", "") or results[0].get("content", "")

        logger.info(
            "fetch_url completed",
            extra={"url": url, "content_length": len(content)},
        )

        return content

    except Exception:
        logger.exception("fetch_url failed", extra={"url": url})
        return "Error occurred fetching URL. Please check logs."
