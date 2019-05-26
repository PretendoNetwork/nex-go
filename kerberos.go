package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
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
func NewKerberos(pid uint32) Kerberos {
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, pid)
	for i := 0; uint32(i) < 65000+pid%1024; i++ {
		key = MD5Hash(key)
	}
	return Kerberos{
		Key: string(binary.LittleEndian.Uint32(key)),
	}
}

type Ticket struct {
	SessionKey []byte
	PID        uint32
	TicketData []byte
}

func NewTicket(session_key []byte, pid uint32, ticketdat []byte) Ticket {
	return Ticket{
		SessionKey: session_key,
		PID:        pid,
		TicketData: ticketdat,
	}
}

func (t Ticket) Encrypt(pid uint32) []byte {
	kerb := NewKerberos(pid)
	outputstr := NewOutputStream()
	outputstr.Write(t.SessionKey)
	outputstr.UInt32LE(t.PID)
	outputstr.Buffer(t.TicketData)
	return kerb.Encrypt(outputstr.Bytes())
}
