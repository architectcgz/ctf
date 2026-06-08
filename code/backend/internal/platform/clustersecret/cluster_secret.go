package clustersecret

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ContainerFlagSecretName = "container_flag_global_secret"
	fingerprintNamespace    = "ctf-platform:container-flag-secret:v1:"
)

var (
	ErrContainerFlagSecretFingerprintMismatch = errors.New("container flag secret fingerprint mismatch")
	ErrContainerFlagSecretKeyIDMismatch       = errors.New("container flag secret key id mismatch")
)

type RuntimeClusterSecret struct {
	Name              string    `gorm:"column:name;primaryKey;size:128"`
	ActiveKeyID       string    `gorm:"column:active_key_id;size:128;not null"`
	ActiveFingerprint string    `gorm:"column:active_fingerprint;size:128;not null"`
	KeyFingerprints   string    `gorm:"column:key_fingerprints;type:text;not null"`
	CreatedAt         time.Time `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time `gorm:"column:updated_at;not null"`
}

type ContainerFlagSecretKeyring struct {
	ActiveKeyID    string
	ActiveSecret   string
	Secrets        map[string]string
	RequiredKeyIDs []string
	AllowRotation  bool
}

func (RuntimeClusterSecret) TableName() string {
	return "runtime_cluster_secrets"
}

func Fingerprint(secret string) string {
	sum := sha256.Sum256([]byte(fingerprintNamespace + strings.TrimSpace(secret)))
	return hex.EncodeToString(sum[:])
}

func RequiredContainerFlagSecretKeyIDs(ctx context.Context, db *gorm.DB) ([]string, error) {
	if db == nil {
		return nil, fmt.Errorf("cluster secret db is nil")
	}
	var rows []struct {
		KeyID string `gorm:"column:key_id"`
	}
	err := db.WithContext(ctx).Raw(`
		SELECT DISTINCT COALESCE(NULLIF(TRIM(flag_key_id), ''), 'default') AS key_id
		FROM instances
		WHERE TRIM(COALESCE(nonce, '')) <> ''
		  AND destroyed_at IS NULL
		  AND status IN ('pending', 'creating', 'running', 'stopping')
		  AND expires_at > ?
	`, time.Now().UTC()).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	keyIDs := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		keyID := strings.TrimSpace(row.KeyID)
		if keyID == "" {
			continue
		}
		if _, ok := seen[keyID]; ok {
			continue
		}
		seen[keyID] = struct{}{}
		keyIDs = append(keyIDs, keyID)
	}
	return keyIDs, nil
}

func RegisterContainerFlagSecret(ctx context.Context, db *gorm.DB, keyID, secret string) error {
	return RegisterContainerFlagSecretKeyring(ctx, db, ContainerFlagSecretKeyring{
		ActiveKeyID:  keyID,
		ActiveSecret: secret,
		Secrets:      map[string]string{strings.TrimSpace(keyID): strings.TrimSpace(secret)},
	})
}

func CheckContainerFlagSecret(ctx context.Context, db *gorm.DB, keyID, secret string) error {
	return CheckContainerFlagSecretKeyring(ctx, db, ContainerFlagSecretKeyring{
		ActiveKeyID:  keyID,
		ActiveSecret: secret,
		Secrets:      map[string]string{strings.TrimSpace(keyID): strings.TrimSpace(secret)},
	})
}

func CheckContainerFlagSecretKeyring(ctx context.Context, db *gorm.DB, keyring ContainerFlagSecretKeyring) error {
	return checkKeyring(ctx, db, ContainerFlagSecretName, keyring)
}

func RegisterContainerFlagSecretKeyring(ctx context.Context, db *gorm.DB, keyring ContainerFlagSecretKeyring) error {
	keyring.ActiveKeyID = strings.TrimSpace(keyring.ActiveKeyID)
	keyring.ActiveSecret = strings.TrimSpace(keyring.ActiveSecret)
	if keyring.Secrets == nil {
		keyring.Secrets = map[string]string{}
	}
	keyring.Secrets[keyring.ActiveKeyID] = keyring.ActiveSecret
	return registerKeyring(ctx, db, ContainerFlagSecretName, keyring)
}

func registerKeyring(ctx context.Context, db *gorm.DB, name string, keyring ContainerFlagSecretKeyring, initialize ...bool) error {
	if db == nil {
		return fmt.Errorf("cluster secret db is nil")
	}
	name = strings.TrimSpace(name)
	keyID := strings.TrimSpace(keyring.ActiveKeyID)
	secret := strings.TrimSpace(keyring.ActiveSecret)
	if name == "" {
		return fmt.Errorf("cluster secret name is empty")
	}
	if keyID == "" {
		return fmt.Errorf("container flag secret key id is empty")
	}
	if secret == "" {
		return fmt.Errorf("container flag secret is empty")
	}

	fingerprint := Fingerprint(secret)
	shouldInitialize := true
	if len(initialize) > 0 {
		shouldInitialize = initialize[0]
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if shouldInitialize {
			now := time.Now().UTC()
			fingerprints, err := keyringFingerprints(keyring)
			if err != nil {
				return err
			}
			fingerprintsJSON, err := marshalFingerprints(fingerprints)
			if err != nil {
				return err
			}
			seed := RuntimeClusterSecret{
				Name:              name,
				ActiveKeyID:       keyID,
				ActiveFingerprint: fingerprint,
				KeyFingerprints:   fingerprintsJSON,
				CreatedAt:         now,
				UpdatedAt:         now,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&seed).Error; err != nil {
				return err
			}
		}

		var current RuntimeClusterSecret
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("name = ?", name).
			First(&current).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("cluster secret %s is not registered", name)
			}
			return err
		}
		if current.ActiveKeyID != keyID {
			return rotateContainerFlagSecretIfAllowed(tx, &current, keyring, fingerprint)
		}
		if current.ActiveFingerprint != fingerprint {
			return fmt.Errorf("%w: key_id=%s", ErrContainerFlagSecretFingerprintMismatch, keyID)
		}
		return validateRequiredKeyFingerprints(&current, keyring)
	})
}

func rotateContainerFlagSecretIfAllowed(tx *gorm.DB, current *RuntimeClusterSecret, keyring ContainerFlagSecretKeyring, nextFingerprint string) error {
	if current == nil {
		return fmt.Errorf("current cluster secret is nil")
	}
	if !keyring.AllowRotation {
		return fmt.Errorf("%w: registered=%s configured=%s", ErrContainerFlagSecretKeyIDMismatch, current.ActiveKeyID, keyring.ActiveKeyID)
	}
	previousSecret := strings.TrimSpace(keyring.Secrets[current.ActiveKeyID])
	if previousSecret == "" || Fingerprint(previousSecret) != current.ActiveFingerprint {
		return fmt.Errorf("previous active container flag secret is not configured for key id %s", current.ActiveKeyID)
	}
	storedFingerprints, err := unmarshalFingerprints(current.KeyFingerprints)
	if err != nil {
		return err
	}
	configuredFingerprints, err := keyringFingerprints(keyring)
	if err != nil {
		return err
	}
	for keyID, fingerprint := range configuredFingerprints {
		storedFingerprints[keyID] = fingerprint
	}
	storedFingerprints[current.ActiveKeyID] = current.ActiveFingerprint
	storedFingerprints[keyring.ActiveKeyID] = nextFingerprint
	nextFingerprintsJSON, err := marshalFingerprints(storedFingerprints)
	if err != nil {
		return err
	}
	next := RuntimeClusterSecret{
		Name:              current.Name,
		ActiveKeyID:       keyring.ActiveKeyID,
		ActiveFingerprint: nextFingerprint,
		KeyFingerprints:   nextFingerprintsJSON,
	}
	if err := validateRequiredKeyFingerprints(&next, keyring); err != nil {
		return err
	}
	return tx.Model(&RuntimeClusterSecret{}).
		Where("name = ?", current.Name).
		Updates(map[string]any{
			"active_key_id":      keyring.ActiveKeyID,
			"active_fingerprint": nextFingerprint,
			"key_fingerprints":   nextFingerprintsJSON,
			"updated_at":         time.Now().UTC(),
		}).Error
}

func checkKeyring(ctx context.Context, db *gorm.DB, name string, keyring ContainerFlagSecretKeyring) error {
	if db == nil {
		return fmt.Errorf("cluster secret db is nil")
	}
	name = strings.TrimSpace(name)
	keyring.ActiveKeyID = strings.TrimSpace(keyring.ActiveKeyID)
	keyring.ActiveSecret = strings.TrimSpace(keyring.ActiveSecret)
	if keyring.Secrets == nil {
		keyring.Secrets = map[string]string{}
	}
	keyring.Secrets[keyring.ActiveKeyID] = keyring.ActiveSecret
	if name == "" {
		return fmt.Errorf("cluster secret name is empty")
	}
	if keyring.ActiveKeyID == "" {
		return fmt.Errorf("container flag secret key id is empty")
	}
	if keyring.ActiveSecret == "" {
		return fmt.Errorf("container flag secret is empty")
	}

	var current RuntimeClusterSecret
	err := db.WithContext(ctx).
		Where("name = ?", name).
		First(&current).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("cluster secret %s is not registered", name)
		}
		return err
	}
	if current.ActiveKeyID != keyring.ActiveKeyID {
		return fmt.Errorf("%w: registered=%s configured=%s", ErrContainerFlagSecretKeyIDMismatch, current.ActiveKeyID, keyring.ActiveKeyID)
	}
	if current.ActiveFingerprint != Fingerprint(keyring.ActiveSecret) {
		return fmt.Errorf("%w: key_id=%s", ErrContainerFlagSecretFingerprintMismatch, keyring.ActiveKeyID)
	}
	return validateRequiredKeyFingerprints(&current, keyring)
}

func keyringFingerprints(keyring ContainerFlagSecretKeyring) (map[string]string, error) {
	fingerprints := make(map[string]string, len(keyring.Secrets)+1)
	for keyID, secret := range keyring.Secrets {
		keyID = strings.TrimSpace(keyID)
		secret = strings.TrimSpace(secret)
		if keyID == "" || secret == "" {
			continue
		}
		fingerprints[keyID] = Fingerprint(secret)
	}
	activeKeyID := strings.TrimSpace(keyring.ActiveKeyID)
	activeSecret := strings.TrimSpace(keyring.ActiveSecret)
	if activeKeyID == "" {
		return nil, fmt.Errorf("container flag secret key id is empty")
	}
	if activeSecret == "" {
		return nil, fmt.Errorf("container flag secret is empty")
	}
	fingerprints[activeKeyID] = Fingerprint(activeSecret)
	return fingerprints, nil
}

func validateRequiredKeyFingerprints(current *RuntimeClusterSecret, keyring ContainerFlagSecretKeyring) error {
	if current == nil {
		return fmt.Errorf("current cluster secret is nil")
	}
	storedFingerprints, err := unmarshalFingerprints(current.KeyFingerprints)
	if err != nil {
		return err
	}
	configuredFingerprints, err := keyringFingerprints(keyring)
	if err != nil {
		return err
	}
	for _, keyID := range keyring.RequiredKeyIDs {
		keyID = strings.TrimSpace(keyID)
		if keyID == "" {
			continue
		}
		configuredFingerprint := configuredFingerprints[keyID]
		if configuredFingerprint == "" {
			return fmt.Errorf("required container flag secret key %s is not configured", keyID)
		}
		storedFingerprint := storedFingerprints[keyID]
		if storedFingerprint == "" {
			return fmt.Errorf("required container flag secret key %s is not registered", keyID)
		}
		if configuredFingerprint != storedFingerprint {
			return fmt.Errorf("%w: key_id=%s", ErrContainerFlagSecretFingerprintMismatch, keyID)
		}
	}
	return nil
}

func marshalFingerprints(fingerprints map[string]string) (string, error) {
	if fingerprints == nil {
		fingerprints = map[string]string{}
	}
	payload, err := json.Marshal(fingerprints)
	if err != nil {
		return "", fmt.Errorf("marshal cluster secret fingerprints: %w", err)
	}
	return string(payload), nil
}

func unmarshalFingerprints(payload string) (map[string]string, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return map[string]string{}, nil
	}
	var fingerprints map[string]string
	if err := json.Unmarshal([]byte(payload), &fingerprints); err != nil {
		return nil, fmt.Errorf("unmarshal cluster secret fingerprints: %w", err)
	}
	if fingerprints == nil {
		fingerprints = map[string]string{}
	}
	return fingerprints, nil
}
