import os
import tarfile
import hashlib


class ModelPackager:
    """Manages the compression and versioning of Godot/Vectora vector indices."""

    def __init__(self, root_dir: str):
        self.root_dir = root_dir
        self.models_dir = os.path.join(root_dir, "models")
        # Artifacts for release are stored in root /assets folder
        self.dist_dir = os.path.join(root_dir, "assets", "models")
        os.makedirs(self.dist_dir, exist_ok=True)

    def calculate_sha256(self, filepath: str) -> str:
        """Calculates the SHA256 checksum of a file."""
        sha256 = hashlib.sha256()
        with open(filepath, "rb") as f:
            for chunk in iter(lambda: f.read(4096), b""):
                sha256.update(chunk)
        return sha256.hexdigest()

    def package_version(self, engine_version: str, version_tag: str):
        """Compresses specific version data for distribution."""
        path = os.path.join(self.models_dir, engine_version, version_tag)
        if not os.path.exists(path):
            print(f"⚠️ Error: Path {path} not found.")
            return

        archive_name = f"knowledge_{engine_version}_{version_tag}.tar.gz"
        archive_path = os.path.join(self.dist_dir, archive_name)

        print(f"📦 Packaging {engine_version} ({version_tag})...")
        with tarfile.open(archive_path, "w:gz") as tar:
            tar.add(path, arcname=os.path.basename(path))

        print(f"✅ Generated: {archive_name}")


if __name__ == "__main__":
    packager = ModelPackager(".")
    # Example: Package Godot 4.4 r1
    # packager.package_version("godot-4.4", "r1")
