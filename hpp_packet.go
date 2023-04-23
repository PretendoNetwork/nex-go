package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"errors"
)

// HPPPacket represents an HPP packet
type HPPPacket struct {
	Packet
	accessKeySignature []byte
	passwordSignature  []byte
}

// SetAccessKeySignature sets the packet access key signature
func (packet *HPPPacket) SetAccessKeySignature(accessKeySignature string) {
	accessKeySignatureBytes, err := hex.DecodeString(accessKeySignature)
	if err != nil {
		logger.Error("[HPP] Failed to convert AccessKeySignature to bytes")
	}

	packet.accessKeySignature = accessKeySignatureBytes
}

// AccessKeySignature returns the packet access key signature
func (packet *HPPPacket) AccessKeySignature() []byte {
	return packet.accessKeySignature
}

// SetPasswordSignature sets the packet password signature
func (packet *HPPPacket) SetPasswordSignature(passwordSignature string) {
	passwordSignatureBytes, err := hex.DecodeString(passwordSignature)
	if err != nil {
		logger.Error("[HPP] Failed to convert PasswordSignature to bytes")
	}

	packet.passwordSignature = passwordSignatureBytes
}

// PasswordSignature returns the packet password signature
func (packet *HPPPacket) PasswordSignature() []byte {
	return packet.passwordSignature
}

// ValidateAccessKey checks if the access key signature is valid
func (packet *HPPPacket) ValidateAccessKey() error {
	accessKey := packet.Sender().Server().AccessKey()
	buffer := packet.rmcRequest.Bytes()

	accessKeyBytes, err := hex.DecodeString(accessKey)
	if err != nil {
		return err
	}

	calculatedAccessKeySignature := packet.calculateSignature(buffer, accessKeyBytes)
	if !bytes.Equal(calculatedAccessKeySignature, packet.accessKeySignature) {
		return errors.New("[HPP] Access key signature is not valid")
	}

	return nil
}

// ValidatePassword checks if the password signature is valid
func (packet *HPPPacket) ValidatePassword() error {
	pid := packet.Sender().PID()
	buffer := packet.rmcRequest.Bytes()

	password := packet.sender.server.passwordFromPIDHandler(pid)
	if password == "" {
		return errors.New("[HPP] PID does not exist")
	}

	passwordBytes := []byte(password)

	passwordSignatureKey := DeriveKerberosKey(pid, passwordBytes)

	calculatedPasswordSignature := packet.calculateSignature(buffer, passwordSignatureKey)
	if !bytes.Equal(calculatedPasswordSignature, packet.passwordSignature) {
		return errors.New("[HPP] Password signature is invalid")
	}

	return nil
}

func (packet *HPPPacket) calculateSignature(buffer []byte, key []byte) []byte {
	mac := hmac.New(md5.New, key)
	mac.Write(buffer)
	hmac := mac.Sum(nil)

	return hmac
}

// NewHPPPacket returns a new HPP packet
func NewHPPPacket(client *Client, data []byte) (*HPPPacket, error) {
	packet := NewPacket(client, data)

	hppPacket := HPPPacket{Packet: packet}

	if data != nil {
		hppPacket.payload = data

		rmcRequest := NewRMCRequest()
		err := rmcRequest.FromBytes(data)
		if err != nil {
			return &HPPPacket{}, errors.New("[HPP] Error parsing RMC request: " + err.Error())
		}

		hppPacket.rmcRequest = rmcRequest
	}

	return &hppPacket, nil
}
