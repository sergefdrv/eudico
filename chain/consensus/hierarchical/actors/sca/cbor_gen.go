// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package sca

import (
	"fmt"
	"io"
	"math"
	"sort"

	abi "github.com/filecoin-project/go-state-types/abi"
	hierarchical "github.com/filecoin-project/lotus/chain/consensus/hierarchical"
	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf
var _ = cid.Undef
var _ = math.E
var _ = sort.Sort

var lengthBufConstructorParams = []byte{130}

func (t *ConstructorParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufConstructorParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.NetworkName (string) (string)
	if len(t.NetworkName) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.NetworkName was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.NetworkName))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.NetworkName)); err != nil {
		return err
	}

	// t.CheckpointPeriod (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.CheckpointPeriod)); err != nil {
		return err
	}

	return nil
}

func (t *ConstructorParams) UnmarshalCBOR(r io.Reader) error {
	*t = ConstructorParams{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.NetworkName (string) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.NetworkName = string(sval)
	}
	// t.CheckpointPeriod (uint64) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.CheckpointPeriod = uint64(extra)

	}
	return nil
}

var lengthBufCheckpointParams = []byte{129}

func (t *CheckpointParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufCheckpointParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Checkpoint ([]uint8) (slice)
	if len(t.Checkpoint) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Checkpoint was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.Checkpoint))); err != nil {
		return err
	}

	if _, err := w.Write(t.Checkpoint[:]); err != nil {
		return err
	}
	return nil
}

func (t *CheckpointParams) UnmarshalCBOR(r io.Reader) error {
	*t = CheckpointParams{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 1 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Checkpoint ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.Checkpoint: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.Checkpoint = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.Checkpoint[:]); err != nil {
		return err
	}
	return nil
}

var lengthBufSCAState = []byte{134}

func (t *SCAState) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufSCAState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.NetworkName (hierarchical.SubnetID) (string)
	if len(t.NetworkName) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.NetworkName was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.NetworkName))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.NetworkName)); err != nil {
		return err
	}

	// t.TotalSubnets (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.TotalSubnets)); err != nil {
		return err
	}

	// t.MinStake (big.Int) (struct)
	if err := t.MinStake.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Subnets (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Subnets); err != nil {
		return xerrors.Errorf("failed to write cid field t.Subnets: %w", err)
	}

	// t.CheckPeriod (abi.ChainEpoch) (int64)
	if t.CheckPeriod >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.CheckPeriod)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.CheckPeriod-1)); err != nil {
			return err
		}
	}

	// t.Checkpoints (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Checkpoints); err != nil {
		return xerrors.Errorf("failed to write cid field t.Checkpoints: %w", err)
	}

	return nil
}

func (t *SCAState) UnmarshalCBOR(r io.Reader) error {
	*t = SCAState{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 6 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.NetworkName (hierarchical.SubnetID) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.NetworkName = hierarchical.SubnetID(sval)
	}
	// t.TotalSubnets (uint64) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.TotalSubnets = uint64(extra)

	}
	// t.MinStake (big.Int) (struct)

	{

		if err := t.MinStake.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.MinStake: %w", err)
		}

	}
	// t.Subnets (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Subnets: %w", err)
		}

		t.Subnets = c

	}
	// t.CheckPeriod (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			return err
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				return fmt.Errorf("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			return fmt.Errorf("wrong type for int64 field: %d", maj)
		}

		t.CheckPeriod = abi.ChainEpoch(extraI)
	}
	// t.Checkpoints (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Checkpoints: %w", err)
		}

		t.Checkpoints = c

	}
	return nil
}

var lengthBufSubnet = []byte{137}

func (t *Subnet) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufSubnet); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.ID (hierarchical.SubnetID) (string)
	if len(t.ID) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.ID was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.ID))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.ID)); err != nil {
		return err
	}

	// t.ParentID (hierarchical.SubnetID) (string)
	if len(t.ParentID) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.ParentID was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.ParentID))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.ParentID)); err != nil {
		return err
	}

	// t.Stake (big.Int) (struct)
	if err := t.Stake.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Funds (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Funds); err != nil {
		return xerrors.Errorf("failed to write cid field t.Funds: %w", err)
	}

	// t.CrossMsgs (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.CrossMsgs); err != nil {
		return xerrors.Errorf("failed to write cid field t.CrossMsgs: %w", err)
	}

	// t.Nonce (uint64) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Nonce)); err != nil {
		return err
	}

	// t.CircSupply (big.Int) (struct)
	if err := t.CircSupply.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Status (sca.Status) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Status)); err != nil {
		return err
	}

	// t.PrevCheckpoint (schema.Checkpoint) (struct)
	if err := t.PrevCheckpoint.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *Subnet) UnmarshalCBOR(r io.Reader) error {
	*t = Subnet{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 9 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.ID (hierarchical.SubnetID) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.ID = hierarchical.SubnetID(sval)
	}
	// t.ParentID (hierarchical.SubnetID) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.ParentID = hierarchical.SubnetID(sval)
	}
	// t.Stake (big.Int) (struct)

	{

		if err := t.Stake.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Stake: %w", err)
		}

	}
	// t.Funds (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Funds: %w", err)
		}

		t.Funds = c

	}
	// t.CrossMsgs (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.CrossMsgs: %w", err)
		}

		t.CrossMsgs = c

	}
	// t.Nonce (uint64) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Nonce = uint64(extra)

	}
	// t.CircSupply (big.Int) (struct)

	{

		if err := t.CircSupply.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.CircSupply: %w", err)
		}

	}
	// t.Status (sca.Status) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Status = Status(extra)

	}
	// t.PrevCheckpoint (schema.Checkpoint) (struct)

	{

		if err := t.PrevCheckpoint.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.PrevCheckpoint: %w", err)
		}

	}
	return nil
}

var lengthBufFundParams = []byte{129}

func (t *FundParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufFundParams); err != nil {
		return err
	}

	// t.Value (big.Int) (struct)
	if err := t.Value.MarshalCBOR(w); err != nil {
		return err
	}
	return nil
}

func (t *FundParams) UnmarshalCBOR(r io.Reader) error {
	*t = FundParams{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 1 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.Value (big.Int) (struct)

	{

		if err := t.Value.UnmarshalCBOR(br); err != nil {
			return xerrors.Errorf("unmarshaling t.Value: %w", err)
		}

	}
	return nil
}

var lengthBufSubnetIDParam = []byte{129}

func (t *SubnetIDParam) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufSubnetIDParam); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.ID (string) (string)
	if len(t.ID) > cbg.MaxLength {
		return xerrors.Errorf("Value in field t.ID was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len(t.ID))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(t.ID)); err != nil {
		return err
	}
	return nil
}

func (t *SubnetIDParam) UnmarshalCBOR(r io.Reader) error {
	*t = SubnetIDParam{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 1 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.ID (string) (string)

	{
		sval, err := cbg.ReadStringBuf(br, scratch)
		if err != nil {
			return err
		}

		t.ID = string(sval)
	}
	return nil
}
