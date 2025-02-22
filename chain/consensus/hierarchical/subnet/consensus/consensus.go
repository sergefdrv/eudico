package consensus

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/filecoin-project/lotus/chain"
	"github.com/filecoin-project/lotus/chain/beacon"
	"github.com/filecoin-project/lotus/chain/consensus"
	"github.com/filecoin-project/lotus/chain/consensus/delegcns"
	"github.com/filecoin-project/lotus/chain/consensus/dummy"
	"github.com/filecoin-project/lotus/chain/consensus/hierarchical"
	"github.com/filecoin-project/lotus/chain/consensus/hierarchical/subnet"
	"github.com/filecoin-project/lotus/chain/consensus/hierarchical/subnet/resolver"
	"github.com/filecoin-project/lotus/chain/consensus/mir"
	"github.com/filecoin-project/lotus/chain/consensus/tspow"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
)

// Weight defines consensus weight function.
// TODO // FIXME: Make an SubnetConsensus interface from this functions
// to avoid having to use so many switch/cases. Deferring to the next
// refactor.
func Weight(consensus hierarchical.ConsensusType) (store.WeightFunc, error) {
	switch consensus {
	case hierarchical.Delegated:
		return delegcns.Weight, nil
	case hierarchical.PoW:
		return tspow.Weight, nil
	case hierarchical.Mir:
		return mir.Weight, nil
	case hierarchical.Dummy:
		return dummy.Weight, nil
	default:
		return nil, xerrors.New("consensus type not supported")
	}
}

func New(
	ctx context.Context,
	consensus hierarchical.ConsensusType,
	sm *stmgr.StateManager, snMgr subnet.Manager,
	beacon beacon.Schedule, r *resolver.Resolver,
	verifier ffiwrapper.Verifier,
	genesis chain.Genesis, netName dtypes.NetworkName,
) (consensus.Consensus, error) {
	switch consensus {
	case hierarchical.Delegated:
		return delegcns.NewDelegatedConsensus(ctx, sm, snMgr, beacon, r, verifier, genesis, netName), nil
	case hierarchical.PoW:
		return tspow.NewTSPoWConsensus(ctx, sm, snMgr, beacon, r, verifier, genesis, netName), nil
	case hierarchical.Mir:
		return mir.NewConsensus(ctx, sm, snMgr, beacon, r, verifier, genesis, netName)
	case hierarchical.Dummy:
		return dummy.NewConsensus(ctx, sm, snMgr, beacon, r, verifier, genesis, netName)
	default:
		return nil, xerrors.New("consensus type not supported")
	}
}

func Mine(ctx context.Context, api v1api.FullNode, wallet address.Address, cnsType hierarchical.ConsensusType) (err error) {
	switch cnsType {
	case hierarchical.Delegated:
		err = delegcns.Mine(ctx, wallet, api)
	case hierarchical.PoW:
		err = tspow.Mine(ctx, wallet, api)
	case hierarchical.Mir:
		err = mir.Mine(ctx, wallet, api)
	case hierarchical.Dummy:
		err = dummy.Mine(ctx, wallet, api)
	default:
		err = xerrors.New("consensus type not supported")
	}
	return
}
