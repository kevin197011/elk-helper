// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package sso

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/kk/elk-helper/backend/internal/models"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// OIDCConfig holds OIDC IdP settings stored in config_json.
type OIDCConfig struct {
	Issuer        string   `json:"issuer"`
	ClientID      string   `json:"client_id"`
	ClientSecret  string   `json:"client_secret"`
	RedirectURL   string   `json:"redirect_url"`
	Scopes        []string `json:"scopes"`
	UsernameClaim string   `json:"username_claim"`
	SkipTLSVerify bool     `json:"skip_tls_verify"`
}

// ProviderInfo is returned to the login page (enabled, loaded providers only).
type ProviderInfo struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// AdminProviderItem includes parsed config for the admin UI.
type AdminProviderItem struct {
	ID      uint                   `json:"id"`
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

type oidcRuntime struct {
	id          uint
	name        string
	cfg         OIDCConfig
	provider    *oidc.Provider
	verifier    *oidc.IDTokenVerifier
	oauthConfig *oauth2.Config
}

// Service manages SSO providers and OIDC login runtime.
type Service struct {
	db *gorm.DB

	mu      sync.RWMutex
	oidcSet map[uint]*oidcRuntime
	all     []models.SSOProvider
}

// NewService creates an SSO service and loads enabled providers.
func NewService(db *gorm.DB) *Service {
	s := &Service{db: db}
	if err := s.Reload(); err != nil {
		slog.Warn("SSO provider reload failed", "error", err)
	}
	return s
}

// Reload rebuilds in-memory OIDC runtimes from the database.
func (s *Service) Reload() error {
	var list []models.SSOProvider
	if err := s.db.Order("id ASC").Find(&list).Error; err != nil {
		return err
	}

	newOIDC := make(map[uint]*oidcRuntime)
	for _, p := range list {
		if !p.Enabled || p.Type != models.SSOTypeOIDC {
			continue
		}
		rt, err := buildOIDCRuntime(p)
		if err != nil {
			slog.Warn("OIDC provider init failed", "id", p.ID, "name", p.Name, "error", err)
			continue
		}
		newOIDC[p.ID] = rt
	}

	s.mu.Lock()
	s.oidcSet = newOIDC
	s.all = list
	s.mu.Unlock()
	return nil
}

// ListProviders returns enabled OIDC providers that loaded successfully.
func (s *Service) ListProviders() ([]ProviderInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]ProviderInfo, 0)
	for _, p := range s.all {
		if !p.Enabled || p.Type != models.SSOTypeOIDC {
			continue
		}
		if _, ok := s.oidcSet[p.ID]; !ok {
			continue
		}
		infos = append(infos, ProviderInfo{ID: p.ID, Type: p.Type, Name: p.Name})
	}
	return infos, nil
}

// BuildOIDCAuthURL returns the IdP authorization URL for the given provider.
func (s *Service) BuildOIDCAuthURL(id uint, state string) (string, error) {
	s.mu.RLock()
	rt := s.oidcSet[id]
	s.mu.RUnlock()
	if rt == nil {
		return "", errors.New("OIDC provider not found or not enabled")
	}
	return rt.oauthConfig.AuthCodeURL(state), nil
}

// ExchangeOIDC exchanges the authorization code and returns the username claim.
func (s *Service) ExchangeOIDC(ctx context.Context, id uint, code string) (string, error) {
	s.mu.RLock()
	rt := s.oidcSet[id]
	s.mu.RUnlock()
	if rt == nil {
		return "", errors.New("OIDC provider not found or not enabled")
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	if rt.cfg.SkipTLSVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // admin-configured for private IdP
		}
	}
	ctx = oidc.ClientContext(ctx, httpClient)

	tok, err := rt.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("exchange code: %w", err)
	}
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", errors.New("missing id_token in token response")
	}
	idToken, err := rt.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", fmt.Errorf("verify id_token: %w", err)
	}
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("decode claims: %w", err)
	}
	username := pickStringClaim(claims, rt.cfg.UsernameClaim, "preferred_username", "email", "sub")
	if username == "" {
		return "", errors.New("could not extract username from id_token")
	}
	return username, nil
}

// ListAll returns all SSO providers for admin (including disabled).
func (s *Service) ListAll() ([]AdminProviderItem, error) {
	var list []models.SSOProvider
	if err := s.db.Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]AdminProviderItem, 0, len(list))
	for _, p := range list {
		var cfg map[string]interface{}
		_ = json.Unmarshal([]byte(p.ConfigJSON), &cfg)
		items = append(items, AdminProviderItem{
			ID:      p.ID,
			Type:    p.Type,
			Name:    p.Name,
			Enabled: p.Enabled,
			Config:  cfg,
		})
	}
	return items, nil
}

// GetByID loads a provider by ID.
func (s *Service) GetByID(id uint) (*models.SSOProvider, error) {
	var p models.SSOProvider
	if err := s.db.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("SSO provider not found")
		}
		return nil, err
	}
	return &p, nil
}

// CreateProvider inserts a new SSO provider.
func (s *Service) CreateProvider(p *models.SSOProvider) (*models.SSOProvider, error) {
	if err := s.db.Create(p).Error; err != nil {
		return nil, err
	}
	_ = s.Reload()
	return p, nil
}

// UpdateProvider updates an existing provider.
func (s *Service) UpdateProvider(p *models.SSOProvider) error {
	if err := s.db.Save(p).Error; err != nil {
		return err
	}
	return s.Reload()
}

// SetEnabled toggles provider enabled flag.
func (s *Service) SetEnabled(id uint, enabled bool) error {
	if err := s.db.Model(&models.SSOProvider{}).Where("id = ?", id).Update("enabled", enabled).Error; err != nil {
		return err
	}
	return s.Reload()
}

// DeleteProvider soft-deletes a provider.
func (s *Service) DeleteProvider(id uint) error {
	if err := s.db.Delete(&models.SSOProvider{}, id).Error; err != nil {
		return err
	}
	return s.Reload()
}

// ValidateProviderConfig checks required OIDC fields in admin payload.
func ValidateProviderConfig(providerType string, cfg map[string]interface{}) string {
	if providerType != models.SSOTypeOIDC {
		return "type must be oidc"
	}
	for _, k := range []string{"issuer", "client_id", "client_secret", "redirect_url"} {
		if v, _ := cfg[k].(string); strings.TrimSpace(v) == "" {
			return "OIDC missing required field: " + k
		}
	}
	return ""
}

// GenerateState creates a random OIDC state value.
func GenerateState() (string, error) {
	b := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func buildOIDCRuntime(p models.SSOProvider) (*oidcRuntime, error) {
	var cfg OIDCConfig
	if err := json.Unmarshal([]byte(p.ConfigJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid oidc config: %w", err)
	}
	if cfg.Issuer == "" || cfg.ClientID == "" || cfg.RedirectURL == "" {
		return nil, errors.New("issuer, client_id, and redirect_url are required")
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}
	if cfg.UsernameClaim == "" {
		cfg.UsernameClaim = "preferred_username"
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	if cfg.SkipTLSVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
	}
	ctx := oidc.ClientContext(context.Background(), httpClient)

	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("discover issuer: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
	}
	return &oidcRuntime{
		id:          p.ID,
		name:        p.Name,
		cfg:         cfg,
		provider:    provider,
		verifier:    verifier,
		oauthConfig: oauthCfg,
	}, nil
}

func pickStringClaim(claims map[string]interface{}, names ...string) string {
	for _, n := range names {
		if v, ok := claims[n].(string); ok && v != "" {
			return v
		}
	}
	return ""
}
