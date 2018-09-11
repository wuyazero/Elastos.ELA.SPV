package _interface

import (
	"github.com/wuyazero/Elastos.ELA.SPV/net"

	"github.com/wuyazero/Elastos.ELA.Utility/p2p"
)

type P2PClientImpl struct {
	magic uint32
	seeds []string
	pm    *net.PeerManager
}

func (client *P2PClientImpl) InitLocalPeer(initLocal func(peer *net.Peer)) {
	// Set Magic number of the P2P network
	p2p.Magic = client.magic
	// Create peer manager of the P2P network
	local := new(net.Peer)
	initLocal(local)
	client.pm = net.InitPeerManager(local, client.seeds)
}

func (client *P2PClientImpl) SetMessageHandler(msgHandler net.MessageHandler) {
	client.pm.SetMessageHandler(msgHandler)
}

func (client *P2PClientImpl) Start() {
	client.pm.Start()
}

func (client *P2PClientImpl) PeerManager() *net.PeerManager {
	return client.pm
}
