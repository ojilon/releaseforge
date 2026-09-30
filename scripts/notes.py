from pathlib import Path
from datetime import date
import sys


ROOT = Path(__file__).resolve().parent.parent
NOTES_DIR = ROOT / "release_notes"

def create_notes(version, release_type="pre-release"):
    NOTES_DIR.mkdir(exist_ok=True)

    filename = NOTES_DIR / f"v{version}.md"

    if filename.exists():
        print(f"Release notes already exists: {filename}")
        return

    content = f"""# Conductino Study {version}

    **Type:** {release_type}
    **Date:** {date.today().isoformat()}

    ## Highlights
    -

    ## Changes

    -

    ## Bug fixe

    -

    ## Known Issues

    -

    ## Downlaods

    Release artifacts will be attached to the Github releases
    """

    filename.write_text(content, encoding="utf-8")
    print(f"Created release notes: {filename}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python scripts/notes.py 0.0.2")
        print("  python scripts/notes.py 0.0.2 release")
        sys.exit(1)

    version = sys.argv[1]
    release_type = sys.argv[2] if len(sys.argv) > 2 else "pre-release"

    create_notes(version, release_type)

