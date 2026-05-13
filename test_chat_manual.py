#!/usr/bin/env python3
"""Interactive chat test for Vectora.

Para usar:
    uv run python test_chat_manual.py

Digite suas mensagens e veja as respostas. Use 'sair' ou 'quit' para sair.
"""

import asyncio
import sys
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parent))


async def main() -> None:
    """Run interactive test."""
    from dotenv import load_dotenv
    from src.graph import build_graph
    from src.state import State
    from langchain_core.messages import HumanMessage

    load_dotenv()

    print("\n╔════════════════════════════════════════════════════════╗")
    print("║          VECTORA - Manual Chat Test                   ║")
    print("║  Digite suas mensagens para conversar com Vectora     ║")
    print("║  Use 'sair' ou 'quit' para encerrar                  ║")
    print("╚════════════════════════════════════════════════════════╝\n")

    # Initialize graph
    try:
        from src.checkpointer import build_checkpointer_sqlite
        checkpointer = build_checkpointer_sqlite()
        graph = build_graph(checkpointer=checkpointer)
    except Exception as e:  # noqa: BLE001
        print(f"Erro ao inicializar: {e}")
        sys.exit(1)

    messages: list = []
    turn = 0

    while True:
        try:
            user_input = input("\n✏️  Voce: ").strip()

            if not user_input:
                continue

            if user_input.lower() in ["sair", "quit"]:
                print("\n👋 Ate logo!\n")
                break

            messages.append(HumanMessage(content=user_input))

            state = State(messages=messages)
            result = await graph.ainvoke(state)

            turn += 1

            if result.get("messages"):
                latest = result["messages"][-1]
                if hasattr(latest, "content"):
                    print(f"\n🤖 Vectora: {latest.content}")

        except KeyboardInterrupt:
            print("\n\n👋 Interrompido.\n")
            break
        except Exception as e:  # noqa: BLE001
            print(f"\n❌ Erro: {e}")


if __name__ == "__main__":
    asyncio.run(main())
