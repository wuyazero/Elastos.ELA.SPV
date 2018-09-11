package sdk

import (
	"errors"
	"fmt"
	"strings"

	"github.com/wuyazero/Elastos.ELA.SPV/net"

	"github.com/wuyazero/Elastos.ELA.Utility/p2p"
	"github.com/wuyazero/Elastos.ELA.Utility/p2p/msg"
)

type P2PClientImpl struct {
	msgHandler  P2PMessageHandler
	peerManager *net.PeerManager
}

func NewP2PClientImpl(magic uint32, clientId uint64, seeds []string) (*P2PClientImpl, error) {
	// Initialize local peer
	local := new(net.Peer)
	local.SetID(clientId)
	local.SetVersion(ProtocolVersion)
	local.SetPort(SPVClientPort)

	if magic == 0 {
		return nil, errors.New("Magic number has not been set ")
	}
	// Set Magic number of the P2P network
	p2p.Magic = magic

	if len(seeds) == 0 {
		return nil, errors.New("Seeds list is empty ")
	}

	// Create client instance
	client := new(P2PClientImpl)

	// Initialize peer manager
	client.peerManager = net.InitPeerManager(local, toSPVAddr(seeds))

	// Set message handler
	client.peerManager.SetMessageHandler(client)

	return client, nil
}

func (client *P2PClientImpl) SetMessageHandler(handler P2PMessageHandler) {
	client.msgHandler = handler
}

func (client *P2PClientImpl) Start() {
	// Start
	client.peerManager.Start()
}

// Convert seed addresses to SPVServerPort according to the SPV protocol
func toSPVAddr(seeds []string) []string {
	var addrs = make([]string, len(seeds))
	for i, seed := range seeds {
		portIndex := strings.LastIndex(seed, ":")
		if portIndex > 0 {
			addrs[i] = fmt.Sprint(string([]byte(seed)[:portIndex]), ":", SPVServerPort)
		} else {
			addrs[i] = fmt.Sprint(seed, ":", SPVServerPort)
		}
	}
	return addrs
}

// Filter peer handshake according to the SPV protocol
func (client *P2PClientImpl) OnHandshake(v *msg.Version) error {
	if v.Version < ProtocolVersion {
		return errors.New(fmt.Sprint("To support SPV protocol, peer version must greater than ", ProtocolVersion))
	}

	if v.Services/ServiveSPV&1 == 0 {
		return errors.New("SPV service not enabled on connected peer")
	}

	return nil
}

func (client *P2PClientImpl) MakeMessage(cmd string) (p2p.Message, error) {
	return client.msgHandler.MakeMessage(cmd)
}

func (client *P2PClientImpl) OnPeerEstablish(peer *net.Peer) {
	client.msgHandler.OnPeerEstablish(peer)
}

func (client *P2PClientImpl) HandleMessage(peer *net.Peer, msg p2p.Message) error {
	return client.msgHandler.HandleMessage(peer, msg)
}

func (client *P2PClientImpl) PeerManager() *net.PeerManager {
	return client.peerManager
}
