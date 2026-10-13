package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/spf13/cobra"
)

var (
	signApk       string
	signKeystore  string
	signAlias     string
	signApksigner string
)

// signCmd signs an APK using a keystore kept OUTSIDE the repo:
//
//	<data-root>/signing/<project>/keystore.jks
//	<data-root>/signing/<project>/signing.json
//
// Password comes from env RF_KEYSTORE_PASS, else an interactive prompt.
// Logic mirrors scripts/sign.py (apksigner sign + verify).
var signCmd = &cobra.Command{
	Use:   "sign [apk]",
	Short: "Sign an APK with your data-root keystore (then verify)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		apk := strings.TrimSpace(signApk)
		if len(args) > 0 {
			apk = strings.TrimSpace(args[0])
		}
		if apk == "" {
			if r, err := android.ResolveApk(info.Root, android.ReleaseApkPath, android.DebugApkPath); err == nil {
				apk = r
			} else {
				return fmt.Errorf("sign: no APK (build first or pass a path)")
			}
		}
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		cfg, _, err := android.ResolveSigning(root, info.Name, signKeystore, signAlias, signApksigner)
		if err != nil {
			return err
		}
		if strings.TrimSpace(cfg.Alias) == "" {
			return fmt.Errorf("sign: no alias — set --alias or add it to %s", android.SigningConfigPath(root, info.Name))
		}
		pass := strings.TrimSpace(os.Getenv(android.EnvKeystorePass))
		if pass == "" {
			fmt.Print("Keystore password: ")
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			pass = strings.TrimSpace(line)
		}
		if pass == "" {
			return fmt.Errorf("sign: empty password (or set %s)", android.EnvKeystorePass)
		}
		fmt.Printf("signing %s ...\n", apk)
		if err := android.SignApk(apk, cfg.KeystorePath, cfg.Alias, pass, cfg.ApksignerPath); err != nil {
			return err
		}
		fmt.Println("signed + verified.")
		return nil
	},
}

func init() {
	signCmd.Flags().StringVar(&signApk, "apk", "", "APK path (default: Gradle output)")
	signCmd.Flags().StringVar(&signKeystore, "keystore", "", "keystore path (default: data-root signing dir)")
	signCmd.Flags().StringVar(&signAlias, "alias", "", "key alias (default: signing.json)")
	signCmd.Flags().StringVar(&signApksigner, "apksigner", "", "custom apksigner binary")
}
