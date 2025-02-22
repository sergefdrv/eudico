//go:build !debug && !2k && !testground && !calibnet && !nerpanet && !butterflynet && !interopnet
// +build !debug,!2k,!testground,!calibnet,!nerpanet,!butterflynet,!interopnet

package build

import (
	"math"
	"os"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/network"
	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
)

var DrandSchedule = map[abi.ChainEpoch]DrandEnum{
	0:                  DrandIncentinet,
	UpgradeSmokeHeight: DrandMainnet,
}

const GenesisNetworkVersion = network.Version0

const BootstrappersFile = "mainnet.pi"
const GenesisFile = "mainnet.car"

const UpgradeBreezeHeight = 41280

const BreezeGasTampingDuration = 120

const UpgradeSmokeHeight = 51000

const UpgradeIgnitionHeight = 94000
const UpgradeRefuelHeight = 130800

const UpgradeAssemblyHeight = 138720

const UpgradeTapeHeight = 140760

// This signals our tentative epoch for mainnet launch. Can make it later, but not earlier.
// Miners, clients, developers, custodians all need time to prepare.
// We still have upgrades and state changes to do, but can happen after signaling timing here.
const UpgradeLiftoffHeight = 148888

const UpgradeKumquatHeight = 170000

const UpgradeCalicoHeight = 265200
const UpgradePersianHeight = UpgradeCalicoHeight + (builtin2.EpochsInHour * 60)

const UpgradeOrangeHeight = 336458

// 2020-12-22T02:00:00Z
var UpgradeClausHeight = abi.ChainEpoch(343200)

// 2021-03-04T00:00:30Z
const UpgradeTrustHeight = 550321

// 2021-04-12T22:00:00Z
const UpgradeNorwegianHeight = 665280

// 2021-04-29T06:00:00Z
const UpgradeTurboHeight = 712320

// 2021-06-30T22:00:00Z
const UpgradeHyperdriveHeight = 892800

// 2021-10-26T13:30:00Z
const UpgradeChocolateHeight = 1231620

// 2022-03-01T15:00:00Z
var UpgradeOhSnapHeight = abi.ChainEpoch(1594680)

var SupportedProofTypes = []abi.RegisteredSealProof{
	abi.RegisteredSealProof_StackedDrg32GiBV1,
	abi.RegisteredSealProof_StackedDrg64GiBV1,
}
var ConsensusMinerMinPower = abi.NewStoragePower(10 << 40)
var MinVerifiedDealSize = abi.NewStoragePower(1 << 20)
var PreCommitChallengeDelay = abi.ChainEpoch(150)

func init() {
	if os.Getenv("LOTUS_USE_TEST_ADDRESSES") != "1" {
		SetAddressNetwork(address.Mainnet)
	}

	if os.Getenv("LOTUS_DISABLE_SNAPDEALS") == "1" {
		UpgradeOhSnapHeight = math.MaxInt64
	}

	Devnet = false

	BuildType = BuildMainnet
}

const BlockDelaySecs = uint64(builtin2.EpochDurationSeconds)

const PropagationDelaySecs = uint64(6)

// BootstrapPeerThreshold is the minimum number peers we need to track for a sync worker to start
const BootstrapPeerThreshold = 4

// we skip checks on message validity in this block to sidestep the zero-bls signature
var WhitelistedBlock = MustParseCid("bafy2bzaceapyg2uyzk7vueh3xccxkuwbz3nxewjyguoxvhx77malc2lzn2ybi")

const DelegatedPoWFinality = 5
const PoWFinality = 5
const MirFinality = 2
const DummyFinality = 1
const FilecoinFinality = 5

const DelegatedPoWCheckpointPeriod = 10
const PoWCheckpointPeriod = 10
const MirCheckpointPeriod = 10
const DummyCheckpointPeriod = 10
const FilecoinCheckpointPeriod = 10

const GenesisPoWTarget = "4519783675352289407433363"
