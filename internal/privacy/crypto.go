/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package privacy

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/aead/ecdh"
	"github.com/jorrizza/ed2curve25519"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
)

var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPubKeyFormat indicates an error processing what is
	// supposed to be a hex encoded Ed25519 public key
	ErrInvalidPubKeyFormat = errors.New("Invalid PublicKey length/format")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

// pad implements pkcs7 padding of []byte
func pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

// unpad implements pkcs7 de-padding
func unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}

// pubKeyBytes extracts underlying bytes from a publickey
func pubKeyBytes(a crypto.PublicKey) []byte {
	b := a.([32]uint8)
	return b[:]
}

// genKeyedSalt returns a new 128 bit salt from hashing a public key with
// another salt, allowing the generation of key-specific salts when the
// same encryption is performed with multiple keys
func genKeyedSalt(pubkey crypto.PublicKey, salt []byte) ([]byte, error) {
	// compute derived salt from public key and common salt
	b2, err := blake2b.New512(nil)
	if err != nil {
		return nil, err
	}
	b2.Write(pubKeyBytes(pubkey))
	b2.Write(salt)
	derived := b2.Sum(nil)
	// truncate 128 bit
	return derived[:16], nil
}

// deriveKey returns 32 bytes that are suitable for use as 256 bit key
// material, derived from the keys' shared secret
func deriveKey(sessionPriv crypto.PrivateKey, lockPub crypto.PublicKey, salt []byte) ([]byte, error) {
	// compute derived salt from public key and common salt using a keyed
	// hash with the pubkey as key
	b2, err := blake2b.New512(pubKeyBytes(lockPub))
	if err != nil {
		return nil, err
	}
	b2.Write(pubKeyBytes(lockPub))
	b2.Write(salt)
	derived := b2.Sum(nil)

	// compute raw ECDH secret over Curve25519
	kex := ecdh.X25519()
	secret := kex.ComputeSecret(sessionPriv, lockPub)

	// for use as AES-256 key, extract a uniformly distributed key out of
	// the shared secret using a key derivation function
	return argon2.IDKey(secret, derived, 1, 64*1024, 4, 32), nil
}

// decodePKString takes a hex encoded Ed25519 public key and
// returns a fully typed decoded Curve25519 version of the key
func decodePKString(s string) (crypto.PublicKey, error) {
	// decode hex string
	b, err := hex.DecodeString(
		strings.ToLower(s),
	)
	if err != nil {
		return nil, err
	}
	if len(b) != keyLenBytes {
		return nil, ErrInvalidPubKeyFormat
	}
	// cast to ed25519.PublicKey
	ed := make([]byte, ed25519.PublicKeySize)
	copy(ed, b)

	// convert Ed25519 to Curve25519
	cv := ed2curve25519.Ed25519PublicKeyToCurve25519(ed)

	// cast to crypto.PublicKey
	buf := [keyLenBytes]uint8{}
	copy(buf[:], cv)
	pk := interface{}(buf).(crypto.PublicKey)

	// check the decoded value is a valid Curve25519 keypoint
	if err := ecdh.X25519().Check(pk); err != nil {
		return nil, err
	}
	return pk, nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
