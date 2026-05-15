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

        # if len(all_messages) == 0:
        #     current_loop_messages = [SystemMessage(SYSTEM_PROMPT), human_message]

        result = await graph.ainvoke(
            {"messages": current_loop_messages}, config=config, context=context
        )

        model_name = ""
        last_message = result["messages"][-1]

        if isinstance(last_message, AIMessage):
            model_name = last_message.response_metadata.get("model", "")

        content = (
            last_message.content
            if hasattr(last_message, "content")
            else str(last_message)
        )

        print(f"[bold cyan]RESPOSTA ({model_name}): \n")
        print(Markdown(content))
        print(last_message)
        print(Markdown("\n\n  ---  \n\n"))

        all_messages = result["messages"]

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
