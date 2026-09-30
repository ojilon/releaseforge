from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parent.parent
GRADLE_PROPERTIES = ROOT / "gradle.properties"

def read_version():
    text = GRADLE_PROPERTIES.read_text(encoding="utf-8")

    code = re.search(r"^app\.versionCode=(.+)$", text, re.MULTILINE)
    name = re.search(r"^app\.versionName=(.+)", text, re.MULTILINE)

    if not code or not name:
        raise RuntimeError("Version information not found in gradle.properties")

    return code.group(1).strip(), name.group(1).strip()


def set_version(version_name):
    text = GRADLE_PROPERTIES.read_text(encoding="utf-8")

    code_match = re.search(r"^app\.versionCode=(.+)$", text, re.MULTILINE)
    name_match = re.search(r"^app\.versionName=(.+)$", text, re.MULTILINE)

    if not code_match or not name_match:
        raise RuntimeError("Version information not found in gradle.properties")


    current_code = int(code_match.group(1).strip())
    new_code = current_code + 1

    text = re.sub(r"^app\.versionCode=.+$", f"app.versionCode={new_code}", text, flags=re.MULTILINE)

    text = re.sub(r"^app\.versionName=.+$", f"app.versionName={version_name}", text, flags=re.MULTILINE)

    GRADLE_PROPERTIES.write_text(text, encoding="utf-8")

    print("Version updated:")
    print(f" versionCode: {new_code}")
    print(f" versionName: {version_name}")


if __name__ == "__main__":
    if len(sys.argv) == 1:
        code, name = read_version()
        print(f"versionCode: {code}")
        print(f"versionName: {name}")

    elif len(sys.argv) == 2:
        set_version(sys.argv[1])

    else:
        print("Usage:")
        print(" python scripts/version.py")
        print(" python scripts/version.py 0.1.0")
        sys.exit(1)
