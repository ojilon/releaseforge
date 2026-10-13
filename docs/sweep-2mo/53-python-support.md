# 53 — Python Support (pyproject, venv, tasks)

Detected (`detect.go` python branch) but task-poor. Target: read
`pyproject.toml` (project.version, scripts), `requirements.txt` freeze,
venv detection (`.venv`, `VIRTUAL_ENV`); seed tasks
`{build: "python -m build", test: "pytest -q"}`; version R/W on
`pyproject.toml` or config-declared file; `release` zips wheel/sdist.

Read: https://packaging.python.org/ ; pip/venv docs.
Tasks: runner + version R/W + fixture. Acceptance: `releaseforge build` in fixture runs build.
