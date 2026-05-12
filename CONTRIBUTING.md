# Contributing to Vectora

First off, thank you for considering contributing to Vectora! It's people like you that make Vectora such a great tool for the open-source community.

We welcome all kinds of contributions: bug reports, feature requests, documentation improvements, and code contributions. This document provides the guidelines and steps to set up your development environment properly.

## Development Environment Setup

To ensure consistency and high quality across the codebase, we enforce a strict set of tools for development, linting, and formatting. You must install the following prerequisites before contributing.

### Prerequisites

You need to have the following tools installed globally on your machine:

1. **Python Management:**

   - **[pyenv](https://github.com/pyenv/pyenv):** Used to easily install and manage multiple Python versions. We currently target Python 3.13+.

2. **Package & Project Management:**

   - **[uv](https://github.com/astral-sh/uv):** An extremely fast Python package and project manager. We use `uv` exclusively instead of `pip` or `poetry`.

3. **Code Formatting & Linting (Python):**

   - **[Ruff](https://docs.astral.sh/ruff/):** An extremely fast Python linter and code formatter.
   - **[Pyright](https://microsoft.github.io/pyright/):** A fast type checker meant for large Python source bases. We strictly enforce type hints across the project.
   - **[ty](https://github.com/astral-sh/ty):** A type checker runner. We recommend using `uvx ty` to seamlessly run type checks (like Pyright) without needing to configure or manage their environments manually.

4. **Code Formatting (Markdown, YAML, TS/JS):**

   - **[Prettier](https://prettier.io/):** Used to format Markdown, YAML, and frontend files. You must install it globally via Node Package Manager or Bun (`npm install -g prettier` or `bun add -g prettier`).

5. **Pre-commit Hooks:**
   - **[pre-commit](https://pre-commit.com/):** A framework for managing and maintaining multi-language pre-commit hooks. It ensures that no code is committed without passing the required linting and formatting checks.

## Getting Started

Once you have installed the prerequisites, follow these steps to set up your local development environment.

### 1. Clone the Repository

Fork the repository on GitHub and clone your fork locally.

```bash
git clone https://github.com/your-org/vectora.git
cd vectora
```

### 2. Install Dependencies

Use `uv` to create a virtual environment and install the project dependencies. This step will also resolve all required packages based on `pyproject.toml`.

```bash
uv sync
```

### 3. Set Up Pre-commit Hooks

You must install the pre-commit hooks before making any changes. This will automatically run Ruff, Mypy, Prettier, and other checks defined in `.pre-commit-config.yaml` before every commit.

```bash
pre-commit install
```

## Commit Guidelines

We strictly follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This leads to more readable messages that are easy to follow when looking through the project history.

Your commit messages should follow this format:

- `feat:` for new features.
- `fix:` for bug fixes.
- `docs:` for documentation changes.
- `chore:` for maintenance, dependency updates, etc.
- `refactor:` for code refactoring without changing behavior.
- `test:` for adding or modifying tests.

## Submitting Changes

When your changes are ready, open a Pull Request (PR) against the `main` branch. Ensure that all local tests and pre-commit hooks pass successfully before requesting a review.
