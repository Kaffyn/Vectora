#!/usr/bin/env python3
"""Check ToolNode behavior and signatures."""

import inspect
import sys
from pathlib import Path

# Add vectora to path
sys.path.insert(0, str(Path(__file__).parent / "vectora"))

from langchain.tools import tool
from langgraph.prebuilt.tool_node import ToolNode


# Create a simple tool
@tool
def test_tool(x: str) -> str:
    """Test tool."""
    return f"Result: {x}"


# Create ToolNode
tools = [test_tool]
tn = ToolNode(tools=tools)

# Check what it returns
print("ToolNode class:", type(tn))

# Check if it has a run method
print("Has run method?:", hasattr(tn, "run"))
print("Has ainvoke method?:", hasattr(tn, "ainvoke"))
print("Has invoke method?:", hasattr(tn, "invoke"))
print("Has __call__?:", callable(tn))

# Let's see what ToolNode actually is
print("\nToolNode base classes:", tn.__class__.__bases__)
print("\nToolNode methods and attributes:")
for name in dir(tn):
    if not name.startswith("_"):
        attr = getattr(tn, name)
        if callable(attr):
            print(f"  {name}() [callable]")
        else:
            print(f"  {name} [attribute]")

# Check invoke signature
if hasattr(tn, "invoke"):
    sig = inspect.signature(tn.invoke)
    print("\ninvoke signature:", sig)

# Check ainvoke signature
if hasattr(tn, "ainvoke"):
    sig = inspect.signature(tn.ainvoke)
    print("ainvoke signature:", sig)
