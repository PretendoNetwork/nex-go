package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"fmt"
)

// KerberosEncryption is used to encrypt/decrypt using Kerberos
type KerberosEncryption struct {
	key    []byte
	cipher *rc4.Cipher
}

// Ticket represents a Kerberos authentication ticket
type Ticket struct {
	sessionKey []byte
	serverPID  uint32
	ticketData []byte
}

// TicketData contains the encrypted ticket info and an optional key used for deriving the encryption key for TicketInfo
type TicketData struct {
	ticketKey  []byte
	ticketInfo []byte
}

// TicketInfo contains the actual data of the ticket
type TicketInfo struct {
	datetime   uint64
	userPID    uint32
	sessionKey []byte
}

// Encrypt will encrypt the given data using Kerberos
func (encryption *KerberosEncryption) Encrypt(buffer []byte) []byte {
	encrypted := make([]byte, len(buffer))
	encryption.cipher.XORKeyStream(encrypted, buffer)

	mac := hmac.New(md5.New, []byte(encryption.key))
	mac.Write(encrypted)
	hmac := mac.Sum(nil)

	return append(encrypted, hmac...)
}

// Decrypt will decrypt the given data using Kerberos
func (encryption *KerberosEncryption) Decrypt(buffer []byte) []byte {
	if !encryption.Validate(buffer) {
		fmt.Println("INVALID KERB CHECKSUM")
	}

	offset := len(buffer)
	offset = offset + -0x10

	encrypted := buffer[:offset]

	decrypted := make([]byte, len(encrypted))
	encryption.cipher.XORKeyStream(decrypted, encrypted)

	return decrypted
}

// Validate will check the HMAC of the encrypted data
func (encryption *KerberosEncryption) Validate(buffer []byte) bool {
	offset := len(buffer)
	offset = offset + -0x10

	data := buffer[:offset]
	checksum := buffer[offset:]

	cipher := hmac.New(md5.New, []byte(encryption.key))
	cipher.Write(data)
	mac := cipher.Sum(nil)

	return bytes.Equal(mac, checksum)
}

// NewKerberosEncryption returns a new KerberosEncryption instance
func NewKerberosEncryption(key []byte) *KerberosEncryption {
	cipher, _ := rc4.NewCipher(key)

	return &KerberosEncryption{
		key:    key,
		cipher: cipher,
	}
}
