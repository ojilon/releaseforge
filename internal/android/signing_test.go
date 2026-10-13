package android

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSigningDefaults(t *testing.T) {
	root := t.TempDir()
	cfg, dir, err := ResolveSigning(root, "myapp", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if dir != SigningDir(root, "myapp") {
		t.Fatalf("dir %q", dir)
	}
	want := DefaultKeystorePath(root, "myapp")
	if cfg.KeystorePath != want {
		t.Fatalf("keystore %q want %q", cfg.KeystorePath, want)
	}
}

func TestResolveSigningFileAndExplicit(t *testing.T) {
	root := t.TempDir()
	dir := SigningDir(root, "myapp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	ks := filepath.Join(dir, "my.jks")
	os.WriteFile(ks, []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "signing.json"),
		[]byte(`{"keystore_path":"my.jks","alias":"filealias"}`), 0o644)
	cfg, _, err := ResolveSigning(root, "myapp", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Alias != "filealias" || cfg.KeystorePath != ks {
		t.Fatalf("got %+v", cfg)
	}
	// Explicit alias wins over file.
	cfg, _, _ = ResolveSigning(root, "myapp", "", "explicit", "")
	if cfg.Alias != "explicit" {
		t.Fatalf("explicit alias lost: %+v", cfg)
	}
}

func TestKeystorePasswordEnv(t *testing.T) {
	t.Setenv(EnvKeystorePass, "s3cret")
	p, err := KeystorePassword()
	if err != nil || p != "s3cret" {
		t.Fatalf("got %q %v", p, err)
	}
}
