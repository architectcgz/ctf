package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ctf-platform/internal/platform/randomstring"
)

func resolveContainerFlagGlobalSecret(secret, secretFile string, allowAutoGenerate bool) (string, error) {
	secret = strings.TrimSpace(secret)
	secretFile = strings.TrimSpace(secretFile)

	if secret != "" {
		if secretFile == "" {
			return secret, nil
		}
		persistedSecret, exists, err := loadPersistedContainerFlagGlobalSecret(secretFile)
		if err != nil {
			return "", fmt.Errorf("load container.flag_global_secret_file: %w", err)
		}
		if exists {
			if persistedSecret != secret {
				return "", fmt.Errorf("container.flag_global_secret does not match persisted secret file %s", secretFile)
			}
			return secret, nil
		}
		if err := persistContainerFlagGlobalSecret(secretFile, secret); err != nil {
			return "", fmt.Errorf("persist container.flag_global_secret_file: %w", err)
		}
		return secret, nil
	}

	if secretFile == "" {
		return "", fmt.Errorf("container.flag_global_secret must be set via CTF_CONTAINER_FLAG_GLOBAL_SECRET or persisted via container.flag_global_secret_file")
	}

	persistedSecret, exists, err := loadPersistedContainerFlagGlobalSecret(secretFile)
	if err != nil {
		return "", fmt.Errorf("load container.flag_global_secret_file: %w", err)
	}
	if exists {
		return persistedSecret, nil
	}

	if !allowAutoGenerate {
		return "", fmt.Errorf("container.flag_global_secret must be explicitly configured via CTF_CONTAINER_FLAG_GLOBAL_SECRET or pre-created container.flag_global_secret_file")
	}

	generatedSecret, err := randomstring.Generate()
	if err != nil {
		return "", fmt.Errorf("generate container.flag_global_secret: %w", err)
	}
	generatedSecret = strings.TrimSpace(generatedSecret)
	if err := persistContainerFlagGlobalSecret(secretFile, generatedSecret); err != nil {
		return "", fmt.Errorf("persist generated container.flag_global_secret_file: %w", err)
	}
	return generatedSecret, nil
}

func (c *Config) resolveContainerFlagSecretKeyring() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	activeSecret := strings.TrimSpace(c.Container.FlagGlobalSecret)
	if activeSecret == "" {
		return fmt.Errorf("container.flag_global_secret must not be empty")
	}
	if len(activeSecret) < 32 {
		return fmt.Errorf("container.flag_global_secret must be at least 32 bytes, current length: %d", len(activeSecret))
	}

	activeKeyID := strings.TrimSpace(c.Container.FlagGlobalSecretKeyID)
	if activeKeyID == "" {
		activeKeyID = "default"
	}

	keys := map[string]string{
		activeKeyID: activeSecret,
	}
	for _, item := range c.Container.FlagGlobalSecretKeyring {
		keyID := strings.TrimSpace(item.KeyID)
		secret := strings.TrimSpace(item.Secret)
		if keyID == "" {
			return fmt.Errorf("container.flag_global_secret_keyring contains an empty key_id")
		}
		if secret == "" {
			return fmt.Errorf("container.flag_global_secret_keyring[%s] secret must not be empty", keyID)
		}
		if len(secret) < 32 {
			return fmt.Errorf("container.flag_global_secret_keyring[%s] secret must be at least 32 bytes, current length: %d", keyID, len(secret))
		}
		if existing, exists := keys[keyID]; exists && existing != secret {
			return fmt.Errorf("container.flag_global_secret_keyring[%s] conflicts with another secret", keyID)
		}
		keys[keyID] = secret
	}

	c.Container.FlagGlobalSecretKeyID = activeKeyID
	c.Container.ResolvedFlagSecretKeyID = activeKeyID
	c.Container.ResolvedFlagSecrets = keys
	return nil
}

func loadPersistedContainerFlagGlobalSecret(secretFile string) (string, bool, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(secretFile))
	if cleanPath == "" || cleanPath == "." {
		return "", false, fmt.Errorf("secret file path is empty")
	}
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	secret := strings.TrimSpace(string(data))
	if secret == "" {
		return "", false, fmt.Errorf("secret file %s is empty", cleanPath)
	}
	return secret, true, nil
}

func persistContainerFlagGlobalSecret(secretFile, secret string) error {
	cleanPath := filepath.Clean(strings.TrimSpace(secretFile))
	secret = strings.TrimSpace(secret)
	if cleanPath == "" || cleanPath == "." {
		return fmt.Errorf("secret file path is empty")
	}
	if secret == "" {
		return fmt.Errorf("secret is empty")
	}

	if persistedSecret, exists, err := loadPersistedContainerFlagGlobalSecret(cleanPath); err != nil {
		return err
	} else if exists {
		if persistedSecret != secret {
			return fmt.Errorf("secret file %s already exists with a different value", cleanPath)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o700); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(cleanPath), ".flag-global-secret-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if err := tmpFile.Chmod(0o600); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if _, err := tmpFile.WriteString(secret + "\n"); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, cleanPath); err != nil {
		return err
	}
	return nil
}
