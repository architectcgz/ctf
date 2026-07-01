package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

type OAuthStore struct {
	db     *gorm.DB
	cache  *redislib.Client
	config config.AuthOAuthConfig
}

type oauthClientRecord struct {
	ID                      int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ClientID                string    `gorm:"column:client_id;type:text;not null;uniqueIndex"`
	ClientName              string    `gorm:"column:client_name;type:text;not null"`
	ClientURI               string    `gorm:"column:client_uri;type:text"`
	RedirectURIs            string    `gorm:"column:redirect_uris;type:jsonb;not null"`
	GrantTypes              string    `gorm:"column:grant_types;type:jsonb;not null"`
	ResponseTypes           string    `gorm:"column:response_types;type:jsonb;not null"`
	Scope                   string    `gorm:"column:scope;type:text;not null"`
	TokenEndpointAuthMethod string    `gorm:"column:token_endpoint_auth_method;type:text;not null;default:none"`
	CreatedAt               time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt               time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (oauthClientRecord) TableName() string {
	return "oauth_clients"
}

type oauthConsentRecord struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64      `gorm:"column:user_id;not null;uniqueIndex:uk_oauth_consents_user_client_scope"`
	ClientID  string     `gorm:"column:client_id;type:text;not null;uniqueIndex:uk_oauth_consents_user_client_scope"`
	Scope     string     `gorm:"column:scope;type:text;not null;uniqueIndex:uk_oauth_consents_user_client_scope"`
	GrantedAt time.Time  `gorm:"column:granted_at;not null"`
	ExpiresAt *time.Time `gorm:"column:expires_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at;index:idx_oauth_consents_user_client"`
}

func (oauthConsentRecord) TableName() string {
	return "oauth_consents"
}

func NewOAuthStore(db *gorm.DB, cache *redislib.Client, cfg config.AuthOAuthConfig) *OAuthStore {
	return &OAuthStore{
		db:     db,
		cache:  cache,
		config: cfg,
	}
}

func (s *OAuthStore) SaveClient(ctx context.Context, client authcontracts.OAuthClient) error {
	if err := requireContext(ctx); err != nil {
		return err
	}
	record, err := oauthClientRecordFromContract(client)
	if err != nil {
		return err
	}

	err = s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "client_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"client_name",
			"client_uri",
			"redirect_uris",
			"grant_types",
			"response_types",
			"scope",
			"token_endpoint_auth_method",
			"updated_at",
		}),
	}).Create(&record).Error
	if err != nil {
		return fmt.Errorf("save oauth client: %w", err)
	}
	return nil
}

func (s *OAuthStore) FindClientByID(ctx context.Context, clientID string) (*authcontracts.OAuthClient, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	if clientID == "" {
		return nil, nil
	}

	var record oauthClientRecord
	err := s.db.WithContext(ctx).Where("client_id = ?", clientID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find oauth client: %w", err)
	}

	client, err := record.toContract()
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (s *OAuthStore) UpsertConsent(ctx context.Context, consent authcontracts.OAuthConsent) error {
	if err := requireContext(ctx); err != nil {
		return err
	}
	now := time.Now().UTC()
	if !consent.GrantedAt.IsZero() {
		now = consent.GrantedAt.UTC()
	}
	record := oauthConsentRecord{
		UserID:    consent.UserID,
		ClientID:  consent.ClientID,
		Scope:     consent.Scope,
		GrantedAt: now,
		ExpiresAt: utcTimePtr(consent.ExpiresAt),
		RevokedAt: nil,
	}

	err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "client_id"}, {Name: "scope"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"granted_at",
			"expires_at",
			"revoked_at",
		}),
	}).Create(&record).Error
	if err != nil {
		return fmt.Errorf("upsert oauth consent: %w", err)
	}
	return nil
}

func (s *OAuthStore) FindActiveConsent(ctx context.Context, userID int64, clientID, scope string) (*authcontracts.OAuthConsent, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 || clientID == "" || scope == "" {
		return nil, nil
	}

	var record oauthConsentRecord
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ? AND scope = ?", userID, clientID, scope).
		Where("revoked_at IS NULL").
		Where("expires_at IS NULL OR expires_at > ?", time.Now().UTC()).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find oauth consent: %w", err)
	}
	return &authcontracts.OAuthConsent{
		ID:        record.ID,
		UserID:    record.UserID,
		ClientID:  record.ClientID,
		Scope:     record.Scope,
		GrantedAt: record.GrantedAt.UTC(),
		ExpiresAt: utcTimePtr(record.ExpiresAt),
		RevokedAt: utcTimePtr(record.RevokedAt),
	}, nil
}

func (s *OAuthStore) RevokeConsent(ctx context.Context, userID int64, clientID, scope string) error {
	if err := requireContext(ctx); err != nil {
		return err
	}
	if userID <= 0 || clientID == "" || scope == "" {
		return nil
	}

	now := time.Now().UTC()
	err := s.db.WithContext(ctx).Model(&oauthConsentRecord{}).
		Where("user_id = ? AND client_id = ? AND scope = ? AND revoked_at IS NULL", userID, clientID, scope).
		Update("revoked_at", now).Error
	if err != nil {
		return fmt.Errorf("revoke oauth consent: %w", err)
	}
	return nil
}

func (s *OAuthStore) StoreAuthorizationCode(ctx context.Context, code string, claims authcontracts.OAuthAuthorizationCode, ttl time.Duration) error {
	claims.IssuedAt = claims.IssuedAt.UTC()
	claims.ExpiresAt = claims.ExpiresAt.UTC()
	return s.storeJSON(ctx, s.authorizationCodeKey(code), claims, ttl)
}

func (s *OAuthStore) ConsumeAuthorizationCode(ctx context.Context, code string) (*authcontracts.OAuthAuthorizationCode, error) {
	var claims authcontracts.OAuthAuthorizationCode
	found, err := s.consumeJSON(ctx, s.authorizationCodeKey(code), &claims)
	if err != nil || !found {
		return nil, err
	}
	return &claims, nil
}

func (s *OAuthStore) StoreAccessToken(ctx context.Context, token string, claims authcontracts.OAuthTokenClaims, ttl time.Duration) error {
	claims.IssuedAt = claims.IssuedAt.UTC()
	claims.ExpiresAt = claims.ExpiresAt.UTC()
	return s.storeJSON(ctx, s.accessTokenKey(token), claims, ttl)
}

func (s *OAuthStore) ResolveAccessToken(ctx context.Context, token string) (*authcontracts.OAuthTokenClaims, error) {
	var claims authcontracts.OAuthTokenClaims
	found, err := s.loadJSON(ctx, s.accessTokenKey(token), &claims)
	if err != nil || !found {
		return nil, err
	}
	if !claims.ExpiresAt.After(time.Now().UTC()) {
		_ = s.cache.Del(ctx, s.accessTokenKey(token)).Err()
		return nil, nil
	}
	return &claims, nil
}

func (s *OAuthStore) StoreRefreshToken(ctx context.Context, token string, claims authcontracts.OAuthTokenClaims, ttl time.Duration) error {
	claims.IssuedAt = claims.IssuedAt.UTC()
	claims.ExpiresAt = claims.ExpiresAt.UTC()
	return s.storeJSON(ctx, s.refreshTokenKey(token), claims, ttl)
}

func (s *OAuthStore) ConsumeRefreshToken(ctx context.Context, token string) (*authcontracts.OAuthTokenClaims, error) {
	var claims authcontracts.OAuthTokenClaims
	found, err := s.consumeJSON(ctx, s.refreshTokenKey(token), &claims)
	if err != nil || !found {
		return nil, err
	}
	if !claims.ExpiresAt.After(time.Now().UTC()) {
		return nil, nil
	}
	return &claims, nil
}

func (s *OAuthStore) StoreConsentNonce(ctx context.Context, nonce string, ttl time.Duration) error {
	return s.storeJSON(ctx, s.consentNonceKey(nonce), map[string]string{"nonce": nonce}, ttl)
}

func (s *OAuthStore) ConsumeConsentNonce(ctx context.Context, nonce string) (bool, error) {
	var payload map[string]string
	return s.consumeJSON(ctx, s.consentNonceKey(nonce), &payload)
}

func (s *OAuthStore) storeJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := requireContext(ctx); err != nil {
		return err
	}
	if key == "" || ttl <= 0 {
		return apperror.ErrInternal.WithCause(fmt.Errorf("oauth redis key and ttl are required"))
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if err := s.cache.Set(ctx, key, payload, ttl).Err(); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *OAuthStore) loadJSON(ctx context.Context, key string, dest any) (bool, error) {
	if err := requireContext(ctx); err != nil {
		return false, err
	}
	payload, err := s.cache.Get(ctx, key).Result()
	if errors.Is(err, redislib.Nil) {
		return false, nil
	}
	if err != nil {
		return false, apperror.ErrInternal.WithCause(err)
	}
	if err := json.Unmarshal([]byte(payload), dest); err != nil {
		return false, apperror.ErrInternal.WithCause(err)
	}
	return true, nil
}

func (s *OAuthStore) consumeJSON(ctx context.Context, key string, dest any) (bool, error) {
	if err := requireContext(ctx); err != nil {
		return false, err
	}
	payload, err := s.cache.GetDel(ctx, key).Result()
	if errors.Is(err, redislib.Nil) {
		return false, nil
	}
	if err != nil {
		return false, apperror.ErrInternal.WithCause(err)
	}
	if err := json.Unmarshal([]byte(payload), dest); err != nil {
		return false, apperror.ErrInternal.WithCause(err)
	}
	return true, nil
}

func (s *OAuthStore) authorizationCodeKey(code string) string {
	return fmt.Sprintf("%s:code:%s", s.config.RedisKeyPrefix, hashOAuthSecret(code))
}

func (s *OAuthStore) accessTokenKey(token string) string {
	return fmt.Sprintf("%s:access:%s", s.config.RedisKeyPrefix, hashOAuthSecret(token))
}

func (s *OAuthStore) refreshTokenKey(token string) string {
	return fmt.Sprintf("%s:refresh:%s", s.config.RedisKeyPrefix, hashOAuthSecret(token))
}

func (s *OAuthStore) consentNonceKey(nonce string) string {
	return fmt.Sprintf("%s:consent_nonce:%s", s.config.RedisKeyPrefix, hashOAuthSecret(nonce))
}

func hashOAuthSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func oauthClientRecordFromContract(client authcontracts.OAuthClient) (oauthClientRecord, error) {
	redirectURIs, err := marshalStringList(client.RedirectURIs)
	if err != nil {
		return oauthClientRecord{}, err
	}
	grantTypes, err := marshalStringList(client.GrantTypes)
	if err != nil {
		return oauthClientRecord{}, err
	}
	responseTypes, err := marshalStringList(client.ResponseTypes)
	if err != nil {
		return oauthClientRecord{}, err
	}
	now := time.Now().UTC()
	record := oauthClientRecord{
		ID:                      client.ID,
		ClientID:                client.ClientID,
		ClientName:              client.ClientName,
		ClientURI:               client.ClientURI,
		RedirectURIs:            redirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		Scope:                   client.Scope,
		TokenEndpointAuthMethod: client.TokenEndpointAuthMethod,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if record.TokenEndpointAuthMethod == "" {
		record.TokenEndpointAuthMethod = "none"
	}
	return record, nil
}

func (r oauthClientRecord) toContract() (authcontracts.OAuthClient, error) {
	redirectURIs, err := unmarshalStringList(r.RedirectURIs)
	if err != nil {
		return authcontracts.OAuthClient{}, err
	}
	grantTypes, err := unmarshalStringList(r.GrantTypes)
	if err != nil {
		return authcontracts.OAuthClient{}, err
	}
	responseTypes, err := unmarshalStringList(r.ResponseTypes)
	if err != nil {
		return authcontracts.OAuthClient{}, err
	}
	return authcontracts.OAuthClient{
		ID:                      r.ID,
		ClientID:                r.ClientID,
		ClientName:              r.ClientName,
		ClientURI:               r.ClientURI,
		RedirectURIs:            redirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		Scope:                   r.Scope,
		TokenEndpointAuthMethod: r.TokenEndpointAuthMethod,
		CreatedAt:               r.CreatedAt.UTC(),
		UpdatedAt:               r.UpdatedAt.UTC(),
	}, nil
}

func marshalStringList(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return "", apperror.ErrInternal.WithCause(err)
	}
	return string(payload), nil
}

func unmarshalStringList(raw string) ([]string, error) {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return values, nil
}

func utcTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}
