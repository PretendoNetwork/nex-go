package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"fmt"
)

// Kerberos represents a basic Kerberos handling struct
type Kerberos struct {
	Key string
}

// Decrypt decrypts the data of Kerberos response
func (encryption *Kerberos) Decrypt(buffer []byte) []byte {
	if !encryption.Validate(buffer) {
		fmt.Println("INVALID KERB CHECKSUM")
	}

	offset := len(buffer)
	offset = offset + -0x10

	data := buffer[:offset]

	RC4, _ := rc4.NewCipher([]byte(encryption.Key))

	crypted := make([]byte, len(data))
	RC4.XORKeyStream(crypted, data)

	return crypted
}

// Encrypt encrypts the data of Kerberos request
func (encryption *Kerberos) Encrypt(buffer []byte) []byte {
	RC4, _ := rc4.NewCipher([]byte(encryption.Key))

	crypted := make([]byte, len(buffer))
	RC4.XORKeyStream(crypted, buffer)

	cipher := hmac.New(md5.New, []byte(encryption.Key))
	cipher.Write(crypted)
	checksum := cipher.Sum(nil)

	return append(crypted, checksum...)
}

// Validate validates the Kerberos data
func (encryption *Kerberos) Validate(buffer []byte) bool {
	offset := len(buffer)
	offset = offset + -0x10

	data := buffer[:offset]
	checksum := buffer[offset:]

	cipher := hmac.New(md5.New, []byte(encryption.Key))
	cipher.Write(data)
	mac := cipher.Sum(nil)

	return bytes.Equal(mac, checksum)
}

// NewKerberos returns a new instances of basic Kerberos
func NewKerberos(key string) Kerberos {
	return Kerberos{
		Key: key,
	}
}
