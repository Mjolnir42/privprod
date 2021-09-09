/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package flowdata // import "github.com/mjolnir42/privprod/internal/flowdata"

// Plaintext contains the sensitive information for encryption
type Plaintext struct {
	RecordID   string `json:"RecordID"`
	SrcAddress string `json:"SrcAddress"`
	DstAddress string `json:"DstAddress"`
}

// ExportPlaintext returns the record's data that will become encrypted
func (r Record) ExportPlaintext() Plaintext {
	return Plaintext{
		RecordID:   r.RecordID,
		SrcAddress: r.SrcAddress,
		DstAddress: r.DstAddress,
	}
}

// EncryptedRecord is the struct for exporting encrypted data, with the
// value field containing an encrypted serialization of a plaintext struct
type EncryptedRecord struct {
	RecordID     string `json:"RecordID"`
	SessionKeyID string `json:"keyID"`
	Salt         string `json:"salt"`
	Signature    string `json:"signature"`
	Value        string `json:"value"`
	RawSalt      []byte `json:"-"`
	RawSignature []byte `json:"-"`
	RawValue     []byte `json:"-"`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
