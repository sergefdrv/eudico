package tendermint

import (
	"context"
	"fmt"
	"strings"

	"github.com/Gurpartap/async"
	"github.com/hashicorp/go-multierror"
	logging "github.com/ipfs/go-log/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/minio/blake2b-simd"
	"github.com/tendermint/tendermint/libs/rand"
	tmclient "github.com/tendermint/tendermint/rpc/client/http"
	"go.opencensus.io/stats"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	bstore "github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain"
	"github.com/filecoin-project/lotus/chain/beacon"
	"github.com/filecoin-project/lotus/chain/consensus"
	"github.com/filecoin-project/lotus/chain/consensus/common"
	"github.com/filecoin-project/lotus/chain/consensus/hierarchical"
	"github.com/filecoin-project/lotus/chain/consensus/hierarchical/subnet"
	"github.com/filecoin-project/lotus/chain/consensus/hierarchical/subnet/resolver"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper"
	"github.com/filecoin-project/lotus/metrics"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
)

const (
	tendermintRPCAddressEnv     = "EUDICO_TENDERMINT_RPC"
	defaultTendermintRPCAddress = "http://127.0.0.1:26657"

	// Subnet tag is fixed length subnet ID added to messages sent to Tendermint
	tagLength = 8

	// MaxHeightDrift TODO: is that correct or should be adapted?
	MaxHeightDrift = 5
)

var (
	log                     = logging.Logger("tendermint-consensus")
	_   consensus.Consensus = &Tendermint{}
)

type Tendermint struct {
	store    *store.ChainStore
	beacon   beacon.Schedule
	sm       *stmgr.StateManager
	verifier ffiwrapper.Verifier
	genesis  *types.TipSet
	subMgr   subnet.SubnetMgr
	netName  address.SubnetID
	r        *resolver.Resolver

	// Used to access Tendermint RPC
	client *tmclient.HTTP
	// Offset in Tendermint blockchain
	offset int64
	// Subnet tag
	tag []byte
	// Tendermint validator secp256k1 address
	tendermintValidatorAddress string
	// Eudico client secp256k1 address
	eudicoClientAddress address.Address
	// Secp256k1 public key
	eudicoClientPubKey []byte
	// Mapping between Tendermint validator addresses and Eudico miner addresses
	tendermintEudicoAddresses map[string]address.Address
}

func NewConsensus(sm *stmgr.StateManager, submgr subnet.SubnetMgr, b beacon.Schedule, r *resolver.Resolver, v ffiwrapper.Verifier, g chain.Genesis, netName dtypes.NetworkName) consensus.Consensus {
	ctx := context.TODO()

	subnetID := address.SubnetID(netName)
	log.Infof("New Tendermint consensus for %s subnet", subnetID)
	tag := blake2b.Sum256([]byte(subnetID))

	c, err := tmclient.New(NodeAddr())
	if err != nil {
		log.Fatalf("unable to create a Tendermint RPC client: %s", err)
	}

	valAddr, valPubKey, clientAddr, err := getValidatorsInfo(ctx, c)
	if err != nil {
		log.Fatalf("unable to get or handle Tendermint validators info: %s", err)
	}
	log.Info("Tendermint validator addr:", valAddr)
	log.Info("Tendermint validator pub key:", valPubKey)
	log.Info("Eudico client addr: ", clientAddr)

	regMsg, err := NewRegistrationMessageBytes(subnetID, tag[:tagLength], rand.Bytes(16))
	if err != nil {
		log.Fatalf("unable to create a registration message: %s", err)
	}
	regSubnet, err := registerNetwork(ctx, c, regMsg)
	if err != nil {
		log.Fatalf("unable to registrate network: %s", err)
	}
	log.Info("subnet registered")
	log.Warnf("Tendermint offset for %s is %d", regSubnet.Name, regSubnet.Offset)

	return &Tendermint{
		store:                      sm.ChainStore(),
		beacon:                     b,
		sm:                         sm,
		verifier:                   v,
		genesis:                    g,
		subMgr:                     submgr,
		netName:                    subnetID,
		client:                     c,
		offset:                     regSubnet.Offset,
		tag:                        tag[:tagLength],
		tendermintValidatorAddress: valAddr,
		eudicoClientAddress:        clientAddr,
		eudicoClientPubKey:         valPubKey,
		tendermintEudicoAddresses:  make(map[string]address.Address),
	}
}

func (tm *Tendermint) ValidateBlock(ctx context.Context, b *types.FullBlock) (err error) {
	log.Infof("starting block validation process at @%d", b.Header.Height)

	if err := common.BlockSanityChecks(hierarchical.Tendermint, b.Header); err != nil {
		return xerrors.Errorf("incoming header failed basic sanity checks: %w", err)
	}

	h := b.Header

	baseTs, err := tm.store.LoadTipSet(types.NewTipSetKey(h.Parents...))
	if err != nil {
		return xerrors.Errorf("load parent tipset failed (%s): %w", h.Parents, err)
	}

	// fast checks first
	if h.Height != baseTs.Height()+1 {
		return xerrors.Errorf("block height not parent height+1: %d != %d", h.Height, baseTs.Height()+1)
	}

	now := uint64(build.Clock.Now().Unix())
	if h.Timestamp > now+build.AllowableClockDriftSecs {
		return xerrors.Errorf("block was from the future (now=%d, blk=%d): %w", now, h.Timestamp, consensus.ErrTemporal)
	}
	if h.Timestamp > now {
		log.Warn("Got block from the future, but within threshold", h.Timestamp, build.Clock.Now().Unix())
	}

	msgsChecks := common.CheckMsgsWithoutBlockSig(ctx, tm.store, tm.sm, tm.subMgr, tm.r, tm.netName, b, baseTs)

	minerCheck := async.Err(func() error {
		if err := tm.minerIsValid(b.Header.Miner); err != nil {
			return xerrors.Errorf("minerIsValid failed: %w", err)
		}
		return nil
	})

	pweight, err := Weight(context.TODO(), nil, baseTs)
	if err != nil {
		return xerrors.Errorf("getting parent weight: %w", err)
	}

	if types.BigCmp(pweight, b.Header.ParentWeight) != 0 {
		return xerrors.Errorf("parent weight different: %s (header) != %s (computed)",
			b.Header.ParentWeight, pweight)
	}

	stateRootCheck := common.CheckStateRoot(ctx, tm.store, tm.sm, b, baseTs)

	await := []async.ErrorFuture{
		minerCheck,
		stateRootCheck,
	}

	await = append(await, msgsChecks...)

	var merr error
	for _, fut := range await {
		if err := fut.AwaitContext(ctx); err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	if merr != nil {
		mulErr := merr.(*multierror.Error)
		mulErr.ErrorFormat = func(es []error) string {
			if len(es) == 1 {
				return fmt.Sprintf("1 error occurred:\n\t* %+v\n\n", es[0])
			}

			points := make([]string, len(es))
			for i, err := range es {
				points[i] = fmt.Sprintf("* %+v", err)
			}

			return fmt.Sprintf(
				"%d errors occurred:\n\t%s\n\n",
				len(es), strings.Join(points, "\n\t"))
		}
		return mulErr
	}

	// Tendermint specific checks.
	height := int64(h.Height) + tm.offset
	log.Infof("Try to access Tendermint RPC from ValidateBlock")
	resp, err := tm.client.Block(ctx, &height)
	if err != nil {
		return xerrors.Errorf("unable to get the Tendermint block at height %d", height)
	}

	val, err := tm.client.Validators(ctx, &height, nil, nil)
	if err != nil {
		return xerrors.Errorf("unable to get the Tendermint block validators at height %d", height)
	}

	var validMinerEudicoAddress address.Address
	var convErr error
	validMinerEudicoAddress, ok := tm.tendermintEudicoAddresses[resp.Block.ProposerAddress.String()]
	if !ok {
		proposerPubKey := findValidatorPubKeyByAddress(val.Validators, resp.Block.ProposerAddress.Bytes())
		if proposerPubKey == nil {
			return xerrors.Errorf("unable to find pubKey for proposer %w", resp.Block.ProposerAddress)
		}

		validMinerEudicoAddress, convErr = getFilecoinAddrByTendermintPubKey(proposerPubKey)
		if convErr != nil {
			return xerrors.Errorf("unable to get proposer addr %w", err)
		}
		tm.tendermintEudicoAddresses[resp.Block.ProposerAddress.String()] = validMinerEudicoAddress
	}
	if b.Header.Miner != validMinerEudicoAddress {
		return xerrors.Errorf("invalid miner address %w in the block header", b.Header.Miner)
	}

	sealed, err := isBlockSealed(b, resp.Block)
	if err != nil {
		log.Infof("block sealed err: %s", err.Error())
		return err
	}
	if !sealed {
		log.Infof("block is not sealed %d", b.Header.Height)
		return xerrors.New("block is not sealed")
	}

	log.Infof("block at @%d is valid", b.Header.Height)

	return nil
}

func (tm *Tendermint) validateBlock(ctx context.Context, b *types.FullBlock) (err error) {
	log.Infof("STARTED ADDITIONAL VALIDATION FOR BLOCK %d", b.Header.Height)
	defer log.Infof("FINISHED ADDITIONAL VALIDATION FOR  %d", b.Header.Height)

	if err := common.BlockSanityChecks(hierarchical.Tendermint, b.Header); err != nil {
		return xerrors.Errorf("incoming header failed basic sanity checks: %w", err)
	}

	h := b.Header

	baseTs, err := tm.store.LoadTipSet(types.NewTipSetKey(h.Parents...))
	if err != nil {
		return xerrors.Errorf("load parent tipset failed (%s): %w", h.Parents, err)
	}

	// fast checks first
	if h.Height != baseTs.Height() {
		return xerrors.Errorf("block height not parent height+1: %d != %d", h.Height, baseTs.Height()+1)
	}

	now := uint64(build.Clock.Now().Unix())
	if h.Timestamp > now+build.AllowableClockDriftSecs {
		return xerrors.Errorf("block was from the future (now=%d, blk=%d): %w", now, h.Timestamp, consensus.ErrTemporal)
	}
	if h.Timestamp > now {
		log.Warn("Got block from the future, but within threshold", h.Timestamp, build.Clock.Now().Unix())
	}

	msgsChecks := common.CheckMsgsWithoutBlockSig(ctx, tm.store, tm.sm, tm.subMgr, tm.r, tm.netName, b, baseTs)

	minerCheck := async.Err(func() error {
		if err := tm.minerIsValid(b.Header.Miner); err != nil {
			return xerrors.Errorf("minerIsValid failed: %w", err)
		}
		return nil
	})

	pweight, err := Weight(context.TODO(), nil, baseTs)
	if err != nil {
		return xerrors.Errorf("getting parent weight: %w", err)
	}

	if types.BigCmp(pweight, b.Header.ParentWeight) != 0 {
		return xerrors.Errorf("parrent weight different: %s (header) != %s (computed)",
			b.Header.ParentWeight, pweight)
	}

	stateRootCheck := common.CheckStateRoot(ctx, tm.store, tm.sm, b, baseTs)

	await := []async.ErrorFuture{
		minerCheck,
		stateRootCheck,
	}

	await = append(await, msgsChecks...)

	var merr error
	for _, fut := range await {
		if err := fut.AwaitContext(ctx); err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	if merr != nil {
		mulErr := merr.(*multierror.Error)
		mulErr.ErrorFormat = func(es []error) string {
			if len(es) == 1 {
				return fmt.Sprintf("1 error occurred:\n\t* %+v\n\n", es[0])
			}

			points := make([]string, len(es))
			for i, err := range es {
				points[i] = fmt.Sprintf("* %+v", err)
			}

			return fmt.Sprintf(
				"%d errors occurred:\n\t%s\n\n",
				len(es), strings.Join(points, "\n\t"))
		}
		return mulErr
	}

	height := int64(h.Height) + tm.offset
	tendermintBlock, err := tm.client.Block(ctx, &height)
	if err != nil {
		return xerrors.Errorf("unable to get the Tendermint block by height %d", height)
	}

	sealed, err := isBlockSealed(b, tendermintBlock.Block)
	if err != nil {
		log.Infof("block sealed err: %s", err.Error())
		return err
	}
	if !sealed {
		log.Infof("block is not sealed %d", b.Header.Height)
		return xerrors.New("block is not sealed")
	}
	return nil
}

func (tm *Tendermint) IsEpochBeyondCurrMax(epoch abi.ChainEpoch) bool {
	if tm.genesis == nil {
		return false
	}

	tendermintLastBlock, err := tm.client.Block(context.TODO(), nil)
	if err != nil {
		//TODO: Tendermint: Discuss what we should return here.
		return false
	}
	//TODO: Tendermint: Discuss what we should return here.
	return tendermintLastBlock.Block.Height+MaxHeightDrift < int64(epoch)
}

func (tm *Tendermint) ValidateBlockPubsub(ctx context.Context, self bool, msg *pubsub.Message) (pubsub.ValidationResult, string) {
	if self {
		return validateLocalBlock(ctx, msg)
	}

	// track validation time
	begin := build.Clock.Now()
	defer func() {
		log.Debugf("block validation time: %s", build.Clock.Since(begin))
	}()

	stats.Record(ctx, metrics.BlockReceived.M(1))

	recordFailureFlagPeer := func(what string) {
		// bv.Validate will flag the peer in that case
		panic(what)
	}

	blk, what, err := decodeAndCheckBlock(msg)
	if err != nil {
		log.Error("got invalid block over pubsub: ", err)
		recordFailureFlagPeer(what)
		return pubsub.ValidationReject, what
	}

	// validate the block meta: the Message CID in the header must match the included messages
	err = common.ValidateMsgMeta(ctx, blk)
	if err != nil {
		log.Warnf("error validating message metadata: %s", err)
		recordFailureFlagPeer("invalid_block_meta")
		return pubsub.ValidationReject, "invalid_block_meta"
	}

	// all good, accept the block
	msg.ValidatorData = blk
	stats.Record(ctx, metrics.BlockValidationSuccess.M(1))
	return pubsub.ValidationAccept, ""
}

func (tm *Tendermint) minerIsValid(maddr address.Address) error {
	switch maddr.Protocol() {
	case address.BLS:
		fallthrough
	case address.SECP256K1:
		return nil
	}

	return xerrors.Errorf("miner address must be a key")
}

// Weight defines weight.
// TODO: should we adopt weight for tendermint?
func Weight(ctx context.Context, stateBs bstore.Blockstore, ts *types.TipSet) (types.BigInt, error) {
	if ts == nil {
		return types.NewInt(0), nil
	}

	return big.NewInt(int64(ts.Height() + 1)), nil
}
