"""Debug script to check tools configuration."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent / "vectora"))

from tools import TOOLS, TOOLS_BY_NAME

print(f"Total tools: {len(TOOLS)}")
print("\nTools available:")
for tool in TOOLS:
    print(
        f"  - {tool.name}: {tool.description[:60] if tool.description else 'No description'}"
    )

print(f"\nTools by name: {list(TOOLS_BY_NAME.keys())}")

# Check if list_dir is available
if "list_dir" in TOOLS_BY_NAME:
    print("\n[OK] list_dir tool is available")
    list_dir_tool = TOOLS_BY_NAME["list_dir"]
    print(f"  Type: {type(list_dir_tool)}")
    print(f"  Callable: {callable(list_dir_tool)}")
else:
    print("\n[ERROR] list_dir tool not found!")
