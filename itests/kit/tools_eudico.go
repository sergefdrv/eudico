package kit

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	addr "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	napi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/filecoin-project/lotus/chain/consensus/hierarchical"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
)

const (
	finalityTimeout  = 1200
	balanceSleepTime = 3
)

func getSubnetChainNotify(ctx context.Context, subnetAddr addr.SubnetID, api napi.FullNode) (heads <-chan []*napi.HeadChange, err error) {
	switch subnetAddr.String() {
	case "root":
		heads, err = api.ChainNotify(ctx)
	default:
		heads, err = api.SubnetChainNotify(ctx, subnetAddr)
	}
	return
}

func getSubnetChainHead(ctx context.Context, subnetAddr addr.SubnetID, api napi.FullNode) (head *types.TipSet, err error) {
	switch subnetAddr.String() {
	case "root":
		head, err = api.ChainHead(ctx)
	default:
		head, err = api.SubnetChainHead(ctx, subnetAddr)
	}
	return
}

func getSubnetStateActor(ctx context.Context, subnetAddr addr.SubnetID, addr addr.Address, api napi.FullNode) (a *types.Actor, err error) {
	switch subnetAddr.String() {
	case "root":
		a, err = api.StateGetActor(ctx, addr, types.EmptyTSK)
	default:
		a, err = api.SubnetStateGetActor(ctx, subnetAddr, addr, types.EmptyTSK)
	}
	return
}

func WaitSubnetActorBalance(ctx context.Context, subnetAddr addr.SubnetID, addr addr.Address, balance big.Int, api napi.FullNode) (int, error) {
	heads, err := getSubnetChainNotify(ctx, subnetAddr, api)
	if err != nil {
		return 0, err
	}

	n := 0
	timer := time.After(finalityTimeout * time.Second)

	for {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("context closed")
		case <-heads:
			a, err := getSubnetStateActor(ctx, subnetAddr, addr, api)
			switch {
			case err != nil && !strings.Contains(err.Error(), types.ErrActorNotFound.Error()):
				return 0, err
			case err != nil && strings.Contains(err.Error(), types.ErrActorNotFound.Error()):
				n++
			case err == nil:
				if big.Cmp(balance, a.Balance) == 0 {
					return n, nil
				}
				n++
			}
		case <-timer:
			return 0, fmt.Errorf("finality timer exceeded")
		}
	}
}

func WaitForBalance(ctx context.Context, addr addr.Address, balance uint64, api napi.FullNode) error {
	currentBalance, err := api.WalletBalance(ctx, addr)
	if err != nil {
		return err
	}
	targetBalance := types.FromFil(balance)
	ticker := time.NewTicker(balanceSleepTime * time.Second)
	defer ticker.Stop()

	timer := time.After(finalityTimeout * time.Second)

	for big.Cmp(currentBalance, targetBalance) != 1 {
		select {
		case <-ctx.Done():
			return fmt.Errorf("closed channel")
		case <-ticker.C:
			currentBalance, err = api.WalletBalance(ctx, addr)
			if err != nil {
				return err
			}
		case <-timer:
			return fmt.Errorf("balance timer exceeded")
		}
	}

	return nil
}

// SubnetMinerMinesBlocks checks that the miner can mine some `m` blocks of `n` all blocks in the subnet.
func SubnetMinerMinesBlocks(ctx context.Context, m, n int, subnetAddr addr.SubnetID, miner addr.Address, api napi.FullNode) error {
	subnetHeads, err := getSubnetChainNotify(ctx, subnetAddr, api)
	if err != nil {
		return err
	}
	if n < 2 || n > 100 || m > n {
		return fmt.Errorf("wrong blocks number")
	}

	// ChainNotify returns channel with chain head updates.
	// First message is guaranteed to be of len == 1, and type == 'current'.
	// Without forks we can expect that its len always to be 1.
	initHead := <-subnetHeads
	if len(initHead) < 1 {
		return fmt.Errorf("empty chain head")
	}
	currHeight := initHead[0].Val.Height()

	i := 0
	minerMinedBlocks := 0
	for i < n {
		select {
		case <-ctx.Done():
			return fmt.Errorf("closed channel")
		case <-subnetHeads:
			if m >= minerMinedBlocks {
				return nil
			}

			head, err := getSubnetChainHead(ctx, subnetAddr, api)
			if err != nil {
				return err
			}

			height := head.Height()
			if height < currHeight {
				return fmt.Errorf("the current height is lower then the previous height")
			}
			if height == currHeight {
				continue
			}
			currHeight = height

			for _, b := range head.Blocks() {
				if b.Miner == miner {
					minerMinedBlocks++
				}
			}
		}

		i++
	}

	if m >= minerMinedBlocks {
		return nil
	}

	return fmt.Errorf("failed to mine %d blocks", m)

}

// SubnetHeightCheckForBlocks checks `n` blocks with correct heights in the subnet will be mined.
func SubnetHeightCheckForBlocks(ctx context.Context, n int, subnetAddr addr.SubnetID, api napi.FullNode) error {
	heads, err := getSubnetChainNotify(ctx, subnetAddr, api)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("closed channel")
	case <-heads:
	}

	currHead, err := getSubnetChainHead(ctx, subnetAddr, api)
	if err != nil {
		return err
	}

	i := 0
	for i < n {
		select {
		case <-ctx.Done():
			return fmt.Errorf("closed channel")
		case <-heads:
			newHead, err := getSubnetChainHead(ctx, subnetAddr, api)
			if err != nil {
				return err
			}

			if newHead.Height() <= currHead.Height() {
				return fmt.Errorf("wrong %d block height: prev block height - %d, current head height - %d",
					i, currHead.Height(), newHead.Height())
			}

			currHead = newHead
			i++
		}
	}

	return nil
}

// MessageForSend send the message
// TODO: use MessageForSend from cli package.
func MessageForSend(ctx context.Context, s api.FullNode, params lcli.SendParams) (*api.MessagePrototype, error) {
	if params.From == addr.Undef {
		defaddr, err := s.WalletDefaultAddress(ctx)
		if err != nil {
			return nil, err
		}
		params.From = defaddr
	}

	msg := types.Message{
		From:  params.From,
		To:    params.To,
		Value: params.Val,

		Method: params.Method,
		Params: params.Params,
	}

	if params.GasPremium != nil {
		msg.GasPremium = *params.GasPremium
	} else {
		msg.GasPremium = types.NewInt(0)
	}
	if params.GasFeeCap != nil {
		msg.GasFeeCap = *params.GasFeeCap
	} else {
		msg.GasFeeCap = types.NewInt(0)
	}
	if params.GasLimit != nil {
		msg.GasLimit = *params.GasLimit
	} else {
		msg.GasLimit = 0
	}
	validNonce := false
	if params.Nonce != nil {
		msg.Nonce = *params.Nonce
		validNonce = true
	}

	prototype := &api.MessagePrototype{
		Message:    msg,
		ValidNonce: validNonce,
	}
	return prototype, nil
}

func GetFreeTCPLocalAddr() (addr string, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close() // nolint
			return fmt.Sprintf("127.0.0.1:%d", l.Addr().(*net.TCPAddr).Port), nil
		}
	}
	return
}

func GetFreeLibp2pLocalAddr() (m multiaddr.Multiaddr, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close() // nolint
			return multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", l.Addr().(*net.TCPAddr).Port))
		}
	}
	return
}

func GetLibp2pAddr(privKey []byte) (m multiaddr.Multiaddr, err error) {
	saddr, err := GetFreeLibp2pLocalAddr()
	if err != nil {
		return nil, err
	}

	priv, err := crypto.UnmarshalPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	peerID, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		panic(err)
	}

	peerInfo := peer.AddrInfo{
		ID:    peerID,
		Addrs: []multiaddr.Multiaddr{saddr},
	}

	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		return nil, err
	}

	return addrs[0], nil
}

func NodeLibp2pAddr(node *TestFullNode) (m multiaddr.Multiaddr, err error) {
	privKey, err := node.PrivKey(context.Background())
	if err != nil {
		return nil, err
	}

	a, err := GetLibp2pAddr(privKey)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func SpawnSideSubnet(
	ctx context.Context,
	cancel context.CancelFunc,
	t *testing.T,
	wg *sync.WaitGroup,
	miner addr.Address,
	parent addr.SubnetID,
	name string,
	stake abi.TokenAmount,
	alg hierarchical.ConsensusType,
	api v1api.FullNode,
) {
	defer func() {
		wg.Done()
		t.Logf("[*] miner in subnet %s stopped", name)
	}()
	subnetParams := &hierarchical.SubnetParams{
		Addr:   miner,
		Parent: parent,
		Name:   name,
		Stake:  stake,
		Consensus: hierarchical.ConsensusParams{
			Alg:           alg,
			MinValidators: 0,
			DelegMiner:    miner,
		},
	}
	actorAddr, err := api.AddSubnet(ctx, subnetParams)
	require.NoError(t, err)

	subnetAddr := addr.NewSubnetID(parent, actorAddr)

	networkName, err := api.StateNetworkName(ctx)
	require.NoError(t, err)
	require.Equal(t, dtypes.NetworkName("/root"), networkName)

	t.Logf("[*] subnet %s addr: %v", name, subnetAddr)

	val, err := types.ParseFIL("10")
	require.NoError(t, err)

	_, err = api.StateLookupID(ctx, miner, types.EmptyTSK)
	require.NoError(t, err)

	sc, err := api.JoinSubnet(ctx, miner, big.Int(val), subnetAddr, "")
	require.NoError(t, err)

	_, err = api.StateWaitMsg(ctx, sc, 1, 100, false)
	require.NoError(t, err)

	t.Logf("[*] miner in subnet %s starting", name)
	smp := hierarchical.MiningParams{}
	err = api.MineSubnet(ctx, miner, subnetAddr, false, &smp)
	if err != nil {
		t.Error("subnet 2 miner error:", err)
		cancel()
		return
	}

	err = SubnetHeightCheckForBlocks(ctx, 3, subnetAddr, api)
	if err != nil {
		t.Errorf("subnet %s mining error: %v", name, err)
		cancel()
		return
	}

	err = api.MineSubnet(ctx, miner, subnetAddr, true, &smp)
	if err != nil {
		t.Error("subnet 4 miner error:", err)
		cancel()
		return
	}

}
