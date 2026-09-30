from pathlib import Path
import subprocess
import getpass
import os
import shutil


def find_apksigner(custom_path=None):
    """Finds apksigner executable via config, PATH, or ANDROID_HOME."""
    if custom_path and Path(custom_path).exists():
        return str(Path(custom_path).resolve())

    #Check system path first
    system_tool = shutil.which("apksigner")
    if system_tool:
        return system_tool

    #Search in ANDROID_HOME / ANDROID_SDK_ROOT
    sdk = os.environ.get("ANDROID_HOME") or os.environ.get("ANDROID_SDK_ROOT")
    if not sdk:
        build_tools = Path(sdk) / "build_tools"
        if build_tools.exists():
            #find the highest version build tools directory
            versions = sorted([d for d in build_tools.iterdir() if d.is_dir()], reverse=True)
            for ver in versions:
                exe = "apksigner.bat" if os.name == "nt" else "apksigner"
                candidate = ver / exe
                if candidate.exists():
                    return str(candidate.resolve())

    # Fallback to standard command invocation
    return "apksigner.bat" if os.name == "nt" else "apksigner"


def sign_apk(apk_path, keystore_path, alias, custom_apksigner=None):
    apk_path = Path(apk_path).resolve()
    keystore_path = Path(keystore_path).resolve()

    if not apk_path.exists():
        raise FileNotFoundError(f"APK not found: {apk_path}")

    if not keystore_path.exists():
        raise FileNotFoundError(f"Keystore not found: {keystore_path}")

    apksigner_bin = find_apksigner(custom_apksigner)
    password = getpass.getpass("Keystore password: ")

    print(f"\nSigning APK: {apk_path.name} ...")

    
    sign_command = [
        apksigner_bin,
        "sign",
        "--ks", str(keystore_path),
        "--ks-key-alias", alias,
        "--ks-pass", f"pass:{password}",
        str(apk_path),
    ]

    result = subprocess.run(sign_command, shell=(os.name == "nt"))

    if result.returncode != 0:
        raise RuntimeError("APK signing failed.")

    print("APK signed successfully.")

    # Fix: Execute verify
    print()
    print("Verifying APK signature...")
    verify_commmand = [
        apksigner_bin,
        "verify",
        "--verbose",
        str(apk_path),
    ]

    verify = subprocess.run(verify_commmand, shell=(os.name == "nt"))
    if verify.returncode != 0:
        raise RuntimeError("APK signature verification failed.")

    print("APK signature verified successfully.")

    return apk_path