package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/network"
	"github.com/filecoin-project/lotus/chain/sharding"
	adt0 "github.com/filecoin-project/specs-actors/actors/util/adt"
	"github.com/google/uuid"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/ipfs/go-merkledag"
	"github.com/ipld/go-car"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	lapi "github.com/filecoin-project/lotus/api"
	bstore "github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin"
	init_ "github.com/filecoin-project/lotus/chain/actors/builtin/init"
	"github.com/filecoin-project/lotus/chain/actors/builtin/reward"
	"github.com/filecoin-project/lotus/chain/actors/builtin/system"
	"github.com/filecoin-project/lotus/chain/consensus"
	"github.com/filecoin-project/lotus/chain/consensus/delegcns"
	"github.com/filecoin-project/lotus/chain/gen"
	genesis2 "github.com/filecoin-project/lotus/chain/gen/genesis"
	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/lotus/chain/wallet"
	lcli "github.com/filecoin-project/lotus/cli"
	cliutil "github.com/filecoin-project/lotus/cli/util"
	"github.com/filecoin-project/lotus/cmd/lotus-sim/simulation/mock"
	"github.com/filecoin-project/lotus/genesis"
	"github.com/filecoin-project/lotus/journal"
	"github.com/filecoin-project/lotus/node"
)

var StartShardingSubKey = node.AddInvoke()

var delegatedCmd = &cli.Command{
	Name:  "delegated",
	Usage: "Delegated consensus testbed",
	Subcommands: []*cli.Command{
		delegatedGenesisCmd,
		delegatedMinerCmd,

		daemonCmd(node.Options(
			node.Override(new(consensus.Consensus), delegcns.NewDelegatedConsensus),
			node.Override(new(store.WeightFunc), delegcns.Weight),
			node.Override(new(stmgr.Executor), delegcns.TipSetExecutor()),
			node.Override(new(stmgr.UpgradeSchedule), delegcns.DefaultUpgradeSchedule()),

			node.Override(new(*sharding.ShardingSub), sharding.NewShardSub),
			node.Override(StartShardingSubKey, func(s *sharding.ShardingSub) {
				s.Start()
			}),
		)),
	},
}

var delegatedGenesisCmd = &cli.Command{
	Name:      "genesis",
	Usage:     "Generate genesis for delegated consensus",
	ArgsUsage: "[miner secpk addr] [outfile]",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return xerrors.Errorf("expected 2 arguments")
		}

		miner, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return xerrors.Errorf("parsing miner address: %w", err)
		}
		if miner.Protocol() != address.SECP256K1 {
			return xerrors.Errorf("must be secp address")
		}

		j := journal.NilJournal()
		bs := bstore.WrapIDStore(bstore.NewMemorySync())
		syscalls := vm.Syscalls(mock.Verifier)

		memks := wallet.NewMemKeyStore()
		w, err := wallet.NewWallet(memks)
		if err != nil {
			return err
		}

		vreg, err := w.WalletNew(cctx.Context, types.KTBLS)
		if err != nil {
			return err
		}
		rem, err := w.WalletNew(cctx.Context, types.KTBLS)
		if err != nil {
			return err
		}

		template := genesis.Template{
			NetworkVersion: network.Version13,
			Accounts: []genesis.Actor{{
				Type:    genesis.TAccount,
				Balance: types.FromFil(2),
				Meta:    json.RawMessage(`{"Owner":"` + miner.String() + `"}`), // correct??
			}},
			Miners:      nil,
			NetworkName: "eudico-" + uuid.New().String(),
			Timestamp:   uint64(time.Now().Unix()),

			VerifregRootKey: genesis.Actor{
				Type:    genesis.TAccount,
				Balance: types.FromFil(2),
				Meta:    json.RawMessage(`{"Owner":"` + vreg.String() + `"}`), // correct??
			},
			RemainderAccount: genesis.Actor{
				Type: genesis.TAccount,
				Meta: json.RawMessage(`{"Owner":"` + rem.String() + `"}`), // correct??
			},
		}

		b, err := MakeDelegatedGenesisBlock(context.TODO(), j, bs, syscalls, template)
		if err != nil {
			return xerrors.Errorf("make genesis block: %w", err)
		}

		fmt.Printf("GENESIS MINER ADDRESS: t0%d\n", genesis2.MinerStart)

		f, err := os.OpenFile(cctx.Args().Get(1), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		offl := offline.Exchange(bs)
		blkserv := blockservice.New(bs, offl)
		dserv := merkledag.NewDAGService(blkserv)

		if err := car.WriteCarWithWalker(context.TODO(), dserv, []cid.Cid{b.Genesis.Cid()}, f, gen.CarWalkFunc); err != nil {
			return xerrors.Errorf("write genesis car: %w", err)
		}

		log.Warnf("WRITING GENESIS FILE AT %s", f.Name())

		if err := f.Close(); err != nil {
			return err
		}

		return nil
	},
}

var delegatedMinerCmd = &cli.Command{
	Name:  "miner",
	Usage: "run delegated conesensus miner",
	Action: func(cctx *cli.Context) error {
		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := cliutil.ReqContext(cctx)

		head, err := api.ChainHead(ctx)
		if err != nil {
			return xerrors.Errorf("getting head: %w", err)
		}

		minerid, err := address.NewFromString("t0100")
		if err != nil {
			return err
		}
		miner, err := api.StateAccountKey(ctx, minerid, types.EmptyTSK)
		if err != nil {
			return err
		}

		log.Info("starting mining on @", head.Height())

		timer := time.NewTicker(time.Duration(build.BlockDelaySecs) * time.Second)
		for {
			select {
			case <-timer.C:
				base, err := api.ChainHead(ctx)
				if err != nil {
					log.Errorw("creating block failed", "error", err)
					continue
				}

				log.Info("try mining at @", base.Height())

				msgs, err := api.MpoolSelect(ctx, base.Key(), 1)
				if err != nil {
					log.Errorw("selecting messages failed", "error", err)
				}

				bh, err := api.MinerCreateBlock(context.TODO(), &lapi.BlockTemplate{
					Miner:            miner,
					Parents:          base.Key(),
					Ticket:           nil,
					Eproof:           nil,
					BeaconValues:     nil,
					Messages:         msgs,
					Epoch:            base.Height() + 1,
					Timestamp:        base.MinTimestamp() + build.BlockDelaySecs,
					WinningPoStProof: nil,
				})
				if err != nil {
					log.Errorw("creating block failed", "error", err)
					continue
				}

				err = api.SyncSubmitBlock(ctx, &types.BlockMsg{
					Header:        bh.Header,
					BlsMessages:   bh.BlsMessages,
					SecpkMessages: bh.SecpkMessages,
				})
				if err != nil {
					log.Errorw("submitting block failed", "error", err)
				}

				log.Info("mined a block! ", bh.Cid(), " msgs ", len(msgs))
			case <-ctx.Done():
				return nil
			}
		}
	},
}

func MakeDelegatedGenesisBlock(ctx context.Context, j journal.Journal, bs bstore.Blockstore, sys vm.SyscallBuilder, template genesis.Template) (*genesis2.GenesisBootstrap, error) {
	if j == nil {
		j = journal.NilJournal()
	}
	st, _, err := MakeInitialStateTree(ctx, bs, template)
	if err != nil {
		return nil, xerrors.Errorf("make initial state tree failed: %w", err)
	}

	stateroot, err := st.Flush(ctx)
	if err != nil {
		return nil, xerrors.Errorf("flush state tree failed: %w", err)
	}

	// temp chainstore
	//cs := store.NewChainStore(bs, bs, datastore.NewMapDatastore(), j)

	/*	// Verify PreSealed Data
		stateroot, err = VerifyPreSealedData(ctx, cs, sys, stateroot, template, keyIDs, template.NetworkVersion)
		if err != nil {
			return nil, xerrors.Errorf("failed to verify presealed data: %w", err)
		}

		stateroot, err = SetupStorageMiners(ctx, cs, sys, stateroot, template.Miners, template.NetworkVersion)
		if err != nil {
			return nil, xerrors.Errorf("setup miners failed: %w", err)
		}*/

	store := adt.WrapStore(ctx, cbor.NewCborStore(bs))
	emptyroot, err := adt0.MakeEmptyArray(store).Root()
	if err != nil {
		return nil, xerrors.Errorf("amt build failed: %w", err)
	}

	mm := &types.MsgMeta{
		BlsMessages:   emptyroot,
		SecpkMessages: emptyroot,
	}
	mmb, err := mm.ToStorageBlock()
	if err != nil {
		return nil, xerrors.Errorf("serializing msgmeta failed: %w", err)
	}
	if err := bs.Put(mmb); err != nil {
		return nil, xerrors.Errorf("putting msgmeta block to blockstore: %w", err)
	}

	log.Infof("Empty Genesis root: %s", emptyroot)

	tickBuf := make([]byte, 32)
	_, _ = rand.Read(tickBuf)
	genesisticket := &types.Ticket{
		VRFProof: tickBuf,
	}

	b := &types.BlockHeader{
		Miner:                 system.Address,
		Ticket:                genesisticket,
		Parents:               []cid.Cid{},
		Height:                0,
		ParentWeight:          types.NewInt(0),
		ParentStateRoot:       stateroot,
		Messages:              mmb.Cid(),
		ParentMessageReceipts: emptyroot,
		BLSAggregate:          nil,
		BlockSig:              nil,
		Timestamp:             template.Timestamp,
		ElectionProof:         new(types.ElectionProof),
		BeaconEntries: []types.BeaconEntry{
			{
				Round: 0,
				Data:  make([]byte, 32),
			},
		},
		ParentBaseFee: abi.NewTokenAmount(build.InitialBaseFee),
	}

	sb, err := b.ToStorageBlock()
	if err != nil {
		return nil, xerrors.Errorf("serializing block header failed: %w", err)
	}

	if err := bs.Put(sb); err != nil {
		return nil, xerrors.Errorf("putting header to blockstore: %w", err)
	}

	return &genesis2.GenesisBootstrap{
		Genesis: b,
	}, nil
}

func SetupShardActor(ctx context.Context, bs bstore.Blockstore) (*types.Actor, error) {
	cst := cbor.NewCborStore(bs)
	st := delegcns.ConstructShardState()

	statecid, err := cst.Put(ctx, st)
	if err != nil {
		return nil, err
	}

	act := &types.Actor{
		Code:    delegcns.ShardActorCodeID,
		Balance: big.Zero(),
		Head:    statecid,
	}

	return act, nil
}

func MakeInitialStateTree(ctx context.Context, bs bstore.Blockstore, template genesis.Template) (*state.StateTree, map[address.Address]address.Address, error) {
	// Create empty state tree

	cst := cbor.NewCborStore(bs)
	_, err := cst.Put(context.TODO(), []struct{}{})
	if err != nil {
		return nil, nil, xerrors.Errorf("putting empty object: %w", err)
	}

	sv, err := state.VersionForNetwork(template.NetworkVersion)
	if err != nil {
		return nil, nil, xerrors.Errorf("getting state tree version: %w", err)
	}

	state, err := state.NewStateTree(cst, sv)
	if err != nil {
		return nil, nil, xerrors.Errorf("making new state tree: %w", err)
	}

	av, err := actors.VersionForNetwork(template.NetworkVersion)
	if err != nil {
		return nil, nil, xerrors.Errorf("getting network version: %w", err)
	}

	// Create system actor

	sysact, err := genesis2.SetupSystemActor(ctx, bs, av)
	if err != nil {
		return nil, nil, xerrors.Errorf("setup system actor: %w", err)
	}
	if err := state.SetActor(system.Address, sysact); err != nil {
		return nil, nil, xerrors.Errorf("set system actor: %w", err)
	}

	// Create init actor

	idStart, initact, keyIDs, err := genesis2.SetupInitActor(ctx, bs, template.NetworkName, template.Accounts, template.VerifregRootKey, template.RemainderAccount, av)
	if err != nil {
		return nil, nil, xerrors.Errorf("setup init actor: %w", err)
	}
	if err := state.SetActor(init_.Address, initact); err != nil {
		return nil, nil, xerrors.Errorf("set init actor: %w", err)
	}

	// Setup shard actor

	shardact, err := SetupShardActor(ctx, bs)
	if err != nil {
		return nil, nil, err
	}

	err = state.SetActor(delegcns.ShardActorAddr, shardact)
	if err != nil {
		return nil, nil, xerrors.Errorf("set reward actor: %w", err)
	}

	// Setup reward
	// RewardActor's state is overwritten by SetupStorageMiners, but needs to exist for miner creation messages
	rewact, err := genesis2.SetupRewardActor(ctx, bs, big.Zero(), av)
	if err != nil {
		return nil, nil, xerrors.Errorf("setup reward actor: %w", err)
	}

	err = state.SetActor(reward.Address, rewact)
	if err != nil {
		return nil, nil, xerrors.Errorf("set reward actor: %w", err)
	}

	bact, err := genesis2.MakeAccountActor(ctx, cst, av, builtin.BurntFundsActorAddr, big.Zero())
	if err != nil {
		return nil, nil, xerrors.Errorf("setup burnt funds actor state: %w", err)
	}
	if err := state.SetActor(builtin.BurntFundsActorAddr, bact); err != nil {
		return nil, nil, xerrors.Errorf("set burnt funds actor: %w", err)
	}

	// Create accounts
	for _, info := range template.Accounts {

		switch info.Type {
		case genesis.TAccount:
			if err := genesis2.CreateAccountActor(ctx, cst, state, info, keyIDs, av); err != nil {
				return nil, nil, xerrors.Errorf("failed to create account actor: %w", err)
			}

		case genesis.TMultisig:

			ida, err := address.NewIDAddress(uint64(idStart))
			if err != nil {
				return nil, nil, err
			}
			idStart++

			if err := genesis2.CreateMultisigAccount(ctx, cst, state, ida, info, keyIDs, av); err != nil {
				return nil, nil, err
			}
		default:
			return nil, nil, xerrors.New("unsupported account type")
		}

	}

	totalFilAllocated := big.Zero()

	err = state.ForEach(func(addr address.Address, act *types.Actor) error {
		if act.Balance.Nil() {
			panic(fmt.Sprintf("actor %s (%s) has nil balance", addr, builtin.ActorNameByCode(act.Code)))
		}
		totalFilAllocated = big.Add(totalFilAllocated, act.Balance)
		return nil
	})
	if err != nil {
		return nil, nil, xerrors.Errorf("summing account balances in state tree: %w", err)
	}

	totalFil := big.Mul(big.NewInt(int64(build.FilBase)), big.NewInt(int64(build.FilecoinPrecision)))
	remainingFil := big.Sub(totalFil, totalFilAllocated)
	if remainingFil.Sign() < 0 {
		return nil, nil, xerrors.Errorf("somehow overallocated filecoin (allocated = %s)", types.FIL(totalFilAllocated))
	}

	template.RemainderAccount.Balance = remainingFil

	switch template.RemainderAccount.Type {
	case genesis.TAccount:
		var ainfo genesis.AccountMeta
		if err := json.Unmarshal(template.RemainderAccount.Meta, &ainfo); err != nil {
			return nil, nil, xerrors.Errorf("unmarshaling account meta: %w", err)
		}

		_, ok := keyIDs[ainfo.Owner]
		if ok {
			return nil, nil, fmt.Errorf("remainder account has already been declared, cannot be assigned 90: %s", ainfo.Owner)
		}

		keyIDs[ainfo.Owner] = builtin.ReserveAddress
		err = genesis2.CreateAccountActor(ctx, cst, state, template.RemainderAccount, keyIDs, av)
		if err != nil {
			return nil, nil, xerrors.Errorf("creating remainder acct: %w", err)
		}

	case genesis.TMultisig:
		if err = genesis2.CreateMultisigAccount(ctx, cst, state, builtin.ReserveAddress, template.RemainderAccount, keyIDs, av); err != nil {
			return nil, nil, xerrors.Errorf("failed to set up remainder: %w", err)
		}
	default:
		return nil, nil, xerrors.Errorf("unknown account type for remainder: %w", err)
	}

	return state, keyIDs, nil
}
