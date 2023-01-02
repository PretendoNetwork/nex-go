package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"math/rand"
)

// KerberosEncryption is used to encrypt/decrypt using Kerberos
type KerberosEncryption struct {
	key    []byte
	cipher *rc4.Cipher
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
		logger.Error("Keberos hmac validation failed")
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

// Ticket represents a Kerberos authentication ticket
type Ticket struct {
	sessionKey   []byte
	targetPID    uint32
	internalData []byte
}

// SessionKey returns the Tickets session key
func (ticket *Ticket) SessionKey() []byte {
	return ticket.sessionKey
}

// SetSessionKey sets the Tickets session key
func (ticket *Ticket) SetSessionKey(sessionKey []byte) {
	ticket.sessionKey = sessionKey
}

// TargetPID returns the Tickets target PID
func (ticket *Ticket) TargetPID() uint32 {
	return ticket.targetPID
}

// SetTargetPID sets the Tickets target PID
func (ticket *Ticket) SetTargetPID(targetPID uint32) {
	ticket.targetPID = targetPID
}

// InternalData returns the Tickets internal data buffer
func (ticket *Ticket) InternalData() []byte {
	return ticket.internalData
}

// SetInternalData sets the Tickets internal data buffer
func (ticket *Ticket) SetInternalData(internalData []byte) {
	ticket.internalData = internalData
}

// Encrypt writes the ticket data to the provided stream and returns the encrypted byte slice
func (ticket *Ticket) Encrypt(key []byte, stream *StreamOut) []byte {
	encryption := NewKerberosEncryption(key)

	// Session key is not a NEX buffer type
	stream.Grow(int64(len(ticket.sessionKey)))
	stream.WriteBytesNext(ticket.sessionKey)

	stream.WriteUInt32LE(ticket.targetPID)
	stream.WriteBuffer(ticket.internalData)

	return encryption.Encrypt(stream.Bytes())
}

// NewKerberosTicket returns a new Ticket instance
func NewKerberosTicket() *Ticket {
	return &Ticket{}
}

// TicketInternalData contains information sent to the secure server
type TicketInternalData struct {
	timestamp  *DateTime
	userPID    uint32
	sessionKey []byte
}

// Timestamp returns the TicketInternalDatas timestamp
func (ticketInternalData *TicketInternalData) Timestamp() *DateTime {
	return ticketInternalData.timestamp
}

// SetTimestamp sets the TicketInternalDatas timestamp
func (ticketInternalData *TicketInternalData) SetTimestamp(timestamp *DateTime) {
	ticketInternalData.timestamp = timestamp
}

// UserPID returns the TicketInternalDatas user PID
func (ticketInternalData *TicketInternalData) UserPID() uint32 {
	return ticketInternalData.userPID
}

// SetUserPID sets the TicketInternalDatas user PID
func (ticketInternalData *TicketInternalData) SetUserPID(userPID uint32) {
	ticketInternalData.userPID = userPID
}

// SessionKey returns the TicketInternalDatas session key
func (ticketInternalData *TicketInternalData) SessionKey() []byte {
	return ticketInternalData.sessionKey
}

// SetSessionKey sets the TicketInternalDatas session key
func (ticketInternalData *TicketInternalData) SetSessionKey(sessionKey []byte) {
	ticketInternalData.sessionKey = sessionKey
}

// Encrypt writes the ticket data to the provided stream and returns the encrypted byte slice
func (ticketInternalData *TicketInternalData) Encrypt(key []byte, stream *StreamOut) []byte {
	stream.WriteUInt64LE(ticketInternalData.timestamp.Value())
	stream.WriteUInt32LE(ticketInternalData.userPID)

	// Session key is not a NEX buffer type
	stream.Grow(int64(len(ticketInternalData.sessionKey)))
	stream.WriteBytesNext(ticketInternalData.sessionKey)

	data := stream.Bytes()

	if stream.Server.KerberosTicketVersion() == 1 {
		ticketKey := make([]byte, 16)
		rand.Read(ticketKey)

		finalKey := MD5Hash(append(key, ticketKey...))

		encryption := NewKerberosEncryption(finalKey)

		encrypted := encryption.Encrypt(data)

		finalStream := NewStreamOut(stream.Server)

		finalStream.WriteBuffer(ticketKey)
		finalStream.WriteBuffer(encrypted)

		return finalStream.Bytes()
	} else {
		encryption := NewKerberosEncryption([]byte(key))
		return encryption.Encrypt(data)
	}
}

// Decrypt decrypts the given data and populates the struct
func (ticketInternalData *TicketInternalData) Decrypt(stream *StreamIn, key []byte) {
	if stream.Server.KerberosTicketVersion() == 1 {
		ticketKey, _ := stream.ReadBuffer()
		data, _ := stream.ReadBuffer()

		key = MD5Hash(append(key, ticketKey...))

		stream = NewStreamIn(data, stream.Server)
	}

	encryption := NewKerberosEncryption(key)
	decrypted := encryption.Decrypt(stream.Bytes())

	stream = NewStreamIn(decrypted, stream.Server)

	ticketInternalData.SetTimestamp(stream.ReadDateTime())
	ticketInternalData.SetUserPID(stream.ReadUInt32LE())
	ticketInternalData.SetSessionKey(stream.ReadBytesNext(int64(stream.Server.KerberosKeySize())))
}

// NewKerberosTicketInternalData returns a new TicketInternalData instance
func NewKerberosTicketInternalData() *TicketInternalData {
	return &TicketInternalData{}
}

// DeriveKerberosKey derives a users kerberos encryption key based on their PID and password
func DeriveKerberosKey(pid uint32, password []byte) []byte {
	for i := 0; i < 65000+int(pid)%1024; i++ {
		password = MD5Hash(password)
	}

	return password
}
