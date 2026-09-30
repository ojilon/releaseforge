import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
CONFIG_FILE = ROOT / "scripts" / "config.json"

DEFAULT_CONFIG = {
    "keystore_path": "conductino-release.jks",
    "keystore_alias": "conductino",
    "gradle_user_home": None,
    "apksigner_path": None,
}


def load_config():
    if not CONFIG_FILE.exists():
        return DEFAULT_CONFIG

    with open(CONFIG_FILE, "r", encoding="utf-8") as f:
        data = json.load(f)

    #fill in defaults incase of missing keys
    for key, val in DEFAULT_CONFIG.items():
        data.setdefault(key, val)
    return data