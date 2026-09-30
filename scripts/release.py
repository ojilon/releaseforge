from pathlib import Path
import subprocess
import sys
import os

from version import set_version
from notes import create_notes
from package import package_release
from zip_release import create_zip
from sign import sign_apk
from config import load_config


ROOT = Path(__file__).resolve().parent.parent

def get_gradle_cmd(gradle_kind="gradlew"):
    if gradle_kind == "-l":
        return "gradle"
    
    """Returns absolute path to gradlew executable."""
    if os.name == "nt":
        gradlew = ROOT / "gradlew.bat"
    else:
        gradlew = ROOT / "gradlew"

    return str(gradlew) if gradlew.exists() else "gradle"

def run_gradle_task(task_name, gradle_user_home=None, gradle_kind="gradle"):
    cmd = [get_gradle_cmd(gradle_kind), task_name]
    if gradle_user_home:
        cmd.extend(["-g", gradle_user_home])

    print("\n" + "="*60)
    print("Running:", " ".join(cmd))

    #use shell=true on Windows if executing batch file directly
    result = subprocess.run(cmd, cwd=ROOT, shell=(os.name == "nt"))
    if result.returncode != 0:
        raise RuntimeError(f"Gradle tesk '{task_name}' failed with exit code {result.returncode}")

def publish_github_release(version, release_type):
    tag = f"v{version}"
    release_dir = ROOT / "release" / version
    notes_file = ROOT / "release_notes" / f"v{version}.md"

    # Gather all build artifacts in the version release folder (.apk and .zip)
    artifacts = list(release_dir.glob("*.apk")) + list(release_dir.glob("*.zip"))
    artifact_paths = [str(f.resolve()) for f in artifacts]

    cmd = [
        "gh", "release", "create", tag,
        "--title", f"Conductino Study {version}",
        "--notes-file", str(notes_file.resolve())
    ]

    # Mark as pre-release if specified
    if release_type == "pre-release":
        cmd.append("--prerelease")

    # Append artifact paths
    cmd.extend(artifact_paths)

    print()
    print("=" * 60)
    print("Publishing GitHub Release...")
    print("=" * 60)

    # Use shell=True on Windows to execute 'gh' reliably
    result = subprocess.run(cmd, cwd=ROOT, shell=(os.name == 'nt'))
    if result.returncode != 0:
        raise RuntimeError("Failed to publish release to GitHub.")

    print(f"Successfully published {tag} to GitHub!")


def main():
    if len(sys.argv) != 4:
        print("Usage:")
        print(" python scripts/release.py VERSION_TYPE -l(system gradle / leave blank to use gradlew)")
        print()
        print("Examples:")
        print(" python scripts/release.py 0.0.3 pre-release -l(write '-l' to force using your local gradle, not the repo's gradlew)")
        print(" python scripts/release.py 1.0.0 release  -l(or leave blank)")
        sys.exit(1)

    version = sys.argv[1]
    release_type = sys.argv[2]
    gradle_kind = sys.argv[3] #fetch the '-l'

    print()
    print("Conductino Study Release Tool")
    print("------------------------------")
    print(f"Version: {version}")
    print(f"Type: {release_type}")

    #Update version
    print()
    print("[1/9] Updating version...")
    set_version(version)

    cfg = load_config()

    #Run unit tests
    print()
    print("[2/9] Running unit tests...")
    run_gradle_task("testDebugUnitTest", cfg.get("gradle_user_home"), gradle_kind)

    #Build Debug APK
    print()
    print("[3/9] Building Debug APK...")
    run_gradle_task("assembleDebug", cfg.get("gradle_user_home"), gradle_kind)

    #Build release APK
    print()
    print("[4/9] Building Release APK...")
    run_gradle_task("assembleRelease", cfg.get("gradle_user_home"), gradle_kind)

    #Package APKs
    print()
    print("[5/9 Packaging APKs....]")
    package_release(version)

    # Sign release APK locally
    print()
    print("[6/9] Signing Release APK...")

    release_dir = ROOT / "release" / version
    unsigned_apk = release_dir / f"Conductino-Study-{version}-release.apk"

    keystore_arg = cfg.get("keystore_path", "conductino-release.jks")
    keystore = Path(keystore_arg)
    if not keystore.is_absolute():
        keystore = ROOT / keystore

    sign_apk(
        apk_path=unsigned_apk,
        keystore_path=keystore,
        alias=cfg.get("keystore_alias", "conductino"),
        custom_apksigner=cfg.get("apksigner_path")
    )

    #Create release notes
    print()
    print("[7/9] Creating release notes")
    create_notes(version, release_type)

    #Create ZIP
    print()
    print("[8/9] Creating release ZIP")
    create_zip(version)

    print()
    print("=" * 60)
    print("RELEASE PACKAGE COMPLETE")
    print("=" * 60)
    print()
    print(f"Output: {ROOT / 'release' / version}")
    print(f"Notes:  {ROOT / 'release_notes' / f'v{version}.md'}")

    #publish to Github
    print()
    print("[9/9] Publishing Release to GitHub...")
    publish_github_release(version, release_type)


if __name__ == '__main__':
    try:
        main()
    except Exception as e:
        print()
        print("RELEASE FAILED")
        print(e)
        sys.exit(1)