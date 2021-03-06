package bitswap

import (
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmRNhSdqzMcuRxX9A1egBeQ3BhDTguDV5HPwi8wRykkPU8/go-testutil"

	bsnet "github.com/ipfs/go-bitswap/network"
)

type Network interface {
	Adapter(testutil.Identity) bsnet.BitSwapNetwork

	HasPeer(peer.ID) bool
}
