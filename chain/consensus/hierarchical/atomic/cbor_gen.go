// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package atomic

import (
	"fmt"
	"io"
	"math"
	"sort"

	abi "github.com/filecoin-project/go-state-types/abi"
	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf
var _ = cid.Undef
var _ = math.E
var _ = sort.Sort

var lengthBufMergeParams = []byte{129}

func (t *MergeParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufMergeParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.State ([]uint8) (slice)
	if len(t.State) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.State was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.State))); err != nil {
		return err
	}

	if _, err := w.Write(t.State[:]); err != nil {
		return err
	}
	return nil
}

func (t *MergeParams) UnmarshalCBOR(r io.Reader) error {
	*t = MergeParams{}

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

	// t.State ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.State: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.State = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.State[:]); err != nil {
		return err
	}
	return nil
}

var lengthBufUnlockParams = []byte{130}

func (t *UnlockParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufUnlockParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Params (atomic.LockParams) (struct)
	if err := t.Params.MarshalCBOR(w); err != nil {
		return err
	}

	// t.State ([]uint8) (slice)
	if len(t.State) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.State was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.State))); err != nil {
		return err
	}

	if _, err := w.Write(t.State[:]); err != nil {
		return err
	}
	return nil
}

func (t *UnlockParams) UnmarshalCBOR(r io.Reader) error {
	*t = UnlockParams{}

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

	// t.Params (atomic.LockParams) (struct)

	{

		b, err := br.ReadByte()
		if err != nil {
			return err
		}
		if b != cbg.CborNull[0] {
			if err := br.UnreadByte(); err != nil {
				return err
			}
			t.Params = new(LockParams)
			if err := t.Params.UnmarshalCBOR(br); err != nil {
				return xerrors.Errorf("unmarshaling t.Params pointer: %w", err)
			}
		}

	}
	// t.State ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.State: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.State = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.State[:]); err != nil {
		return err
	}
	return nil
}

var lengthBufLockParams = []byte{130}

func (t *LockParams) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufLockParams); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Method (abi.MethodNum) (uint64)

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Method)); err != nil {
		return err
	}

	// t.Params ([]uint8) (slice)
	if len(t.Params) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Params was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.Params))); err != nil {
		return err
	}

	if _, err := w.Write(t.Params[:]); err != nil {
		return err
	}
	return nil
}

func (t *LockParams) UnmarshalCBOR(r io.Reader) error {
	*t = LockParams{}

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

	// t.Method (abi.MethodNum) (uint64)

	{

		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}
		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}
		t.Method = abi.MethodNum(extra)

	}
	// t.Params ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.Params: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.Params = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.Params[:]); err != nil {
		return err
	}
	return nil
}

var lengthBufLockedOutput = []byte{129}

func (t *LockedOutput) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufLockedOutput); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Cid (cid.Cid) (struct)

	if err := cbg.WriteCidBuf(scratch, w, t.Cid); err != nil {
		return xerrors.Errorf("failed to write cid field t.Cid: %w", err)
	}

	return nil
}

func (t *LockedOutput) UnmarshalCBOR(r io.Reader) error {
	*t = LockedOutput{}

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

	// t.Cid (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.Cid: %w", err)
		}

		t.Cid = c

	}
	return nil
}

var lengthBufLockedState = []byte{130}

func (t *LockedState) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufLockedState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Lock (bool) (bool)
	if err := cbg.WriteBool(w, t.Lock); err != nil {
		return err
	}

	// t.S ([]uint8) (slice)
	if len(t.S) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.S was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.S))); err != nil {
		return err
	}

	if _, err := w.Write(t.S[:]); err != nil {
		return err
	}
	return nil
}

func (t *LockedState) UnmarshalCBOR(r io.Reader) error {
	*t = LockedState{}

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

	// t.Lock (bool) (bool)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajOther {
		return fmt.Errorf("booleans must be major type 7")
	}
	switch extra {
	case 20:
		t.Lock = false
	case 21:
		t.Lock = true
	default:
		return fmt.Errorf("booleans are either major type 7, value 20 or 21 (got %d)", extra)
	}
	// t.S ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.S: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}

	if extra > 0 {
		t.S = make([]uint8, extra)
	}

	if _, err := io.ReadFull(br, t.S[:]); err != nil {
		return err
	}
	return nil
}
