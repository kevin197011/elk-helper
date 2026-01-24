// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const encryptedPrefix = "enc:"

func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, encryptedPrefix)
}

// MaybeEncrypt encrypts plaintext when key is provided. If key is nil/empty, returns value as-is.
func MaybeEncrypt(value string, key []byte) (string, error) {
	if value == "" || IsEncrypted(value) || len(key) == 0 {
		return value, nil
	}
	return Encrypt(value, key)
}

// MaybeDecrypt decrypts encrypted values when key is provided.
// If value is plaintext, returns value as-is.
func MaybeDecrypt(value string, key []byte) (string, error) {
	if value == "" || !IsEncrypted(value) {
		return value, nil
	}
	if len(key) == 0 {
		return "", errors.New("encrypted value present but APP_ENCRYPTION_KEY is not configured")
	}
	return Decrypt(value, key)
}

// Encrypt encrypts plaintext into "enc:<base64(nonce|ciphertext)>" using AES-256-GCM.
func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes (got %d)", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, ciphertext...)
	return encryptedPrefix + base64.RawStdEncoding.EncodeToString(payload), nil
}

// Decrypt decrypts "enc:<base64(nonce|ciphertext)>" values encrypted by Encrypt.
func Decrypt(value string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes (got %d)", len(key))
	}
	if !IsEncrypted(value) {
		return value, nil
	}

	raw := strings.TrimPrefix(value, encryptedPrefix)
	payload, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil {
		return "", fmt.Errorf("decode encrypted payload: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return "", errors.New("encrypted payload too short")
	}

	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}
