package replace

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/go-state-types/exitcode"
	actor "github.com/filecoin-project/lotus/chain/consensus/actors"
	"github.com/filecoin-project/lotus/chain/consensus/hierarchical/atomic"
	"github.com/filecoin-project/specs-actors/v7/actors/builtin"
	"github.com/filecoin-project/specs-actors/v7/actors/runtime"
	cid "github.com/ipfs/go-cid"
	xerrors "golang.org/x/xerrors"
)

//go:generate go run ./gen/gen.go

// example "Replace" actor that atomically replaces the cid from one owner
// to the other.

var _ runtime.VMActor = ReplaceActor{}
var _ atomic.LockableActor = ReplaceActor{}

// ReplaceState determines the actor state.
// FIXME: We are using a non-efficient locking strategy for now
// where the whole map is locked for an atomic execution.
// We could use a more fine-grained approach. Consider it in the next
// iteration.
type ReplaceState struct {
	Owners *atomic.LockedState
}

type Owners struct {
	M map[string]cid.Cid
}

func (o *Owners) Merge(other atomic.LockableState) error {
	tt, ok := other.(*Owners)
	if !ok {
		return xerrors.Errorf("type of LockableState not Owners")
	}

	for k, v := range tt.M {
		_, ok := o.M[k]
		if ok {
			return xerrors.Errorf("merge conflict. key for owner already set")
		}
		o.M[k] = v
	}
	return nil

}

func ConstructState() *ReplaceState {
	owners := &Owners{M: map[string]cid.Cid{}}
	return &ReplaceState{Owners: atomic.WrapLockableState(owners)}
}

type ReplaceActor struct{}

func (a ReplaceActor) Exports() []interface{} {
	return []interface{}{
		builtin.MethodConstructor: a.Constructor,
		atomic.MethodLock:         a.Lock,
		atomic.MethodMerge:        a.Merge,
		atomic.MethodAbort:        a.Abort,
		5:                         a.Replace,
		6:                         a.Own,
	}
}

func (a ReplaceActor) Code() cid.Cid {
	return actor.ReplaceActorCodeID
}

func (a ReplaceActor) IsSingleton() bool {
	return false
}

func (a ReplaceActor) State() cbor.Er {
	return new(ReplaceState)
}

func (a ReplaceActor) Constructor(rt runtime.Runtime, _ *abi.EmptyValue) *abi.EmptyValue {
	rt.ValidateImmediateCallerType(builtin.InitActorCodeID)
	rt.StateCreate(ConstructState())
	return nil
}

type OwnParams struct {
	Seed string
}

func (a ReplaceActor) Own(rt runtime.Runtime, params *OwnParams) *abi.EmptyValue {
	rt.ValidateImmediateCallerAcceptAny()

	var (
		st  ReplaceState
		err error
	)
	rt.StateTransaction(&st, func() {
		ValidateLockedState(rt, &st)
		own := st.Owners.State().(*Owners)
		_, ok := own.M[rt.Caller().String()]
		if ok {
			rt.Abortf(exitcode.ErrIllegalState, "address already owning something")
		}
		own.M[rt.Caller().String()], err = abi.CidBuilder.Sum([]byte(params.Seed))
		builtin.RequireNoErr(rt, err, exitcode.ErrIllegalArgument, "error computing cid")
	})

	return nil
}

type ReplaceParams struct {
	Addr address.Address
}

func (a ReplaceActor) Replace(rt runtime.Runtime, params *ReplaceParams) *abi.EmptyValue {
	rt.ValidateImmediateCallerAcceptAny()

	var st ReplaceState
	rt.StateTransaction(&st, func() {
		ValidateLockedState(rt, &st)
		own := st.Owners.State().(*Owners)
		// Replace
		own.M[rt.Caller().String()], own.M[params.Addr.String()] =
			own.M[params.Addr.String()], own.M[rt.Caller().String()]
	})

	return nil
}

func (a ReplaceActor) Lock(rt runtime.Runtime, params *atomic.LockParams) cid.Cid {
	// Anyone can lock the state
	rt.ValidateImmediateCallerAcceptAny()

	var st ReplaceState
	rt.StateTransaction(&st, func() {
		switch params.Method {
		case 5:
			builtin.RequireNoErr(rt, st.Owners.LockState(), exitcode.ErrIllegalArgument, "error locking state")
		default:
			rt.Abortf(exitcode.ErrIllegalArgument, "provided method doesn't support atomic execution. No need to lock")
		}
	})

	c, err := st.Owners.Cid()
	builtin.RequireNoErr(rt, err, exitcode.ErrIllegalArgument, "error computing Cid for locked state")
	return c
}

func (st *ReplaceState) unlock(rt runtime.Runtime) {
	builtin.RequireNoErr(rt, st.Owners.UnlockState(), exitcode.ErrIllegalArgument, "error unlocking state")
}

func (a ReplaceActor) Merge(rt runtime.Runtime, params *atomic.UnlockParams) *abi.EmptyValue {
	// FIXME: Who should call this function? System actor or the caller of Lock()?
	// Deferring to when we figure out the end-to-end protocol.
	rt.ValidateImmediateCallerAcceptAny()

	var st ReplaceState
	rt.StateTransaction(&st, func() {
		switch params.Params.Method {
		case 5:
			builtin.RequireNoErr(rt, st.Owners.S.Merge(params.Output), exitcode.ErrIllegalState, "error merging output")
			st.unlock(rt)
		default:
			rt.Abortf(exitcode.ErrIllegalArgument, "this method has nothing to merge")
		}
	})

	return nil
}

func (a ReplaceActor) Abort(rt runtime.Runtime, params *atomic.LockParams) *abi.EmptyValue {
	// FIXME: Who should call this function? System actor or the caller of Lock()?
	// Deferring to when we figure out the end-to-end protocol.
	rt.ValidateImmediateCallerAcceptAny()

	var st ReplaceState
	rt.StateTransaction(&st, func() {
		switch params.Method {
		case 5:
			st.unlock(rt)
		default:
			rt.Abortf(exitcode.ErrIllegalArgument, "this method has nothing to unlock")
		}
	})

	return nil
}

func ValidateLockedState(rt runtime.Runtime, st *ReplaceState) {
	builtin.RequireNoErr(rt,
		atomic.ValidateIfLocked([]*atomic.LockedState{st.Owners}...),
		exitcode.ErrIllegalState, "state locked")
}
