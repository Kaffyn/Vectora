import os

if __name__ == "__main__":
    os.environ.setdefault("LOG_LEVEL", "INFO")
    os.environ.setdefault("LOG_JSON", "false")

    from config import Config

    config = Config.instance()
    if not config.get_llm_provider():
        from setup_wizard_tui import run_setup

        run_setup()
    else:
        from chat import run_chat

        run_chat()
