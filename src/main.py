"""CLI Entry Point for Vectora Application.

Handles command-line argument parsing and application startup.
Routes to chat TUI, setup wizard, or other CLI commands.
"""

import asyncio
import logging

from langchain_core.messages import AIMessage, BaseMessage, HumanMessage
from langgraph.graph.state import RunnableConfig
from langgraph.pregel.main import BaseCheckpointSaver
from rich import print
from rich.markdown import Markdown
from rich.prompt import Prompt

from checkpointer import Checkpointer
from constants import DB_DSN
from context import Context
from graph import build_graph
from log_setup import setup_logging
from utils import async_lifespan

logger = logging.getLogger(__name__)


async def run_graph(checkpointer: BaseCheckpointSaver) -> None:
    graph = build_graph(checkpointer)

    thread_id = 1
    context = Context(user_type="plus", thread_id=thread_id)

    config = RunnableConfig(
        configurable={"thread_id": thread_id},
    )

    logger.info(
        "Graph session started",
        extra={"thread_id": thread_id, "user_type": context.user_type},
    )

    all_messages: list[BaseMessage] = []

    prompt = Prompt()
    Prompt.prompt_suffix = ""

    while True:
        user_input = prompt.ask("[bold cyan]Você: \n")
        print(Markdown("\n\n  ---  \n\n"))

        if user_input.lower() in ["q", "quit"]:
            logger.info(
                "Session ended by user",
                extra={"thread_id": thread_id},
            )
            break

        logger.debug(
            "User input received",
            extra={"thread_id": thread_id, "input_length": len(user_input)},
        )

        human_message = HumanMessage(user_input)
        current_loop_messages = [human_message]

        # Streaming reativo com astream_events
        print("[bold cyan]RESPOSTA:[/bold cyan] \n")
        ai_response = ""
        async for event in graph.astream_events(
            {"messages": current_loop_messages},
            config=config,
            version="v2",
        ):
            event_type = event.get("event")
            data = event.get("data", {})

            # Streaming de tokens do LLM
            if event_type == "on_chat_model_stream":
                chunk = data.get("chunk")
                if chunk and hasattr(chunk, "content"):
                    content = chunk.content
                    if content:
                        ai_response += content
                        print(content, end="", flush=True)

        print(Markdown("\n\n  ---  \n\n"))

        # Ler estado final para salvar histórico
        final_state = await graph.aget_state(config=config)
        all_messages = final_state.values.get("messages", [])

    print(await graph.aget_state(config=config))


async def main() -> None:
    setup_logging()
    logger.info("Vectora Chat started")

    async with (
        async_lifespan(),
        Checkpointer(DB_DSN) as checkpointer,
    ):
        logger.info("Checkpointer initialized (SQLite)")
        await run_graph(checkpointer)

    logger.info("Vectora Chat ended")


if __name__ == "__main__":
    asyncio.run(main())
