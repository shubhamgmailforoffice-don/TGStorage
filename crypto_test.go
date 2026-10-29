
package main

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	original := []byte("TGStorage encryption test")

	encrypted, err := Encrypt(key, original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(encrypted, original) {
		t.Fatal("encrypted data should differ from plaintext")
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, original) {
		t.Fatal("decrypted text does not match original")
	}
}

func TestTamperedDataFails(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	encrypted, err := Encrypt(key, []byte("test data"))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	encrypted[len(encrypted)-1] ^= 0x01

	if _, err := Decrypt(key, encrypted); err == nil {
		t.Fatal("expected decryption to fail after tampering")
	}
}


func TestGenerateKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if len(key1) != 32 {
		t.Fatalf("expected 32-byte AES-256 key, got %d bytes", len(key1))
	}

	if bytes.Equal(key1, key2) {
		t.Fatal("two generated keys should not be identical")
	}
}


func TestWrapAndUnwrapFileKey(t *testing.T) {
	masterKey, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey failed: %v", err)
	}

	fileKey, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	wrapped, err := WrapFileKey(masterKey, fileKey)
	if err != nil {
		t.Fatalf("WrapFileKey failed: %v", err)
	}

	unwrapped, err := UnwrapFileKey(masterKey, wrapped)
	if err != nil {
		t.Fatalf("UnwrapFileKey failed: %v", err)
	}

	if !bytes.Equal(fileKey, unwrapped) {
		t.Fatal("unwrapped key does not match original file key")
	}
}

func TestUnwrapWithWrongMasterKeyFails(t *testing.T) {
	masterKey, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey failed: %v", err)
	}

	wrongMasterKey, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey failed: %v", err)
	}

	fileKey, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	wrapped, err := WrapFileKey(masterKey, fileKey)
	if err != nil {
		t.Fatalf("WrapFileKey failed: %v", err)
	}

	if _, err := UnwrapFileKey(wrongMasterKey, wrapped); err == nil {
		t.Fatal("expected unwrapping to fail with the wrong master key")
	}
}

