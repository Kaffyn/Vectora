import os
import requests
import zipfile
import io


class ModelSyncer:
    """Facilitates synchronization of external documentation for training/RAG."""

    def __init__(self, root_dir: str = "."):
        self.download_cache = os.path.join(root_dir, "models", "cache", "docs")
        os.makedirs(self.download_cache, exist_ok=True)

    def download_godot_docs(self, version: str):
        """Downloads official Godot documentation XMLs (e.g., 4.4)."""
        url = f"https://github.com/godotengine/godot/archive/refs/heads/{version}.zip"
        print(f"🌐 Downloading Godot {version} official documentation...")

        response = requests.get(url)
        if response.status_code == 200:
            with zipfile.ZipFile(io.BytesIO(response.content)) as zip_ref:
                # Extracting only the doc/classes folder
                for file in zip_ref.namelist():
                    if "/doc/classes/" in file:
                        zip_ref.extract(file, self.download_cache)
            print(f"✅ Download completed at: {self.download_cache}")
        else:
            print(f"❌ Download failed: {response.status_code}")


if __name__ == "__main__":
    syncer = ModelSyncer()
    # syncer.download_godot_docs("4.4")
