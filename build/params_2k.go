//go:build debug || 2k
// +build debug 2k

package build

import (
	"os"
	"strconv"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/network"
	"github.com/filecoin-project/lotus/chain/actors/policy"
)

const BootstrappersFile = ""
const GenesisFile = ""

const GenesisNetworkVersion = network.Version15

var UpgradeBreezeHeight = abi.ChainEpoch(-1)

const BreezeGasTampingDuration = 0

var UpgradeSmokeHeight = abi.ChainEpoch(-1)
var UpgradeIgnitionHeight = abi.ChainEpoch(-2)
var UpgradeRefuelHeight = abi.ChainEpoch(-3)
var UpgradeTapeHeight = abi.ChainEpoch(-4)

var UpgradeAssemblyHeight = abi.ChainEpoch(-5)
var UpgradeLiftoffHeight = abi.ChainEpoch(-6)

var UpgradeKumquatHeight = abi.ChainEpoch(-7)
var UpgradeCalicoHeight = abi.ChainEpoch(-9)
var UpgradePersianHeight = abi.ChainEpoch(-10)
var UpgradeOrangeHeight = abi.ChainEpoch(-11)
var UpgradeClausHeight = abi.ChainEpoch(-12)

var UpgradeTrustHeight = abi.ChainEpoch(-13)

var UpgradeNorwegianHeight = abi.ChainEpoch(-14)

var UpgradeTurboHeight = abi.ChainEpoch(-15)

var UpgradeHyperdriveHeight = abi.ChainEpoch(-16)

var UpgradeChocolateHeight = abi.ChainEpoch(-17)

var UpgradeOhSnapHeight = abi.ChainEpoch(-18)

var DrandSchedule = map[abi.ChainEpoch]DrandEnum{
	0: DrandMainnet,
}

var SupportedProofTypes = []abi.RegisteredSealProof{
	abi.RegisteredSealProof_StackedDrg2KiBV1,
	abi.RegisteredSealProof_StackedDrg8MiBV1,
}
var ConsensusMinerMinPower = abi.NewStoragePower(2048)
var MinVerifiedDealSize = abi.NewStoragePower(256)
var PreCommitChallengeDelay = abi.ChainEpoch(10)

func init() {
	policy.SetSupportedProofTypes(SupportedProofTypes...)
	policy.SetConsensusMinerMinPower(ConsensusMinerMinPower)
	policy.SetMinVerifiedDealSize(MinVerifiedDealSize)
	policy.SetPreCommitChallengeDelay(PreCommitChallengeDelay)

	getUpgradeHeight := func(ev string, def abi.ChainEpoch) abi.ChainEpoch {
		hs, found := os.LookupEnv(ev)
		if found {
			h, err := strconv.Atoi(hs)
			if err != nil {
				log.Panicf("failed to parse %s env var", ev)
			}

			return abi.ChainEpoch(h)
		}

		return def
	}

	UpgradeBreezeHeight = getUpgradeHeight("LOTUS_BREEZE_HEIGHT", UpgradeBreezeHeight)
	UpgradeSmokeHeight = getUpgradeHeight("LOTUS_SMOKE_HEIGHT", UpgradeSmokeHeight)
	UpgradeIgnitionHeight = getUpgradeHeight("LOTUS_IGNITION_HEIGHT", UpgradeIgnitionHeight)
	UpgradeRefuelHeight = getUpgradeHeight("LOTUS_REFUEL_HEIGHT", UpgradeRefuelHeight)
	UpgradeTapeHeight = getUpgradeHeight("LOTUS_TAPE_HEIGHT", UpgradeTapeHeight)
	UpgradeAssemblyHeight = getUpgradeHeight("LOTUS_ACTORSV2_HEIGHT", UpgradeAssemblyHeight)
	UpgradeLiftoffHeight = getUpgradeHeight("LOTUS_LIFTOFF_HEIGHT", UpgradeLiftoffHeight)
	UpgradeKumquatHeight = getUpgradeHeight("LOTUS_KUMQUAT_HEIGHT", UpgradeKumquatHeight)
	UpgradeCalicoHeight = getUpgradeHeight("LOTUS_CALICO_HEIGHT", UpgradeCalicoHeight)
	UpgradePersianHeight = getUpgradeHeight("LOTUS_PERSIAN_HEIGHT", UpgradePersianHeight)
	UpgradeOrangeHeight = getUpgradeHeight("LOTUS_ORANGE_HEIGHT", UpgradeOrangeHeight)
	UpgradeClausHeight = getUpgradeHeight("LOTUS_CLAUS_HEIGHT", UpgradeClausHeight)
	UpgradeTrustHeight = getUpgradeHeight("LOTUS_ACTORSV3_HEIGHT", UpgradeTrustHeight)
	UpgradeNorwegianHeight = getUpgradeHeight("LOTUS_NORWEGIAN_HEIGHT", UpgradeNorwegianHeight)
	UpgradeTurboHeight = getUpgradeHeight("LOTUS_ACTORSV4_HEIGHT", UpgradeTurboHeight)
	UpgradeHyperdriveHeight = getUpgradeHeight("LOTUS_HYPERDRIVE_HEIGHT", UpgradeHyperdriveHeight)
	UpgradeChocolateHeight = getUpgradeHeight("LOTUS_CHOCOLATE_HEIGHT", UpgradeChocolateHeight)
	UpgradeOhSnapHeight = getUpgradeHeight("LOTUS_OHSNAP_HEIGHT", UpgradeOhSnapHeight)

	BuildType |= Build2k

}

const BlockDelaySecs = uint64(1)

const PropagationDelaySecs = uint64(1)

// SlashablePowerDelay is the number of epochs after ElectionPeriodStart, after
// which the miner is slashed
//
// Epochs
const SlashablePowerDelay = 20

// Epochs
const InteractivePoRepConfidence = 6

const BootstrapPeerThreshold = 1

var WhitelistedBlock = cid.Undef

const DelegatedPoWFinality = 3
const PoWFinality = 3
const MirFinality = 2
const DummyFinality = 1
const FilecoinFinality = 5

const DelegatedPoWCheckpointPeriod = 10
const PoWCheckpointPeriod = 10
const MirCheckpointPeriod = 10
const DummyCheckpointPeriod = 10
const FilecoinCheckpointPeriod = 10

const GenesisPoWTarget = "2019783675352289407433363"
