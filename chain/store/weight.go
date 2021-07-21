package store

import (
	"context"

	"github.com/filecoin-project/lotus/chain/types"
)

func (cs *ChainStore) Weight(ctx context.Context, ts types.SyncTs) (types.BigInt, error) {
	return cs.weight(ctx, cs.StateBlockstore(), ts)
}
