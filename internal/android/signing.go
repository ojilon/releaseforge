package android

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SigningConfig is the small JSON file that tells ReleaseForge where your
// signing files live. It lives OUTSIDE any git repo, under the data root:
//
//	<data-root>/signing/<project>/signing.json
//	<data-root>/signing/<project>/keystore.jks   <- put your keystore here
//
// Example signing.json:
//
//	{"keystore_path": "keystore.jks", "alias": "mykey", "apksigner_path": ""}
//
// keystore_path may be absolute or relative to the signing dir.
// Passwords are NEVER stored: set env RF_KEYSTORE_PASS before signing.
type SigningConfig struct {
	KeystorePath  string `json:"keystore_path"`
	Alias         string `json:"alias"`
	ApksignerPath string `json:"apksigner_path"`
}

// Env vars honoured for signing (password never stored on disk).
const (
	EnvKeystorePass  = "RF_KEYSTORE_PASS"
	EnvKeystore      = "RF_KEYSTORE"
	EnvKeystoreAlias = "RF_KEYSTORE_ALIAS"
	EnvApksigner     = "RF_APKSIGNER"
)

// SigningDir returns <root>/signing/<project>. Safe to create on demand.
func SigningDir(root, project string) string {
	name := strings.TrimSpace(project)
	if name == "" {
		name = "default"
	}
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	return filepath.Join(root, "signing", name)
}

// SigningConfigPath returns the signing.json path for a project.
func SigningConfigPath(root, project string) string {
	return filepath.Join(SigningDir(root, project), "signing.json")
}

// DefaultKeystorePath is where to drop your .jks when the config
// does not name one: <root>/signing/<project>/keystore.jks
func DefaultKeystorePath(root, project string) string {
	return filepath.Join(SigningDir(root, project), "keystore.jks")
}

// loadFile reads signing.json when present (nil, nil when absent).
func loadFile(path string) (*SigningConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	var c SigningConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("signing config %s: %w", path, err)
	}
	return &c, nil
}

// ResolveSigning merges signing settings by precedence:
// explicit args > project signing.json > global signing.json > env.
// It returns the resolved config plus the signing dir used.
func ResolveSigning(root, project, keystore, alias, apksigner string) (SigningConfig, string, error) {
	dir := SigningDir(root, project)
	cfg := SigningConfig{}

	// Global fallback first, then project file overrides it.
	if g, _ := loadFile(filepath.Join(root, "signing", "signing.json")); g != nil {
		cfg = *g
	}
	if p, err := loadFile(SigningConfigPath(root, project)); err != nil {
		return cfg, dir, err
	} else if p != nil {
		if strings.TrimSpace(p.KeystorePath) != "" {
			cfg.KeystorePath = p.KeystorePath
		}
		if strings.TrimSpace(p.Alias) != "" {
			cfg.Alias = p.Alias
		}
		if strings.TrimSpace(p.ApksignerPath) != "" {
			cfg.ApksignerPath = p.ApksignerPath
		}
	}

	if strings.TrimSpace(keystore) != "" {
		cfg.KeystorePath = strings.TrimSpace(keystore)
	} else if v := strings.TrimSpace(os.Getenv(EnvKeystore)); v != "" && strings.TrimSpace(cfg.KeystorePath) == "" {
		cfg.KeystorePath = v
	}
	if strings.TrimSpace(alias) != "" {
		cfg.Alias = strings.TrimSpace(alias)
	} else if v := strings.TrimSpace(os.Getenv(EnvKeystoreAlias)); v != "" && strings.TrimSpace(cfg.Alias) == "" {
		cfg.Alias = v
	}
	if strings.TrimSpace(apksigner) != "" {
		cfg.ApksignerPath = strings.TrimSpace(apksigner)
	} else if v := strings.TrimSpace(os.Getenv(EnvApksigner)); v != "" && strings.TrimSpace(cfg.ApksignerPath) == "" {
		cfg.ApksignerPath = v
	}

	// Default keystore location when nothing names one.
	if strings.TrimSpace(cfg.KeystorePath) == "" {
		cfg.KeystorePath = DefaultKeystorePath(root, project)
	} else if !filepath.IsAbs(cfg.KeystorePath) {
		// Relative paths resolve against the project signing dir,
		// then against the global signing dir, then CWD.
		for _, base := range []string{dir, filepath.Join(root, "signing")} {
			cand := filepath.Join(base, cfg.KeystorePath)
			if st, err := os.Stat(cand); err == nil && !st.IsDir() {
				cfg.KeystorePath = cand
				break
			}
		}
		if !filepath.IsAbs(cfg.KeystorePath) {
			cfg.KeystorePath = filepath.Join(dir, cfg.KeystorePath)
		}
	}
	return cfg, dir, nil
}

// KeystorePassword returns the signing password from the environment.
// It is never read from disk.
func KeystorePassword() (string, error) {
	if p := os.Getenv(EnvKeystorePass); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("set %s with your keystore password (never stored in files)", EnvKeystorePass)
}

// SignApkCapture signs + verifies like SignApk but returns combined output
// as a string so the TUI can print it in the log pane.
func SignApkCapture(apkPath, keystorePath, alias, password, apksignerBin string) (string, error) {
	if apkPath == "" || keystorePath == "" || alias == "" {
		return "", fmt.Errorf("sign: apk, keystore and alias are required")
	}
	if _, err := os.Stat(apkPath); err != nil {
		return "", fmt.Errorf("sign: apk not found: %s", apkPath)
	}
	if _, err := os.Stat(keystorePath); err != nil {
		return "", fmt.Errorf("sign: keystore not found: %s — put it at %s", keystorePath, keystorePath)
	}
	bin := FindApksigner(apksignerBin)
	var b strings.Builder
	sign := exec.Command(bin, "sign",
		"--ks", keystorePath,
		"--ks-key-alias", alias,
		"--ks-pass", "pass:"+password,
		apkPath)
	if out, err := sign.CombinedOutput(); err != nil {
		b.Write(out)
		return b.String(), fmt.Errorf("apksigner sign failed: %w\n%s", err, strings.TrimSpace(string(out)))
	} else {
		b.Write(out)
	}
	verify := exec.Command(bin, "verify", "--verbose", apkPath)
	if out, err := verify.CombinedOutput(); err != nil {
		b.Write(out)
		return b.String(), fmt.Errorf("apksigner verify failed: %w\n%s", err, strings.TrimSpace(string(out)))
	} else {
		b.Write(out)
	}
	return b.String(), nil
}

// InstallCapture installs apk via adb and returns combined output
// (TUI-friendly; CLI Install streams instead).
func InstallCapture(apkPath, serial string) (string, error) {
	if _, err := os.Stat(apkPath); err != nil {
		return "", fmt.Errorf("install: apk not found: %s (build first)", apkPath)
	}
	args := withSerial(serial, []string{"install", "-r", apkPath})
	out, err := exec.Command("adb", args...).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("adb install failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
