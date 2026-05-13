import os

if __name__ == "__main__":
    os.environ.setdefault("LOG_LEVEL", "INFO")
    os.environ.setdefault("LOG_JSON", "false")

    from chat import run_chat

    run_chat()
