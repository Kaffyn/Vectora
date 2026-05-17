"""Filesystem tools: leitura, escrita, edição de arquivos, grep, listagem e terminal."""

import logging
import platform
import re
import subprocess as sp
from pathlib import Path

from langchain.tools import tool

from vectora.config.settings import settings
from vectora.services.security import (
    is_safe_file_path,
    is_safe_regex_pattern,
    is_safe_shell_command,
)

logger = logging.getLogger(__name__)


@tool
def file_read(file_path: str) -> str:
    """Lê conteúdo completo de um arquivo de texto.

    Args:
        file_path: Caminho relativo ou absoluto do arquivo

    Returns:
        Conteúdo do arquivo como string
    """
    if not settings.enable_file_operations:
        return "File operations are disabled. Enable ENABLE_FILE_OPERATIONS=true to use this tool."

    if not is_safe_file_path(file_path, allowed_dirs=["."]):
        logger.warning("file_read blocked by safety check", extra={"path": file_path})
        return f"Error: File path '{file_path}' is not allowed"

    try:
        content = Path(file_path).read_text(encoding="utf-8")
        logger.info(
            "file_read completed", extra={"path": file_path, "size": len(content)}
        )
        return content
    except FileNotFoundError:
        return f"Error: File '{file_path}' not found"
    except Exception:
        logger.exception("file_read failed", extra={"path": file_path})
        return "Error reading file. Check logs."


@tool
def file_edit(
    file_path: str, old_text: str, new_text: str, replace_all: bool = False
) -> str:
    """Edita arquivo substituindo texto.

    Args:
        file_path: Caminho do arquivo
        old_text: Texto a encontrar (use "" para criar arquivo se não existir)
        new_text: Texto de substituição
        replace_all: Se True, substitui todas as ocorrências; padrão substitui apenas a 1ª

    Returns:
        Confirmação da edição
    """
    if not settings.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_file_path(file_path, allowed_dirs=["."]):
        logger.warning("file_edit blocked by safety check", extra={"path": file_path})
        return f"Error: File path '{file_path}' is not allowed"

    try:
        path = Path(file_path)

        # Cria arquivo novo quando old_text="" e arquivo não existe
        if old_text == "" and not path.exists():
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(new_text, encoding="utf-8")
            logger.info("file_edit created new file", extra={"path": file_path})
            return f"[OK] File created: {file_path}"

        content = path.read_text(encoding="utf-8")

        if old_text and old_text not in content:
            return "Error: Text not found in file"

        new_content = (
            content.replace(old_text, new_text)
            if replace_all
            else content.replace(old_text, new_text, 1)
        )
        path.write_text(new_content, encoding="utf-8")

        count = content.count(old_text) if replace_all else 1
        logger.info(
            "file_edit completed",
            extra={"path": file_path, "occurrences": count, "replace_all": replace_all},
        )
        return f"[OK] File edited successfully ({count} occurrence{'s' if count != 1 else ''} replaced)"
    except Exception:
        logger.exception("file_edit failed", extra={"path": file_path})
        return "Error editing file. Check logs."


@tool
def file_write(file_path: str, content: str) -> str:
    """Cria ou sobrescreve completamente um arquivo com o conteúdo fornecido.

    Use para criar novos arquivos ou substituir o conteúdo completo de um existente.
    Para edições cirúrgicas (substituir trechos), prefira file_edit.

    Args:
        file_path: Caminho do arquivo (absoluto ou relativo)
        content: Conteúdo completo a escrever no arquivo

    Returns:
        Confirmação com caminho e tamanho em bytes
    """
    if not settings.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_file_path(file_path, allowed_dirs=["."]):
        logger.warning("file_write blocked by safety check", extra={"path": file_path})
        return f"Error: File path '{file_path}' is not allowed"

    try:
        path = Path(file_path)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")

        size = path.stat().st_size
        logger.info(
            "file_write completed", extra={"path": file_path, "size_bytes": size}
        )
        return f"[OK] File written: {file_path} ({size} bytes)"
    except Exception:
        logger.exception("file_write failed", extra={"path": file_path})
        return "Error writing file. Check logs."


@tool
def grep(pattern: str, path: str = ".") -> str:
    """Busca padrão em arquivos usando regex.

    Args:
        pattern: Padrão regex para buscar
        path: Caminho da pasta ou arquivo

    Returns:
        Linhas que correspondem ao padrão (arquivo:linha: conteúdo)
    """
    if not settings.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_regex_pattern(pattern):
        return "Error: Invalid or unsafe regex pattern"

    try:
        results = []
        search_path = Path(path)
        files = [search_path] if search_path.is_file() else list(search_path.rglob("*"))

        for file_path in files:
            if not file_path.is_file() or file_path.suffix in {".pyc", ".o", ".exe"}:
                continue
            try:
                content = file_path.read_text(encoding="utf-8", errors="ignore")
                for line_num, line in enumerate(content.split("\n"), 1):
                    if re.search(pattern, line):
                        results.append(f"{file_path}:{line_num}: {line}")
            except Exception:
                pass

        logger.info(
            "grep completed",
            extra={"pattern": pattern, "path": path, "matches": len(results)},
        )
        return "\n".join(results[:100]) if results else "No matches found"
    except Exception:
        logger.exception("grep failed", extra={"pattern": pattern, "path": path})
        return "Error during grep. Check logs."


@tool
def list_dir(path: str = ".", *, recursive: bool = False) -> str:
    """Lista arquivos em um diretório.

    Args:
        path: Caminho do diretório
        recursive: Se True, lista recursivamente

    Returns:
        Lista de arquivos e pastas com prefixo [DIR] ou [FILE]
    """
    if not settings.enable_file_operations:
        return "File operations are disabled."

    try:
        dir_path = Path(path)

        if not dir_path.exists():
            return f"Error: Directory '{path}' not found"
        if not dir_path.is_dir():
            return f"Error: '{path}' is not a directory"

        items = []
        if recursive:
            for item in sorted(dir_path.rglob("*")):
                rel_path = item.relative_to(dir_path)
                prefix = "[DIR]" if item.is_dir() else "[FILE]"
                items.append(f"{prefix} {rel_path}")
        else:
            for item in sorted(dir_path.iterdir()):
                prefix = "[DIR]" if item.is_dir() else "[FILE]"
                items.append(f"{prefix} {item.name}")

        logger.info(
            "list_dir completed",
            extra={"path": path, "recursive": recursive, "count": len(items)},
        )
        return "\n".join(items[:500]) if items else "(empty directory)"
    except Exception:
        logger.exception("list_dir failed", extra={"path": path})
        return "Error listing directory. Check logs."


@tool
def terminal(command: str) -> str:
    """Executa um comando shell. Permite tudo exceto comandos destrutivos (rm -rf, mkfs, etc).

    Args:
        command: Comando shell para executar

    Returns:
        Saída do comando (stdout + stderr) ou mensagem de erro se bloqueado
    """
    # Normaliza comandos Unix → Windows quando necessário
    if platform.system() == "Windows":
        command = re.sub(r"\bmkdir\s+-p\s+", "mkdir ", command)
        command = re.sub(r"\bmkdir\s+-p\s*$", "mkdir .", command)

    if not settings.enable_file_operations:
        return "File operations are disabled."

    if not is_safe_shell_command(command):
        logger.warning(
            "terminal command blocked by safety check",
            extra={"command": command[:50]},
        )
        return (
            f"Error: Command '{command}' is blocked for safety. "
            "Destructive commands like rm -rf, mkfs, dd if=/dev/zero, "
            "and fork bombs are not permitted."
        )

    try:
        result = sp.run(
            command,
            shell=True,
            capture_output=True,
            text=True,
            timeout=30,
            check=False,
        )

        output = result.stdout
        if result.stderr:
            output += result.stderr

        logger.info(
            "terminal_command_executed",
            extra={
                "command": command[:50],
                "exit_code": result.returncode,
                "output_length": len(output),
            },
        )

        return output or f"Command executed with exit code {result.returncode}"

    except sp.TimeoutExpired:
        logger.warning("terminal_command_timeout", extra={"command": command[:50]})
        return "Error: Command timed out after 30 seconds"
    except Exception:
        logger.exception("terminal_command_failed", extra={"command": command[:50]})
        return "Error executing command. Check logs."
