package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
)

// HPPPacket holds all the data about an HPP request
type HPPPacket struct {
	sender             *HPPClient
	accessKeySignature []byte
	passwordSignature  []byte
	payload            []byte
	message            *RMCMessage
	processed          chan bool
}

// Sender returns the Client who sent the packet
func (p *HPPPacket) Sender() ConnectionInterface {
	return p.sender
}

// Payload returns the packets payload
func (p *HPPPacket) Payload() []byte {
	return p.payload
}

// SetPayload sets the packets payload
func (p *HPPPacket) SetPayload(payload []byte) {
	p.payload = payload
}

func (p *HPPPacket) validateAccessKeySignature(signature string) error {
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("Failed to decode access key signature. %s", err)
	}

	p.accessKeySignature = signatureBytes

	calculatedSignature, err := p.calculateAccessKeySignature()
	if err != nil {
		return fmt.Errorf("Failed to calculate access key signature. %s", err)
	}

	if !bytes.Equal(calculatedSignature, p.accessKeySignature) {
		return errors.New("Access key signature does not match")
	}

	return nil
}

func (p *HPPPacket) calculateAccessKeySignature() ([]byte, error) {
	accessKey := p.Sender().Endpoint().AccessKey()

	accessKeyBytes, err := hex.DecodeString(accessKey)
	if err != nil {
		return nil, err
	}

	signature, err := p.calculateSignature(p.payload, accessKeyBytes)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (p *HPPPacket) validatePasswordSignature(signature string) error {
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("Failed to decode password signature. %s", err)
	}

	p.passwordSignature = signatureBytes

	calculatedSignature, err := p.calculatePasswordSignature()
	if err != nil {
		return fmt.Errorf("Failed to calculate password signature. %s", err)
	}

	if !bytes.Equal(calculatedSignature, p.passwordSignature) {
		return errors.New("Password signature does not match")
	}

	return nil
}

func (p *HPPPacket) calculatePasswordSignature() ([]byte, error) {
	sender := p.Sender()
	pid := sender.PID()
	account, _ := sender.Endpoint().(*HPPServer).AccountDetailsByPID(pid)
	if account == nil {
		return nil, errors.New("PID does not exist")
	}

	key := DeriveKerberosKey(pid, []byte(account.Password))

	signature, err := p.calculateSignature(p.payload, key)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (p *HPPPacket) calculateSignature(buffer []byte, key []byte) ([]byte, error) {
	mac := hmac.New(md5.New, key)

	_, err := mac.Write(buffer)
	if err != nil {
		return nil, err
	}

	hmac := mac.Sum(nil)

	return hmac, nil
}

// RMCMessage returns the packets RMC Message
func (p *HPPPacket) RMCMessage() *RMCMessage {
	return p.message
}

// SetRMCMessage sets the packets RMC Message
func (p *HPPPacket) SetRMCMessage(message *RMCMessage) {
	p.message = message
}

// NewHPPPacket creates and returns a new HPPPacket using the provided Client and payload
func NewHPPPacket(client *HPPClient, payload []byte) (*HPPPacket, error) {
	hppPacket := &HPPPacket{
		sender:    client,
		payload:   payload,
		processed: make(chan bool),
	}

	if payload != nil {
		rmcMessage := NewRMCRequest(client.Endpoint())
		err := rmcMessage.FromBytes(payload)
		if err != nil {
			return nil, fmt.Errorf("Failed to decode HPP request. %s", err)
		}

		hppPacket.SetRMCMessage(rmcMessage)
	}

	return hppPacket, nil
}
