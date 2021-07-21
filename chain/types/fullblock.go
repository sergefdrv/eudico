package types

import "github.com/ipfs/go-cid"

type FullBlock struct {
	Header        SyncBlock
	BlsMessages   []*Message
	SecpkMessages []*SignedMessage
}

func (fb *FullBlock) Cid() cid.Cid {
	return fb.Header.Cid()
}
