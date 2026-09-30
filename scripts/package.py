from pathlib import Path
import shutil
import sys

ROOT = Path(__file__).resolve().parent.parent

APK_DIR = ROOT / "app" / "build" / "outputs" / "apk"
RELEASE_DIR = ROOT / "release"

def package_release(version):
    version_dir = RELEASE_DIR / version

    if version_dir.exists():
        print(f"Release directory already exists: {version_dir}")
        print("Remove it manually so as to package again")
        return

    version_dir.mkdir(parents=True)

    debug_apk = APK_DIR / "debug" / "app-debug.apk"
    release_apk = APK_DIR / "release" / "app-release-unsigned.apk"

    if not debug_apk.exists():
        raise FileNotFoundError(f"Debug APK not found: {debug_apk}\n" "Run: gradle assembleDebug")

    if not release_apk.exists():
        raise FileNotFoundError(f"Release APK not found: {release_apk}\n" "Run: gradle assembleRelease")

    debug_destination = (version_dir / f"Conductino-Study-{version}-debug.apk")
    release_destination = (version_dir / f"Conductino-Study-{version}-release.apk")

    shutil.copy2(debug_apk, debug_destination)
    shutil.copy2(release_apk, release_destination)

    print()
    print("Release package created:")
    print(f"  {version_dir}")
    print()
    print(f"  {debug_destination.name}")
    print(f"  {release_destination.name}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage:")
        print(" python scripts/package.py 0.0.2")
        sys.exit(1)

    package_release(sys.argv[1])
