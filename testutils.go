package bitswap

import (
	"context"
	"time"

	tn "github.com/ipfs/go-bitswap/testnet"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	delay "gx/ipfs/QmRJVNatYJwTAHgdSM1Xef9QVQ1Ch3XHdmcrykjP5Y4soL/go-ipfs-delay"
	testutil "gx/ipfs/QmRNhSdqzMcuRxX9A1egBeQ3BhDTguDV5HPwi8wRykkPU8/go-testutil"
	ds "gx/ipfs/QmSpg1CvpXQQow5ernt1gNBXaXV6yxyNqi7XoeerWfzB5w/go-datastore"
	delayed "gx/ipfs/QmSpg1CvpXQQow5ernt1gNBXaXV6yxyNqi7XoeerWfzB5w/go-datastore/delayed"
	ds_sync "gx/ipfs/QmSpg1CvpXQQow5ernt1gNBXaXV6yxyNqi7XoeerWfzB5w/go-datastore/sync"
	blockstore "gx/ipfs/QmegPGspn3RpTMQ23Fd3GVVMopo1zsEMurudbFMZ5UXBLH/go-ipfs-blockstore"
	p2ptestutil "gx/ipfs/QmfDapjsRAfzVpjeEm2tSmX19QpCrkLDXRCDDWJcbbUsFn/go-libp2p-netutil"
)

// WARNING: this uses RandTestBogusIdentity DO NOT USE for NON TESTS!
func NewTestSessionGenerator(
	net tn.Network) SessionGenerator {
	ctx, cancel := context.WithCancel(context.Background())
	return SessionGenerator{
		net:    net,
		seq:    0,
		ctx:    ctx, // TODO take ctx as param to Next, Instances
		cancel: cancel,
	}
}

// TODO move this SessionGenerator to the core package and export it as the core generator
type SessionGenerator struct {
	seq    int
	net    tn.Network
	ctx    context.Context
	cancel context.CancelFunc
}

func (g *SessionGenerator) Close() error {
	g.cancel()
	return nil // for Closer interface
}

func (g *SessionGenerator) Next() Instance {
	g.seq++
	p, err := p2ptestutil.RandTestBogusIdentity()
	if err != nil {
		panic("FIXME") // TODO change signature
	}
	return MkSession(g.ctx, g.net, p)
}

func (g *SessionGenerator) Instances(n int) []Instance {
	var instances []Instance
	for j := 0; j < n; j++ {
		inst := g.Next()
		instances = append(instances, inst)
	}
	for i, inst := range instances {
		for j := i + 1; j < len(instances); j++ {
			oinst := instances[j]
			inst.Exchange.network.ConnectTo(context.Background(), oinst.Peer)
		}
	}
	return instances
}

type Instance struct {
	Peer       peer.ID
	Exchange   *Bitswap
	blockstore blockstore.Blockstore

	blockstoreDelay delay.D
}

func (i *Instance) Blockstore() blockstore.Blockstore {
	return i.blockstore
}

func (i *Instance) SetBlockstoreLatency(t time.Duration) time.Duration {
	return i.blockstoreDelay.Set(t)
}

// session creates a test bitswap session.
//
// NB: It's easy make mistakes by providing the same peer ID to two different
// sessions. To safeguard, use the SessionGenerator to generate sessions. It's
// just a much better idea.
func MkSession(ctx context.Context, net tn.Network, p testutil.Identity) Instance {
	bsdelay := delay.Fixed(0)

	adapter := net.Adapter(p)
	dstore := ds_sync.MutexWrap(delayed.New(ds.NewMapDatastore(), bsdelay))

	bstore, err := blockstore.CachedBlockstore(ctx,
		blockstore.NewBlockstore(ds_sync.MutexWrap(dstore)),
		blockstore.DefaultCacheOpts())
	if err != nil {
		panic(err.Error()) // FIXME perhaps change signature and return error.
	}

	bs := New(ctx, adapter, bstore).(*Bitswap)

	return Instance{
		Peer:            p.ID(),
		Exchange:        bs,
		blockstore:      bstore,
		blockstoreDelay: bsdelay,
	}
}
