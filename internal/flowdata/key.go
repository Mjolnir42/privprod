/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package flowdata // import "github.com/mjolnir42/privprod/internal/flowdata"

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/poly1305"
)

// Key represents a session keyfile record used to encrypt records
type Key struct {
	ID            string `json:"keyID"`
	SlotMap       uint16 `json:"-"`
	Value         []byte `json:"-"`
	Salt          []byte `json:"-"`
	PublicKey     []byte `json:"-"`
	ExportSlotMap int    `json:"decryptionSlotMap"`
	ExportValue   string `json:"encryptedKey"`
	ExportSalt    string `json:"salt"`
	ExportPubKey  string `json:"publicPeerKey"`
	ExportSig     string `json:"signature"`
}

// Serialize encodes the embedded information into new fields in a JSON
// exportable representation
func (k *Key) Serialize() {
	k.ExportSlotMap = int(k.SlotMap)
	k.SlotMap = 0

	k.ExportValue = base64.StdEncoding.EncodeToString(k.Value)
	k.Value = nil

	k.ExportSalt = base64.StdEncoding.EncodeToString(k.Salt)
	k.Salt = nil

	k.ExportPubKey = base64.StdEncoding.EncodeToString(k.PublicKey)
	k.PublicKey = nil
}

// CalculateMAC computes the Poly1305 MAC signature over the serialized
// export values
func (k *Key) CalculateMAC() error {
	var polyKey [32]byte
	var polyMAC [16]byte

	b2h, err := blake2b.New256(nil)
	if err != nil {
		return err
	}
	b2h.Write([]byte(k.ExportValue))
	copy(polyKey[:], b2h.Sum(nil))

	slot := make([]byte, 8)
	binary.LittleEndian.PutUint64(slot, uint64(k.ExportSlotMap))

	poly1305.Sum(
		&polyMAC,
		bytes.Join(
			[][]byte{
				[]byte(k.ExportValue),
				[]byte(k.ExportSalt),
				[]byte(k.ExportPubKey),
				slot,
			},
			nil,
		),
		&polyKey,
	)
	k.ExportSig = base64.StdEncoding.EncodeToString(polyMAC[:])
	return nil
}

// VerifyMAC computes the Poly1305 MAC signature over the serialized
// export values and compares it with the contained signature
func (k *Key) VerifyMAC() (bool, error) {
	var polyKey [32]byte
	var polyMAC [16]byte

	b2h, err := blake2b.New256(nil)
	if err != nil {
		return false, err
	}
	b2h.Write([]byte(k.ExportValue))
	copy(polyKey[:], b2h.Sum(nil))

	sig, err := base64.StdEncoding.DecodeString(k.ExportSig)
	if err != nil {
		return false, err
	}
	copy(polyMAC[:], sig)

	slot := make([]byte, 8)
	binary.LittleEndian.PutUint64(slot, uint64(k.ExportSlotMap))

	return poly1305.Verify(
		&polyMAC,
		bytes.Join(
			[][]byte{
				[]byte(k.ExportValue),
				[]byte(k.ExportSalt),
				[]byte(k.ExportPubKey),
				slot,
			},
			nil,
		),
		&polyKey,
	), nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
