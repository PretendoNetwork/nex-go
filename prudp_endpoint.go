package nex

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/PretendoNetwork/nex-go/types"
)

// PRUDPEndPoint is an implementation of rdv::PRUDPEndPoint.
// A PRUDPEndPoint represents a remote server location the client may connect to using a given remote stream ID.
// Each PRUDPEndPoint handles it's own set of PRUDPConnections, state, and events.
type PRUDPEndPoint struct {
	Server                       *PRUDPServer
	StreamID                     uint8
	DefaultstreamSettings        *StreamSettings
	Connections                  *MutexMap[string, *PRUDPConnection]
	packetEventHandlers          map[string][]func(packet PacketInterface)
	connectionEndedEventHandlers []func(connection *PRUDPConnection)
	ConnectionIDCounter          *Counter[uint32]
	ServerAccount                *Account
	AccountDetailsByPID          func(pid *types.PID) (*Account, *Error)
	AccountDetailsByUsername     func(username string) (*Account, *Error)
}

// RegisterServiceProtocol registers a NEX service with the endpoint
func (pep *PRUDPEndPoint) RegisterServiceProtocol(protocol ServiceProtocol) {
	pep.OnData(protocol.HandlePacket)
}

// OnData adds an event handler which is fired when a new DATA packet is received
func (pep *PRUDPEndPoint) OnData(handler func(packet PacketInterface)) {
	pep.on("data", handler)
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
			go handler(packet)
		}
	}
}

func (pep *PRUDPEndPoint) emitConnectionEnded(connection *PRUDPConnection) {
	for _, handler := range pep.connectionEndedEventHandlers {
		go handler(connection)
	}
}

func (pep *PRUDPEndPoint) processPacket(packet PRUDPPacketInterface, socket *SocketConnection) {
	streamType := packet.SourceVirtualPortStreamType()
	streamID := packet.SourceVirtualPortStreamID()
	discriminator := fmt.Sprintf("%s-%d-%d", socket.Address.String(), streamType, streamID)
	connection, ok := pep.Connections.Get(discriminator)

	if !ok {
		connection = NewPRUDPConnection(socket)
		connection.Endpoint = pep
		connection.ID = pep.ConnectionIDCounter.Next()
		connection.DefaultPRUDPVersion = packet.Version()
		connection.StreamType = streamType
		connection.StreamID = streamID
		connection.StreamSettings = pep.DefaultstreamSettings.Copy()
		connection.startHeartbeat()

		// * Fail-safe. If the server reboots, then
		// * connection has no record of old connections.
		// * An existing client which has not killed
		// * the connection on it's end MAY still send
		// * DATA packets once the server is back
		// * online, assuming it reboots fast enough.
		// * Since the client did NOT redo the SYN
		// * and CONNECT packets, it's reliable
		// * substreams never got remade. This is put
		// * in place to ensure there is always AT
		// * LEAST one substream in place, so the client
		// * can naturally error out due to the RC4
		// * errors.
		// *
		// * NOTE: THE CLIENT MAY NOT HAVE THE REAL
		// * CORRECT NUMBER OF SUBSTREAMS HERE. THIS
		// * IS ONLY DONE TO PREVENT A SERVER CRASH,
		// * NOT TO SAVE THE CLIENT. THE CLIENT IS
		// * EXPECTED TO NATURALLY DIE HERE
		connection.InitializeSlidingWindows(0)

		pep.Connections.Set(discriminator, connection)
	}

	packet.SetSender(connection)
	connection.resetHeartbeat()

	if packet.HasFlag(FlagAck) || packet.HasFlag(FlagMultiAck) {
		pep.handleAcknowledgment(packet)
		return
	}

	switch packet.Type() {
	case SynPacket:
		pep.handleSyn(packet)
	case ConnectPacket:
		pep.handleConnect(packet)
	case DataPacket:
		pep.handleData(packet)
	case DisconnectPacket:
		pep.handleDisconnect(packet)
	case PingPacket:
		pep.handlePing(packet)
	}
}

func (pep *PRUDPEndPoint) handleAcknowledgment(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagMultiAck) {
		pep.handleMultiAcknowledgment(packet)
		return
	}

	connection := packet.Sender().(*PRUDPConnection)

	slidingWindow := connection.SlidingWindow(packet.SubstreamID())
	slidingWindow.ResendScheduler.AcknowledgePacket(packet.SequenceID())
}

func (pep *PRUDPEndPoint) handleMultiAcknowledgment(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)
	stream := NewByteStreamIn(packet.Payload(), pep.Server)
	sequenceIDs := make([]uint16, 0)
	var baseSequenceID uint16
	var slidingWindow *SlidingWindow

	if packet.SubstreamID() == 1 {
		// * New aggregate acknowledgment packets set this to 1
		// * and encode the real substream ID in in the payload
		substreamID, _ := stream.ReadPrimitiveUInt8()
		additionalIDsCount, _ := stream.ReadPrimitiveUInt8()
		baseSequenceID, _ = stream.ReadPrimitiveUInt16LE()
		slidingWindow = connection.SlidingWindow(substreamID)

		for i := 0; i < int(additionalIDsCount); i++ {
			additionalID, _ := stream.ReadPrimitiveUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	} else {
		// TODO - This is how Kinnay's client handles this, but it doesn't make sense for QRV? Since it can have multiple reliable substreams?
		// * Old aggregate acknowledgment packets always use
		// * substream 0
		slidingWindow = connection.SlidingWindow(0)
		baseSequenceID = packet.SequenceID()

		for stream.Remaining() > 0 {
			additionalID, _ := stream.ReadPrimitiveUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	}

	// * MutexMap.Each locks the mutex, can't remove while reading.
	// * Have to just loop again
	slidingWindow.ResendScheduler.packets.Each(func(sequenceID uint16, pending *PendingPacket) bool {
		if sequenceID <= baseSequenceID && !slices.Contains(sequenceIDs, sequenceID) {
			sequenceIDs = append(sequenceIDs, sequenceID)
		}

		return false
	})

	// * Actually remove the packets from the pool
	for _, sequenceID := range sequenceIDs {
		slidingWindow.ResendScheduler.AcknowledgePacket(sequenceID)
	}
}

func (pep *PRUDPEndPoint) handleSyn(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(connection, nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(connection, nil)
	} else {
		ack, _ = NewPRUDPPacketV0(connection, nil)
	}

	connectionSignature, err := packet.calculateConnectionSignature(connection.Socket.Address)
	if err != nil {
		logger.Error(err.Error())
	}

	connection.reset()
	connection.Signature = connectionSignature

	ack.SetType(SynPacket)
	ack.AddFlag(FlagAck)
	ack.AddFlag(FlagHasSize)
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

	pep.emit("syn", ack)

	pep.Server.sendRaw(connection.Socket, ack.Bytes())
}

func (pep *PRUDPEndPoint) handleConnect(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)

	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(connection, nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(connection, nil)
	} else {
		ack, _ = NewPRUDPPacketV0(connection, nil)
	}

	connection.ServerConnectionSignature = packet.getConnectionSignature()
	connection.SessionID = packet.SessionID()

	connectionSignature, err := packet.calculateConnectionSignature(connection.Socket.Address)
	if err != nil {
		logger.Error(err.Error())
	}

	connection.ServerSessionID = packet.SessionID()

	ack.SetType(ConnectPacket)
	ack.AddFlag(FlagAck)
	ack.AddFlag(FlagHasSize)
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
		connection.outgoingUnreliableSequenceIDCounter = NewCounter[uint16](packet.(*PRUDPPacketV1).initialUnreliableSequenceID)
	} else {
		connection.InitializeSlidingWindows(0)
	}

	payload := make([]byte, 0)

	if len(packet.Payload()) != 0 {
		sessionKey, pid, checkValue, err := pep.readKerberosTicket(packet.Payload())
		if err != nil {
			logger.Error(err.Error())
		}

		connection.SetPID(pid)
		connection.setSessionKey(sessionKey)

		responseCheckValue := checkValue + 1
		responseCheckValueBytes := make([]byte, 4)

		binary.LittleEndian.PutUint32(responseCheckValueBytes, responseCheckValue)

		checkValueResponse := types.NewBuffer(responseCheckValueBytes)
		stream := NewByteStreamOut(pep.Server)

		checkValueResponse.WriteTo(stream)

		payload = stream.Bytes()
	}

	ack.SetPayload(payload)
	ack.setSignature(ack.calculateSignature([]byte{}, packet.getConnectionSignature()))

	pep.emit("connect", ack)

	pep.Server.sendRaw(connection.Socket, ack.Bytes())
}

func (pep *PRUDPEndPoint) handleData(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagReliable) {
		pep.handleReliable(packet)
	} else {
		pep.handleUnreliable(packet)
	}
}

func (pep *PRUDPEndPoint) handleDisconnect(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
		pep.acknowledgePacket(packet)
	}

	streamType := packet.SourceVirtualPortStreamType()
	streamID := packet.SourceVirtualPortStreamID()
	discriminator := fmt.Sprintf("%s-%d-%d", packet.Sender().Address().String(), streamType, streamID)
	if connection, ok := pep.Connections.Get(discriminator); ok {
		connection.cleanup()
		pep.Connections.Delete(discriminator)
	}

	pep.emit("disconnect", packet)
}

func (pep *PRUDPEndPoint) handlePing(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
		pep.acknowledgePacket(packet)
	}
}

func (pep *PRUDPEndPoint) readKerberosTicket(payload []byte) ([]byte, *types.PID, uint32, error) {
	stream := NewByteStreamIn(payload, pep.Server)

	ticketData := types.NewBuffer(nil)
	if err := ticketData.ExtractFrom(stream); err != nil {
		return nil, nil, 0, err
	}

	requestData := types.NewBuffer(nil)
	if err := requestData.ExtractFrom(stream); err != nil {
		return nil, nil, 0, err
	}

	// * Sanity checks
	serverAccount, _ := pep.AccountDetailsByUsername(pep.ServerAccount.Username)
	if serverAccount == nil {
		return nil, nil, 0, errors.New("Failed to find endpoint server account")
	}

	if serverAccount.Password != pep.ServerAccount.Password {
		return nil, nil, 0, errors.New("Password for endpoint server account does not match the records from AccountDetailsByUsername")
	}

	serverKey := DeriveKerberosKey(serverAccount.PID, []byte(serverAccount.Password))

	ticket := NewKerberosTicketInternalData()
	if err := ticket.Decrypt(NewByteStreamIn(ticketData.Value, pep.Server), serverKey); err != nil {
		return nil, nil, 0, err
	}

	ticketTime := ticket.Issued.Standard()
	serverTime := time.Now().UTC()

	timeLimit := ticketTime.Add(time.Minute * 2)
	if serverTime.After(timeLimit) {
		return nil, nil, 0, errors.New("Kerberos ticket expired")
	}

	sessionKey := ticket.SessionKey
	kerberos := NewKerberosEncryption(sessionKey)

	decryptedRequestData, err := kerberos.Decrypt(requestData.Value)
	if err != nil {
		return nil, nil, 0, err
	}

	checkDataStream := NewByteStreamIn(decryptedRequestData, pep.Server)

	userPID := types.NewPID(0)
	if err := userPID.ExtractFrom(checkDataStream); err != nil {
		return nil, nil, 0, err
	}

	_, err = checkDataStream.ReadPrimitiveUInt32LE() // * CID of secure server station url
	if err != nil {
		return nil, nil, 0, err
	}

	responseCheck, err := checkDataStream.ReadPrimitiveUInt32LE()
	if err != nil {
		return nil, nil, 0, err
	}

	return sessionKey, userPID, responseCheck, nil
}

func (pep *PRUDPEndPoint) acknowledgePacket(packet PRUDPPacketInterface) {
	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(packet.Sender().(*PRUDPConnection), nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(packet.Sender().(*PRUDPConnection), nil)
	} else {
		ack, _ = NewPRUDPPacketV0(packet.Sender().(*PRUDPConnection), nil)
	}

	ack.SetType(packet.Type())
	ack.AddFlag(FlagAck)
	ack.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	ack.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	ack.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	ack.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	ack.SetSequenceID(packet.SequenceID())
	ack.setFragmentID(packet.getFragmentID())
	ack.SetSubstreamID(packet.SubstreamID())

	pep.Server.sendPacket(ack)

	// * Servers send the DISCONNECT ACK 3 times
	if packet.Type() == DisconnectPacket {
		pep.Server.sendPacket(ack)
		pep.Server.sendPacket(ack)
	}
}

func (pep *PRUDPEndPoint) handleReliable(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
		pep.acknowledgePacket(packet)
	}

	connection := packet.Sender().(*PRUDPConnection)

	slidingWindow := packet.Sender().(*PRUDPConnection).SlidingWindow(packet.SubstreamID())

	for _, pendingPacket := range slidingWindow.Update(packet) {
		if packet.Type() == DataPacket {
			var decryptedPayload []byte

			if packet.Version() != 2 {
				decryptedPayload = pendingPacket.decryptPayload()
			} else {
				// * PRUDPLite does not encrypt payloads
				decryptedPayload = pendingPacket.Payload()
			}

			decompressedPayload, err := connection.StreamSettings.CompressionAlgorithm.Decompress(decryptedPayload)
			if err != nil {
				logger.Error(err.Error())
			}

			payload := slidingWindow.AddFragment(decompressedPayload)

			if packet.getFragmentID() == 0 {
				message := NewRMCMessage(pep.Server)
				err := message.FromBytes(payload)
				if err != nil {
					// TODO - Should this return the error too?
					logger.Error(err.Error())
				}

				slidingWindow.ResetFragmentedPayload()

				packet.SetRMCMessage(message)

				pep.emit("data", packet)
			}
		}
	}
}

func (pep *PRUDPEndPoint) handleUnreliable(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
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

	message := NewRMCMessage(pep.Server)
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
		ping, _ = NewPRUDPPacketV0(connection, nil)
	case 1:
		ping, _ = NewPRUDPPacketV1(connection, nil)
	case 2:
		ping, _ = NewPRUDPPacketLite(connection, nil)
	}

	ping.SetType(PingPacket)
	ping.AddFlag(FlagNeedsAck)
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
		if pc.pid.Value() == pid {
			connection = pc
			return true
		}

		return false
	})

	return connection
}

// NewPRUDPEndPoint returns a new PRUDPEndPoint for a server on the provided stream ID
func NewPRUDPEndPoint(streamID uint8) *PRUDPEndPoint {
	return &PRUDPEndPoint{
		StreamID:                     streamID,
		DefaultstreamSettings:        NewStreamSettings(),
		Connections:                  NewMutexMap[string, *PRUDPConnection](),
		packetEventHandlers:          make(map[string][]func(PacketInterface)),
		connectionEndedEventHandlers: make([]func(connection *PRUDPConnection), 0),
		ConnectionIDCounter:          NewCounter[uint32](0),
	}
}
