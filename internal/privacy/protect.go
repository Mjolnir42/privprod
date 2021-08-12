/*-
 * Copyright (c) 2021, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

// Package privacy implements the internal function library for the
// privacy protector daemon.
package privacy // import "github.com/mjolnir42/privprod/internal/privacy"

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/aead/ecdh"
	"github.com/mjolnir42/erebos"
	"github.com/mjolnir42/flowdata"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/poly1305"
)

type Protector struct {
	Num          int
	Input        chan *erebos.Transport
	Shutdown     chan struct{}
	Death        chan error
	dispatch     chan<- *sarama.ProducerMessage
	producer     sarama.AsyncProducer
	topic        string
	topicIOC     string
	topicSKey    string
	topicENC     string
	sessionKeyID string
	sessionKey   []byte
}

func (p *Protector) Start() {
	if len(Handlers) == 0 {
		p.Death <- fmt.Errorf(`Incorrectly set handlers`)
		<-p.Shutdown
		return
	}

	p.topic = os.Getenv(`KAFKA_PRODUCER_TOPIC_DATA`)
	p.topicIOC = os.Getenv(`KAFKA_PRODUCER_TOPIC_IOC`)
	p.topicSKey = os.Getenv(`KAFKA_PRODUCER_TOPIC_SESSION`)
	p.topicENC = os.Getenv(`KAFKA_PRODUCER_TOPIC_ENCRYPTED`)

	config := sarama.NewConfig()
	config.Net.KeepAlive = 3 * time.Second
	config.Producer.RequiredAcks = sarama.WaitForLocal

	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Retry.Max = 3
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.ClientID = `privacyprotector`

	brokers := os.Getenv(`KAFKA_BROKER_PEERS`)
	var err error
	p.producer, err = sarama.NewAsyncProducer(strings.Split(brokers, `,`), config)
	if p.assert(err) {
		return
	}
	p.dispatch = p.producer.Input()

	p.InitCrypto()

	p.run()
}

func (p *Protector) InitCrypto() {
	var err error
	var unlockPKOneIn, unlockPKTwoIn string
	var unlockPKOne, unlockPKTwo crypto.PublicKey
	key := flowdata.Key{}

	// fetch public keys to lock session key persistance with
	unlockPKOneIn = os.Getenv(`UNLOCK_PUBLICKEY_ONE`)
	unlockPKOne, err = decodePKString(unlockPKOneIn)
	if p.assert(err) {
		return
	}

	unlockPKTwoIn = os.Getenv(`UNLOCK_PUBLICKEY_TWO`)
	unlockPKTwo, err = decodePKString(unlockPKTwoIn)
	if p.assert(err) {
		return
	}

	// generate common portion of the salt
	key.Salt = make([]byte, saltLenBytes)
	_, err = rand.Read(key.Salt)
	if p.assert(err) {
		return
	}

	// generate symmetric session key
	// BUG: session key is not rotated every 24h
	p.sessionKey = make([]byte, keyLenBytes)
	key.Value = make([]byte, keyLenBytes)
	_, err = rand.Read(p.sessionKey)
	if p.assert(err) {
		return
	}
	// copy the plaintext sessionkey into the output buffer, which will
	// be used in-place for encryption
	copy(key.Value, p.sessionKey)

	// set SessionKeyID
	p.sessionKeyID = uuid.NewV4().String()
	key.ID = p.sessionKeyID

	// generate ephemeral asymmetric key used for publishing the
	// generated session key
	var priv crypto.PrivateKey
	var pub crypto.PublicKey
	priv, pub, err = ecdh.X25519().GenerateKey(nil)
	if p.assert(err) {
		return
	}
	key.PublicKey = pubKeyBytes(pub)

	// encrypt session key with unlock keys
	for _, pk := range []crypto.PublicKey{unlockPKOne, unlockPKTwo} {
		slt, err := genKeyedSalt(pk, key.Salt)
		if p.assert(err) {
			return
		}
		skey, err := deriveKey(priv, pk, key.Salt)
		if p.assert(err) {
			return
		}
		// setup AES encryption
		block, err := aes.NewCipher(skey)
		if p.assert(err) {
			return
		}
		// XOR session key with generated keystream in OFB mode
		cipher.NewOFB(
			block,
			slt,
		).XORKeyStream(
			key.Value,
			key.Value,
		)
	}
	// unlink private session key
	priv = nil

	key.Serialize()
	err = key.CalculateMAC()
	if p.assert(err) {
		return
	}

	// publish encrypted version of session key
	jb, err := json.Marshal(&key)
	if p.assert(err) {
		return
	}

	p.dispatch <- &sarama.ProducerMessage{
		Topic: p.topicSKey,
		Value: sarama.ByteEncoder(jb),
	}
}

func (p *Protector) run() {
	// required during shutdown
	inputEmpty := false
	errorEmpty := false
	successEmpty := false
	producerClosed := false

runloop:
	for {
		select {
		case <-p.Shutdown:
			goto drainloop
		case msg := <-p.producer.Errors():
			log.Printf("Producer error: %s\n",
				msg.Err.Error(),
			)
		case <-p.producer.Successes():
			// noop, no tracking of supplied messages
			continue runloop
		case msg := <-p.Input:
			if msg == nil {
				continue runloop
			}
			p.process(msg)
		}
	}
	p.producer.Close()
	return

	// drain the input channel
drainloop:
	for {
		select {
		case msg := <-p.Input:
			if msg == nil {
				inputEmpty = true

				if !producerClosed {
					p.producer.Close()
					producerClosed = true
				}

				// channels are closed
				if inputEmpty && errorEmpty && successEmpty {
					break drainloop
				}
				continue drainloop
			}
			p.process(msg)
		case msg := <-p.producer.Errors():
			if msg == nil {
				errorEmpty = true

				// channels are closed
				if inputEmpty && errorEmpty && successEmpty {
					break drainloop
				}
				continue drainloop
			}
		case msg := <-p.producer.Successes():
			if msg == nil {
				successEmpty = true

				// channels are closed
				if inputEmpty && errorEmpty && successEmpty {
					break drainloop
				}
				continue drainloop
			}
		}
	}
}

func (p *Protector) process(msg *erebos.Transport) {
	// flowdata decode
	decoded := flowdata.Message{}
	if err := json.Unmarshal(msg.Value, decoded); err != nil {
		logrus.Errorln(err)
		return
	}

recordloop:
	for record := range decoded.Convert() {
		storeEncrypted := false

		// copy of the struct must be done after the RecordID has been
		// generated to be able to track the pseudotext<>ciphertext
		// relationship, but also before any pseudomization is being done
		record.RecordID = uuid.NewV4().String()
		original := record.Copy()

		src := net.ParseIP(record.SrcAddress).To16()
		dst := net.ParseIP(record.DstAddress).To16()

		if discard(src) || discard(dst) {
			continue recordloop
		}

		if isPrivate(src) && isEmployeePriv(src) {
			storeEncrypted = true

			hash, _ := blake2b.New256(pseudoKey)
			hash.Write(dataPad)
			hash.Write([]byte(src))
			record.SrcAddress = fmtEmployeePriv(hash.Sum(nil))
		} else if isCompany(src) && isEmployeePub(src) {
			storeEncrypted = true

			hash, _ := blake2b.New256(pseudoKey)
			hash.Write(dataPad)
			hash.Write([]byte(src))
			record.SrcAddress = fmtEmployeePub(hash.Sum(nil))
		} else if isPublic(src) {
			storeEncrypted = true

			go p.publishIOC(record.ToIOC(src.String()))
			hash, _ := blake2b.New256(pseudoKey)
			hash.Write(dataPad)
			hash.Write([]byte(src))
			record.SrcAddress = fmtCustomer(hash.Sum(nil))
		}

		if isPrivate(dst) && isEmployeePriv(dst) {
			storeEncrypted = true

			hash, _ := blake2b.New256(pseudoKey)
			hash.Write(dataPad)
			hash.Write([]byte(dst))
			record.DstAddress = fmtEmployeePriv(hash.Sum(nil))
		} else if isCompany(dst) && isEmployeePub(dst) {
			storeEncrypted = true

			hash, _ := blake2b.New256(pseudoKey)
			hash.Write(dataPad)
			hash.Write([]byte(dst))
			record.DstAddress = fmtEmployeePub(hash.Sum(nil))
		} else if isPublic(dst) {
			storeEncrypted = true
			go p.publishIOC(record.ToIOC(dst.String()))

			hash, _ := blake2b.New256(pseudoKey)
			hash.Write(dataPad)
			hash.Write([]byte(dst))
			record.DstAddress = fmtCustomer(hash.Sum(nil))
		}

		jbytes, err := json.Marshal(&record)
		if err != nil {
			continue recordloop
		}

		p.dispatch <- &sarama.ProducerMessage{
			Topic: p.topic,
			Value: sarama.ByteEncoder(jbytes),
		}

		if storeEncrypted {
			go p.encrypt(original.ExportPlaintext())
		}
	}
}

func (p *Protector) InputChannel() chan *erebos.Transport {
	return p.Input
}

func (p *Protector) ShutdownChannel() chan struct{} {
	return p.Shutdown
}

func (p *Protector) publishIOC(ioc flowdata.IOC) {
	jb, err := json.Marshal(&ioc)
	if p.assert(err) {
		return
	}

	p.dispatch <- &sarama.ProducerMessage{
		Topic: p.topicIOC,
		Value: sarama.ByteEncoder(jb),
	}
}

func (p *Protector) encrypt(input flowdata.Plaintext) {
	// binary encoding of received input struct
	var plain bytes.Buffer
	var raw, padded, jb []byte
	var block cipher.Block
	var err error
	var b2 hash.Hash

	encoder := gob.NewEncoder(&plain)
	err = encoder.Encode(input)
	if p.assert(err) {
		return
	}
	raw = plain.Bytes()

	// setup encrypted struct
	ctxt := flowdata.EncryptedRecord{}
	ctxt.RecordID = input.RecordID
	ctxt.SessionKeyID = p.sessionKeyID

	// generate salt
	ctxt.RawSalt = make([]byte, saltLenBytes)
	_, err = rand.Read(ctxt.RawSalt)
	if p.assert(err) {
		return
	}
	ctxt.Salt = base64.StdEncoding.EncodeToString(ctxt.RawSalt)

	// setup encryption cipher
	block, err = aes.NewCipher(p.sessionKey)
	if p.assert(err) {
		return
	}
	mode := cipher.NewCBCEncrypter(block, ctxt.RawSalt)

	// pkcs7 padding the blocksize
	padded, err = pad(raw, mode.BlockSize())
	if p.assert(err) {
		return
	}

	//
	ctxt.RawValue = make([]byte, len(padded), len(padded))
	mode.CryptBlocks(ctxt.RawValue, padded)
	ctxt.Value = base64.StdEncoding.EncodeToString(ctxt.RawValue)

	// calculate Poly1305 signature
	b2, err = blake2b.New256(nil)
	if p.assert(err) {
		return
	}
	// calculate hash of the ciphertext for use as authentication key
	b2.Write([]byte(ctxt.Value))
	var polyKey [32]byte
	var polyMAC [16]byte
	copy(polyKey[:], b2.Sum(nil))

	// calculate MAC over the output encoded fields instead of the raw
	// []byte fields, so that receiver verification can work directly on
	// received data
	poly1305.Sum(
		&polyMAC, // output buffer
		bytes.Join(
			[][]byte{
				[]byte(ctxt.RecordID),
				[]byte(ctxt.SessionKeyID),
				[]byte(ctxt.Salt),
				[]byte(ctxt.Value),
			},
			nil, // separator
		),
		&polyKey, // authentication key
	)
	ctxt.RawSignature = polyMAC[:]
	ctxt.Signature = base64.StdEncoding.EncodeToString(ctxt.RawSignature)

	// publish encrypted record
	jb, err = json.Marshal(&ctxt)
	if p.assert(err) {
		return
	}

	p.dispatch <- &sarama.ProducerMessage{
		Topic: p.topicENC,
		Value: sarama.ByteEncoder(jb),
	}
}

func (p *Protector) assert(err error) bool {
	if err != nil {
		p.Death <- err
		<-p.Shutdown
		return true
	}
	return false
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
