package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var wg sync.WaitGroup

var authenticationServerAccount *nex.Account
var secureServerAccount *nex.Account
var testUserAccount *nex.Account

func accountDetailsByPID(pid *types.PID) (*nex.Account, *nex.Error) {
	if pid.Equals(authenticationServerAccount.PID) {
		return authenticationServerAccount, nil
	}

	if pid.Equals(secureServerAccount.PID) {
		return secureServerAccount, nil
	}

	if pid.Equals(testUserAccount.PID) {
		return testUserAccount, nil
	}

	return testUserAccount, nex.NewError(nex.ResultCodes.RendezVous.InvalidPID, "Invalid PID")
	//return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidPID, "Invalid PID")
}

func accountDetailsByUsername(username string) (*nex.Account, *nex.Error) {
	if username == authenticationServerAccount.Username {
		return authenticationServerAccount, nil
	}

	if username == secureServerAccount.Username {
		return secureServerAccount, nil
	}

	if username == testUserAccount.Username {
		return testUserAccount, nil
	}

	return testUserAccount, nex.NewError(nex.ResultCodes.RendezVous.InvalidPID, "Invalid username")
	//return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidPID, "Invalid username")
}

var mocking = false

func main() {
	authenticationServerAccount = nex.NewAccount(types.NewPID(1), "Quazal Authentication", "authpassword")
	secureServerAccount = nex.NewAccount(types.NewPID(2), "Quazal Rendez-Vous", "securepassword")
	testUserAccount = nex.NewAccount(types.NewPID(1800000000), "1800000000", "nexuserpassword")

	wg.Add(3)

	var packetSource *gopacket.PacketSource
	if handle, err := pcap.OpenOffline(os.Args[1]); err == nil {
		mocking = true
		packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
	}

	go startAuthenticationServer()
	go startSecureServer()
	go startHPPServer()
	if mocking {
		go mockNetworkTraffic(packetSource)
	}

	wg.Wait()
}

func mockSend(_ *nex.SocketConnection, _ []byte) error {
	// don't actually send
	return nil
}

func mockRecv(ip *layers.IPv4, udp *layers.UDP) {
	client := net.UDPAddr{
		IP:   ip.SrcIP,
		Port: int(udp.SrcPort),
		Zone: "",
	}

	if udp.DstPort == 60000 {
		err := authServer.HandleSocketMessage(udp.Payload, &client, nil)
		if err != nil {
			fmt.Println("[MOCK][AUTH]", err)
		}
	} else if udp.DstPort == 60001 {
		err := secureServer.HandleSocketMessage(udp.Payload, &client, nil)
		if err != nil {
			fmt.Println("[MOCK][AUTH]", err)
		}
	}
}

func mockNetworkTraffic(packets *gopacket.PacketSource) {
	time.Sleep(1 * time.Second) // Wait for server setup to finish

	secureServer.ListenMock(mockSend)
	authServer.ListenMock(mockSend)

	// Use first packet as reference timestamp
	packet, err := packets.NextPacket()
	if err != nil {
		fmt.Println("[MOCK]", err)
		return
	}

	ts := packet.Metadata().Timestamp
	startTime := time.Now()
	tsOffset := startTime.Sub(ts)
	fmt.Println("[Mock] Starting replay session from", ts, "-", tsOffset, "replay offset.")

	serverIp := net.ParseIP(os.Getenv("NEX_TEST_SERVER_IP"))

	packetCounter := 0
	for packet := range packets.Packets() {
		packetCounter++
		if packetCounter%1000 == 0 {
			fmt.Println("[Mock]", packetCounter, "packets replayed. Pcap time", packet.Metadata().Timestamp)
		}

		// Parse the packet, discard if not IPv4+UDP
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			continue // ?
		}
		ip := ipLayer.(*layers.IPv4)
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer == nil {
			continue
		}
		udp := udpLayer.(*layers.UDP)

		// Only mock packets addressed to the server
		if !ip.DstIP.Equal(serverIp) {
			continue
		}

		ts := packet.Metadata().Timestamp
		delay := time.Until(ts.Add(tsOffset))
		if delay < -10*time.Millisecond {
			fmt.Println("[Mock/WARN] Can't Keep Up! Did the system time change (breakpoints) or is the server overloaded? Running", -delay, "behind, adjusting replay offset.")
			tsOffset -= delay
		} else {
			time.Sleep(delay)
		}

		mockRecv(ip, udp)
	}

	fmt.Println("[Mock] Replay session finished in", time.Since(startTime), "with", packetCounter, "packets.")
}
