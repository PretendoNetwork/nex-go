package nex

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
)

// PRUDPEndPoint is an implementation of rdv::PRUDPEndPoint.
// A PRUDPEndPoint represents a remote server location the client may connect to using a given remote stream ID.
// Each PRUDPEndPoint handles it's own set of PRUDPConnections, state, and events.
//
// In NEX there exists nn::nex::SecureEndPoint, which presumably is what differentiates between the authentication
// and secure servers. However the functionality of rdv::PRUDPEndPoint and nn::nex::SecureEndPoint is seemingly
// identical. Rather than duplicate the logic from PRUDPEndpoint, a IsSecureEndpoint flag has been added instead.
type PRUDPEndPoint struct {
	Server                            *PRUDPServer
	StreamID                          uint8
	DefaultStreamSettings             *StreamSettings
	Connections                       *MutexMap[string, *PRUDPConnection]
	packetHandlers                    map[uint16]func(packet PRUDPPacketInterface)
	packetEventHandlers               map[string][]func(packet PacketInterface)
	connectionEndedEventHandlers      []func(connection *PRUDPConnection)
	errorEventHandlers                []func(err *Error)
	ConnectionIDCounter               *Counter[uint32]
	ServerAccount                     *Account
	AccountDetailsByPID               func(pid types.PID) (*Account, *Error)
	AccountDetailsByUsername          func(username string) (*Account, *Error)
	IsSecureEndPoint                  bool
	CalcRetransmissionTimeoutCallback CalcRetransmissionTimeoutCallback
}

// CalcRetransmissionTimeoutCallback is an optional callback which can be used to override the RTO calculation
// for packets sent by this `PRUDPEndpoint`
type CalcRetransmissionTimeoutCallback func(rtt float64, sendCount uint32) time.Duration

// RegisterServiceProtocol registers a NEX service with the endpoint
func (pep *PRUDPEndPoint) RegisterServiceProtocol(protocol ServiceProtocol) {
	protocol.SetEndpoint(pep)
	pep.OnData(protocol.HandlePacket)
}

// RegisterCustomPacketHandler registers a custom handler for a given packet type. Used to override existing handlers or create new ones for custom packet types.
func (pep *PRUDPEndPoint) RegisterCustomPacketHandler(packetType uint16, handler func(packet PRUDPPacketInterface)) {
	pep.packetHandlers[packetType] = handler
}

// OnData adds an event handler which is fired when a new DATA packet is received
func (pep *PRUDPEndPoint) OnData(handler func(packet PacketInterface)) {
	pep.on("data", handler)
}

// OnError adds an event handler which is fired when an error occurs on the endpoint
func (pep *PRUDPEndPoint) OnError(handler func(err *Error)) {
	// * "Ended" events are a special case, so handle them separately
	pep.errorEventHandlers = append(pep.errorEventHandlers, handler)
}

// OnDisconnect adds an event handler which is fired when a new DISCONNECT packet is received
//
// To handle a connection being removed from the server, see OnConnectionEnded which fires on more cases
func (pep *PRUDPEndPoint) OnDisconnect(handler func(packet PacketInterface)) {
	pep.on("disconnect", handler)
}

// OnConnectionEnded adds an event handler which is fired when a connection is removed from the server
//
// Fires both on a natural disconnect and from a timeout
func (pep *PRUDPEndPoint) OnConnectionEnded(handler func(connection *PRUDPConnection)) {
	// * "Ended" events are a special case, so handle them separately
	pep.connectionEndedEventHandlers = append(pep.connectionEndedEventHandlers, handler)
}

func (pep *PRUDPEndPoint) on(name string, handler func(packet PacketInterface)) {
	if _, ok := pep.packetEventHandlers[name]; !ok {
		pep.packetEventHandlers[name] = make([]func(packet PacketInterface), 0)
	}

	pep.packetEventHandlers[name] = append(pep.packetEventHandlers[name], handler)
}

func (pep *PRUDPEndPoint) emit(name string, packet PRUDPPacketInterface) {
	if handlers, ok := pep.packetEventHandlers[name]; ok {
		for _, handler := range handlers {
			handler(packet)
		}
	}
}

func (pep *PRUDPEndPoint) emitConnectionEnded(connection *PRUDPConnection) {
	for _, handler := range pep.connectionEndedEventHandlers {
		handler(connection)
	}
}

// EmitError calls all the endpoints error event handlers with the provided error
func (pep *PRUDPEndPoint) EmitError(err *Error) {
	for _, handler := range pep.errorEventHandlers {
		handler(err)
	}
}

// cleanupConnection cleans up and deletes a connection from this endpoint. Will lock the Connections mutex - make sure
// you don't hold it during a call, or this will deadlock
func (pep *PRUDPEndPoint) cleanupConnection(connection *PRUDPConnection) {
	discriminator := fmt.Sprintf("%s-%d-%d", connection.Socket.Address.String(), connection.StreamType, connection.StreamID)

	found := false
	pep.Connections.RunAndDelete(discriminator, func(key string, conn *PRUDPConnection) {
		found = true
	})

	// * Probably this connection is on a different PRUDPEndPoint
	if !found {
		logger.Warningf("Tried to delete connection %v (ID %v) but it doesn't exist!", discriminator, connection.ID)
	}

	// * We can't do this during RunAndDelete, since we hold the Connections mutex then
	// * This way we avoid any recursive locking
	connection.cleanup()
}

func (pep *PRUDPEndPoint) processPacket(packet PRUDPPacketInterface, socket *SocketConnection) {
	streamType := packet.SourceVirtualPortStreamType()
	streamID := packet.SourceVirtualPortStreamID()
	discriminator := fmt.Sprintf("%s-%d-%d", socket.Address.String(), streamType, streamID)
	connection := pep.Connections.GetOrSetDefault(discriminator, func() *PRUDPConnection {
		connection := NewPRUDPConnection(socket)
		connection.endpoint = pep
		connection.ID = pep.ConnectionIDCounter.Next()
		connection.DefaultPRUDPVersion = packet.Version()
		connection.StreamType = streamType
		connection.StreamID = streamID
		connection.StreamSettings = pep.DefaultStreamSettings.Copy()
		return connection
	})

	connection.Lock()
	defer connection.Unlock()

	packet.SetSender(connection)

	if packet.HasFlag(constants.PacketFlagAck) || packet.HasFlag(constants.PacketFlagMultiAck) {
		pep.handleAcknowledgment(packet)
		return
	}

	if packetHandler, ok := pep.packetHandlers[packet.Type()]; ok {
		packetHandler(packet)
	} else {
		logger.Warningf("Unhandled packet type %d", packet.Type())
	}
}

func (pep *PRUDPEndPoint) handleAcknowledgment(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	if connection.ConnectionState < StateConnected {
		return
	}

	connection.resetHeartbeat()

	if packet.HasFlag(constants.PacketFlagMultiAck) {
		pep.handleMultiAcknowledgment(packet)
		return
	}

	if packet.Type() == constants.PingPacket {
		if packet.SequenceID() == connection.outgoingPingSequenceIDCounter.Value {
			connection.rtt.Adjust(time.Since(connection.lastSentPingTime))
		}
	} else {
		slidingWindow := connection.SlidingWindow(packet.SubstreamID())
		slidingWindow.TimeoutManager.AcknowledgePacket(packet.SequenceID())
	}
}

func (pep *PRUDPEndPoint) handleMultiAcknowledgment(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)
	stream := NewByteStreamIn(packet.Payload(), pep.Server.LibraryVersions, pep.ByteStreamSettings())
	sequenceIDs := make([]uint16, 0)
	var baseSequenceID uint16
	var slidingWindow *SlidingWindow

	if packet.SubstreamID() == 1 || packet.Version() == 2 {
		// * New aggregate acknowledgment packets set this to 1
		// * and encode the real substream ID in in the payload
		substreamID, _ := stream.ReadUInt8()
		additionalIDsCount, _ := stream.ReadUInt8()
		baseSequenceID, _ = stream.ReadUInt16LE()
		slidingWindow = connection.SlidingWindow(substreamID)

		for i := 0; i < int(additionalIDsCount); i++ {
			additionalID, _ := stream.ReadUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	} else {
		// TODO - This is how Kinnay's client handles this, but it doesn't make sense for QRV? Since it can have multiple reliable substreams?
		// * Old aggregate acknowledgment packets always use
		// * substream 0
		slidingWindow = connection.SlidingWindow(0)
		baseSequenceID = packet.SequenceID()

		for stream.Remaining() > 0 {
			additionalID, _ := stream.ReadUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	}

	// * MutexMap.Each locks the mutex, can't remove while reading.
	// * Have to just loop again
	slidingWindow.TimeoutManager.packets.Each(func(sequenceID uint16, pending PRUDPPacketInterface) bool {
		if sequenceID <= baseSequenceID && !slices.Contains(sequenceIDs, sequenceID) {
			sequenceIDs = append(sequenceIDs, sequenceID)
		}

		return false
	})

	// * Actually remove the packets from the pool
	for _, sequenceID := range sequenceIDs {
		slidingWindow.TimeoutManager.AcknowledgePacket(sequenceID)
	}
}

func (pep *PRUDPEndPoint) handleSyn(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)
	connection.resetHeartbeat()

	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(pep.Server, connection, nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(pep.Server, connection, nil)
	} else {
		ack, _ = NewPRUDPPacketV0(pep.Server, connection, nil)
	}

	connectionSignature, err := packet.calculateConnectionSignature(connection.Socket.Address)
	if err != nil {
		logger.Error(err.Error())
	}

	connection.reset()
	connection.Signature = connectionSignature

	ack.SetType(constants.SynPacket)
	ack.AddFlag(constants.PacketFlagAck)
	ack.AddFlag(constants.PacketFlagHasSize)
	ack.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	ack.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	ack.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	ack.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	ack.setConnectionSignature(connectionSignature)
	ack.setSignature(ack.calculateSignature([]byte{}, []byte{}))

	if ack, ok := ack.(*PRUDPPacketV1); ok {
		// * Negotiate with the client what we support
		ack.maximumSubstreamID = packet.(*PRUDPPacketV1).maximumSubstreamID // * No change needed, we can just support what the client wants
		ack.minorVersion = packet.(*PRUDPPacketV1).minorVersion             // * No change needed, we can just support what the client wants
		ack.supportedFunctions = pep.Server.SupportedFunctions & packet.(*PRUDPPacketV1).supportedFunctions
	}

	connection.ConnectionState = StateConnecting

	pep.emit("syn", ack)

	pep.Server.sendRaw(connection.Socket, ack.Bytes())
}

func (pep *PRUDPEndPoint) handleConnect(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	if connection.ConnectionState < StateConnecting {
		return
	}

	connection.resetHeartbeat()

	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(pep.Server, connection, nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(pep.Server, connection, nil)
	} else {
		ack, _ = NewPRUDPPacketV0(pep.Server, connection, nil)
	}

	connection.ServerConnectionSignature = packet.getConnectionSignature()
	connection.SessionID = packet.SessionID()

	connectionSignature, err := packet.calculateConnectionSignature(connection.Socket.Address)
	if err != nil {
		logger.Error(err.Error())
	}

	connection.ServerSessionID = packet.SessionID()

	ack.SetType(constants.ConnectPacket)
	ack.AddFlag(constants.PacketFlagAck)
	ack.AddFlag(constants.PacketFlagHasSize)
	ack.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	ack.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	ack.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	ack.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	ack.setConnectionSignature(make([]byte, len(connectionSignature)))
	ack.SetSessionID(connection.ServerSessionID)
	ack.SetSequenceID(1)

	if ack, ok := ack.(*PRUDPPacketV1); ok {
		// * At this stage the client and server have already
		// * negotiated what they each can support, so configure
		// * the client now and just send the client back the
		// * negotiated configuration
		ack.maximumSubstreamID = packet.(*PRUDPPacketV1).maximumSubstreamID
		ack.minorVersion = packet.(*PRUDPPacketV1).minorVersion
		ack.supportedFunctions = packet.(*PRUDPPacketV1).supportedFunctions

		connection.InitializeSlidingWindows(ack.maximumSubstreamID)
		connection.InitializePacketDispatchQueues(ack.maximumSubstreamID)
		connection.outgoingUnreliableSequenceIDCounter = NewCounter[uint16](packet.(*PRUDPPacketV1).initialUnreliableSequenceID)
	} else {
		connection.InitializeSlidingWindows(0)
		connection.InitializePacketDispatchQueues(0)
	}

	payload := make([]byte, 0)

	if pep.IsSecureEndPoint {
		var decryptedPayload []byte
		if pep.Server.PRUDPV0Settings.EncryptedConnect {
			decryptedPayload, err = connection.StreamSettings.EncryptionAlgorithm.Decrypt(packet.Payload())
			if err != nil {
				logger.Error(err.Error())
				return
			}

		} else {
			decryptedPayload = packet.Payload()
		}

		decompressedPayload, err := connection.StreamSettings.CompressionAlgorithm.Decompress(decryptedPayload)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		sessionKey, pid, checkValue, err := pep.readKerberosTicket(decompressedPayload)
		if err != nil {
			logger.Error(err.Error())
			return
		}

		connection.SetPID(pid)
		connection.setSessionKey(sessionKey)

		responseCheckValue := checkValue + 1
		responseCheckValueBytes := make([]byte, 4)

		binary.LittleEndian.PutUint32(responseCheckValueBytes, responseCheckValue)

		checkValueResponse := types.NewBuffer(responseCheckValueBytes)
		stream := NewByteStreamOut(pep.Server.LibraryVersions, pep.ByteStreamSettings())

		checkValueResponse.WriteTo(stream)

		payload = stream.Bytes()
	}

	compressedPayload, err := connection.StreamSettings.CompressionAlgorithm.Compress(payload)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	var encryptedPayload []byte
	if pep.Server.PRUDPV0Settings.EncryptedConnect {
		encryptedPayload, err = connection.StreamSettings.EncryptionAlgorithm.Encrypt(compressedPayload)
		if err != nil {
			logger.Error(err.Error())
			return
		}
	} else {
		encryptedPayload = compressedPayload
	}

	ack.SetPayload(encryptedPayload)
	ack.setSignature(ack.calculateSignature([]byte{}, packet.getConnectionSignature()))

	connection.ConnectionState = StateConnected
	connection.startHeartbeat()

	pep.emit("connect", ack)

	pep.Server.sendRaw(connection.Socket, ack.Bytes())
}

func (pep *PRUDPEndPoint) handleData(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	if connection.ConnectionState < StateConnected {
		return
	}

	connection.resetHeartbeat()

	if packet.HasFlag(constants.PacketFlagReliable) {
		pep.handleReliable(packet)
	} else {
		pep.handleUnreliable(packet)
	}
}

func (pep *PRUDPEndPoint) handleDisconnect(packet PRUDPPacketInterface) {
	// TODO - Should we check the state here, or just let the connection disconnect at any time?

	if packet.HasFlag(constants.PacketFlagNeedsAck) {
		pep.acknowledgePacket(packet)
	}

	streamType := packet.SourceVirtualPortStreamType()
	streamID := packet.SourceVirtualPortStreamID()
	discriminator := fmt.Sprintf("%s-%d-%d", packet.Sender().Address().String(), streamType, streamID)
	if connection, ok := pep.Connections.Get(discriminator); ok {
		pep.cleanupConnection(connection)
	}

	pep.emit("disconnect", packet)
}

func (pep *PRUDPEndPoint) handlePing(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	if connection.ConnectionState < StateConnected {
		return
	}

	connection.resetHeartbeat()

	if packet.HasFlag(constants.PacketFlagNeedsAck) {
		pep.acknowledgePacket(packet)
	}

	if packet.HasFlag(constants.PacketFlagReliable) {
		substreamID := packet.SubstreamID()
		packetDispatchQueue := connection.PacketDispatchQueue(substreamID)
		packetDispatchQueue.Queue(packet)
	}
}

func (pep *PRUDPEndPoint) readKerberosTicket(payload []byte) ([]byte, types.PID, uint32, error) {
	stream := NewByteStreamIn(payload, pep.Server.LibraryVersions, pep.ByteStreamSettings())

	ticketData := types.NewBuffer(nil)
	if err := ticketData.ExtractFrom(stream); err != nil {
		return nil, 0, 0, err
	}

	requestData := types.NewBuffer(nil)
	if err := requestData.ExtractFrom(stream); err != nil {
		return nil, 0, 0, err
	}

	// * Sanity checks
	serverAccount, _ := pep.AccountDetailsByUsername(pep.ServerAccount.Username)
	if serverAccount == nil {
		return nil, 0, 0, errors.New("Failed to find endpoint server account")
	}

	if serverAccount.Password != pep.ServerAccount.Password {
		return nil, 0, 0, errors.New("Password for endpoint server account does not match the records from AccountDetailsByUsername")
	}

	serverKey := DeriveKerberosKey(serverAccount.PID, []byte(serverAccount.Password))

	ticket := NewKerberosTicketInternalData(pep.Server)
	if err := ticket.Decrypt(NewByteStreamIn(ticketData, pep.Server.LibraryVersions, pep.ByteStreamSettings()), serverKey); err != nil {
		return nil, 0, 0, err
	}

	ticketTime := ticket.Issued.Standard()
	serverTime := time.Now().UTC()

	timeLimit := ticketTime.Add(time.Minute * 2)
	if serverTime.After(timeLimit) {
		return nil, 0, 0, errors.New("Kerberos ticket expired")
	}

	sessionKey := ticket.SessionKey
	kerberos := NewKerberosEncryption(sessionKey)

	decryptedRequestData, err := kerberos.Decrypt(requestData)
	if err != nil {
		return nil, 0, 0, err
	}

	checkDataStream := NewByteStreamIn(decryptedRequestData, pep.Server.LibraryVersions, pep.ByteStreamSettings())

	userPID := types.NewPID(0)
	if err := userPID.ExtractFrom(checkDataStream); err != nil {
		return nil, 0, 0, err
	}

	if userPID != ticket.SourcePID {
		return nil, 0, 0, errors.New("User PID and ticket source PID mismatch")
	}

	_, err = checkDataStream.ReadUInt32LE() // * CID of secure server station url
	if err != nil {
		return nil, 0, 0, err
	}

	responseCheck, err := checkDataStream.ReadUInt32LE()
	if err != nil {
		return nil, 0, 0, err
	}

	return sessionKey, userPID, responseCheck, nil
}

func (pep *PRUDPEndPoint) acknowledgePacket(packet PRUDPPacketInterface) {
	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(pep.Server, packet.Sender().(*PRUDPConnection), nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(pep.Server, packet.Sender().(*PRUDPConnection), nil)
	} else {
		ack, _ = NewPRUDPPacketV0(pep.Server, packet.Sender().(*PRUDPConnection), nil)
	}

	ack.SetType(packet.Type())
	ack.AddFlag(constants.PacketFlagAck)
	ack.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	ack.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	ack.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	ack.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	ack.SetSequenceID(packet.SequenceID())
	ack.setFragmentID(packet.getFragmentID())
	ack.SetSubstreamID(packet.SubstreamID())

	pep.Server.sendPacket(ack)

	// * Servers send the DISCONNECT ACK 3 times
	if packet.Type() == constants.DisconnectPacket {
		pep.Server.sendPacket(ack)
		pep.Server.sendPacket(ack)
	}
}

func (pep *PRUDPEndPoint) handleReliable(packet PRUDPPacketInterface) {
	if packet.HasFlag(constants.PacketFlagNeedsAck) {
		pep.acknowledgePacket(packet)
	}

	connection := packet.Sender().(*PRUDPConnection)

	substreamID := packet.SubstreamID()

	packetDispatchQueue := connection.PacketDispatchQueue(substreamID)
	packetDispatchQueue.Queue(packet)

	for nextPacket, ok := packetDispatchQueue.GetNextToDispatch(); ok; nextPacket, ok = packetDispatchQueue.GetNextToDispatch() {
		if nextPacket.Type() == constants.DataPacket {
			var decryptedPayload []byte

			if nextPacket.Version() != 2 {
				decryptedPayload = nextPacket.decryptPayload()
			} else {
				// * PRUDPLite does not encrypt payloads
				decryptedPayload = nextPacket.Payload()
			}

			decompressedPayload, err := connection.StreamSettings.CompressionAlgorithm.Decompress(decryptedPayload)
			if err != nil {
				logger.Error(err.Error())
			}

			incomingFragmentBuffer := connection.GetIncomingFragmentBuffer(substreamID)
			incomingFragmentBuffer = append(incomingFragmentBuffer, decompressedPayload...)
			connection.SetIncomingFragmentBuffer(substreamID, incomingFragmentBuffer)

			if nextPacket.getFragmentID() == 0 {
				message := NewRMCMessage(pep)
				err := message.FromBytes(incomingFragmentBuffer)
				if err != nil {
					// TODO - Should this return the error too?
					logger.Error(err.Error())
				}

				nextPacket.SetRMCMessage(message)
				connection.ClearOutgoingBuffer(substreamID)

				pep.emit("data", nextPacket)
			}
		}

		packetDispatchQueue.Dispatched(nextPacket)
	}
}

func (pep *PRUDPEndPoint) handleUnreliable(packet PRUDPPacketInterface) {
	if packet.HasFlag(constants.PacketFlagNeedsAck) {
		pep.acknowledgePacket(packet)
	}

	// * Since unreliable DATA packets can in theory reach the
	// * server in any order, and they lack a subsslidingWindowtream, it's
	// * not actually possible to know what order they should
	// * be processed in for each request. So assume all packets
	// * MUST be fragment 0 (unreliable packets do not have frags)
	// *
	// * Example -
	// *
	// * Say there is 2 requests to the same protocol, methods 1
	// * and 2. The starting unreliable sequence ID is 10. If both
	// * method 1 and 2 are called at the same time, but method 1
	// * has a fragmented payload, the packets could, in theory, reach
	// * the server like so:
	// *
	// *	- Method1 - Sequence 10, Fragment 1
	// *	- Method1 - Sequence 13, Fragment 3
	// *	- Method2 - Sequence 12, Fragment 0
	// *	- Method1 - Sequence 11, Fragment 2
	// *	- Method1 - Sequence 14, Fragment 0
	// *
	// * If we reorder these to the proper order, like so:
	// *
	// *	- Method1 - Sequence 10, Fragment 1
	// *	- Method1 - Sequence 11, Fragment 2
	// *	- Method2 - Sequence 12, Fragment 0
	// *	- Method1 - Sequence 13, Fragment 3
	// *	- Method1 - Sequence 14, Fragment 0
	// *
	// * We still have a gap where Method2 was called. It's not
	// * possible to know if the packet with sequence ID 12 belongs
	// * to the Method1 calls or not. We don't even know which methods
	// * the packets are for at this stage yet, since the RMC data
	// * can't be checked until all the fragments are collected and
	// * the payload decrypted. In this case, we would see fragment 0
	// * and assume that's the end of fragments, losing the real last
	// * fragments and resulting in a bad decryption
	// TODO - Is this actually true? I'm just assuming, based on common sense, tbh. Kinnay also does not implement fragmented unreliable packets?
	if packet.getFragmentID() != 0 {
		logger.Warningf("Unexpected unreliable fragment ID. Expected 0, got %d", packet.getFragmentID())
		return
	}

	payload := packet.processUnreliableCrypto()

	message := NewRMCMessage(pep)
	err := message.FromBytes(payload)
	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
	}

	packet.SetRMCMessage(message)

	pep.emit("data", packet)
}

func (pep *PRUDPEndPoint) sendPing(connection *PRUDPConnection) {
	var ping PRUDPPacketInterface

	switch connection.DefaultPRUDPVersion {
	case 0:
		ping, _ = NewPRUDPPacketV0(pep.Server, connection, nil)
	case 1:
		ping, _ = NewPRUDPPacketV1(pep.Server, connection, nil)
	case 2:
		ping, _ = NewPRUDPPacketLite(pep.Server, connection, nil)
	}

	ping.SetType(constants.PingPacket)
	ping.AddFlag(constants.PacketFlagNeedsAck)
	ping.SetSourceVirtualPortStreamType(connection.StreamType)
	ping.SetSourceVirtualPortStreamID(pep.StreamID)
	ping.SetDestinationVirtualPortStreamType(connection.StreamType)
	ping.SetDestinationVirtualPortStreamID(connection.StreamID)
	ping.SetSubstreamID(0)

	pep.Server.sendPacket(ping)
}

// FindConnectionByID returns the PRUDP client connected with the given connection ID
func (pep *PRUDPEndPoint) FindConnectionByID(connectedID uint32) *PRUDPConnection {
	var connection *PRUDPConnection

	pep.Connections.Each(func(discriminator string, pc *PRUDPConnection) bool {
		if pc.ID == connectedID {
			connection = pc
			return true
		}

		return false
	})

	return connection
}

// FindConnectionByPID returns the PRUDP client connected with the given PID
func (pep *PRUDPEndPoint) FindConnectionByPID(pid uint64) *PRUDPConnection {
	var connection *PRUDPConnection

	pep.Connections.Each(func(discriminator string, pc *PRUDPConnection) bool {
		if uint64(pc.pid) == pid && pc.ConnectionState == StateConnected {
			connection = pc
			return true
		}

		return false
	})

	return connection
}

// ComputeRetransmitTimeout computes the RTO (Retransmit timeout) for a given packet
func (pep *PRUDPEndPoint) ComputeRetransmitTimeout(packet PRUDPPacketInterface) time.Duration {
	connection := packet.Sender().(*PRUDPConnection)
	rtt := connection.rtt

	if callback := pep.CalcRetransmissionTimeoutCallback; callback != nil {
		rttAverage := rtt.GetRTTSmoothedAvg()
		rttDeviation := rtt.GetRTTSmoothedDev()
		return callback(rttAverage+rttDeviation*4.0, packet.SendCount())
	}

	var retransmitTimeBase int64
	if packet.Type() == constants.SynPacket {
		retransmitTimeBase = int64(pep.DefaultStreamSettings.SynInitialRTT)
	} else {
		retransmitTimeBase = int64(pep.DefaultStreamSettings.InitialRTT)
		if rtt.Initialized() {
			retransmitTimeBase = int64(rtt.Average()/time.Millisecond) / 8
		}
	}

	retransmitTimeBaseMultiplier := packet.SendCount()

	var retransmitMultiplier float64
	if packet.SendCount() < pep.DefaultStreamSettings.ExtraRetransmitTimeoutTrigger {
		retransmitMultiplier = float64(pep.DefaultStreamSettings.RetransmitTimeoutMultiplier)
	} else {
		retransmitMultiplier = float64(pep.DefaultStreamSettings.ExtraRetransmitTimeoutMultiplier)
	}

	return time.Duration(float64(retransmitTimeBase*int64(retransmitTimeBaseMultiplier))*retransmitMultiplier) * time.Millisecond
}

// AccessKey returns the servers sandbox access key
func (pep *PRUDPEndPoint) AccessKey() string {
	return pep.Server.AccessKey
}

// SetAccessKey sets the servers sandbox access key
func (pep *PRUDPEndPoint) SetAccessKey(accessKey string) {
	pep.Server.AccessKey = accessKey
}

// Send sends the packet to the packets sender
func (pep *PRUDPEndPoint) Send(packet PacketInterface) {
	pep.Server.Send(packet)
}

// LibraryVersions returns the versions that the server has
func (pep *PRUDPEndPoint) LibraryVersions() *LibraryVersions {
	return pep.Server.LibraryVersions
}

// ByteStreamSettings returns the settings to be used for ByteStreams
func (pep *PRUDPEndPoint) ByteStreamSettings() *ByteStreamSettings {
	return pep.Server.ByteStreamSettings
}

// SetByteStreamSettings sets the settings to be used for ByteStreams
func (pep *PRUDPEndPoint) SetByteStreamSettings(byteStreamSettings *ByteStreamSettings) {
	pep.Server.ByteStreamSettings = byteStreamSettings
}

// UseVerboseRMC checks whether or not the endpoint uses verbose RMC
func (pep *PRUDPEndPoint) UseVerboseRMC() bool {
	return pep.Server.UseVerboseRMC
}

// EnableVerboseRMC enable or disables the use of verbose RMC
func (pep *PRUDPEndPoint) EnableVerboseRMC(enable bool) {
	pep.Server.UseVerboseRMC = enable
}

// NewPRUDPEndPoint returns a new PRUDPEndPoint for a server on the provided stream ID
func NewPRUDPEndPoint(streamID uint8) *PRUDPEndPoint {
	pep := &PRUDPEndPoint{
		StreamID:                     streamID,
		DefaultStreamSettings:        NewStreamSettings(),
		Connections:                  NewMutexMap[string, *PRUDPConnection](),
		packetHandlers:               make(map[uint16]func(packet PRUDPPacketInterface)),
		packetEventHandlers:          make(map[string][]func(PacketInterface)),
		connectionEndedEventHandlers: make([]func(connection *PRUDPConnection), 0),
		errorEventHandlers:           make([]func(err *Error), 0),
		ConnectionIDCounter:          NewCounter[uint32](0),
		IsSecureEndPoint:             false,
	}

	pep.packetHandlers[constants.SynPacket] = pep.handleSyn
	pep.packetHandlers[constants.ConnectPacket] = pep.handleConnect
	pep.packetHandlers[constants.DataPacket] = pep.handleData
	pep.packetHandlers[constants.DisconnectPacket] = pep.handleDisconnect
	pep.packetHandlers[constants.PingPacket] = pep.handlePing

	return pep
}
