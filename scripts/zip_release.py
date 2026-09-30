from pathlib import Path
import zipfile
import sys

ROOT = Path(__file__).resolve().parent.parent
RELEASE_DIR = ROOT / "release"


def create_zip(version):
    
    version_dir = RELEASE_DIR / version
    if not version_dir.exists():
        raise FileNotFoundError(f"Release directory not found: {version_dir}\n" f"Run package.py first.")

    apk_files = list(version_dir.glob("*.apk"))
    if not apk_files:
        raise FileNotFoundError(f"No APK files found in: {version_dir}")

    zip_path = version_dir / f"Conductino-Study-{version}.zip"
    if zip_path.exists():
        print(f"ZIP already exists: {zip_path}")
        return

    with zipfile.ZipFile(zip_path, "w", compression=zipfile.ZIP_DEFLATED) as archive:
        for apk in apk_files:
            archive.write(apk, arcname=apk.name)

    print()
    print("Release ZIP created:")
    print(f"  {zip_path}")
    print()
    print("Contents:")

    for apk in apk_files:
        print(f"  {apk.name}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage:")
        print("  python scripts/zip_release.py 0.0.2")
        sys.exit(1)

    create_zip(sys.argv[1])
