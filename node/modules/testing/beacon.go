package testing

import (
	"time"

	"github.com/filecoin-project/lotus/build"
	beacon2 "github.com/filecoin-project/lotus/chainlotus/beacon"
)

func RandomBeacon() (beacon2.Schedule, error) {
	return beacon2.Schedule{
		{Start: 0,
			Beacon: beacon2.NewMockBeacon(time.Duration(build.BlockDelaySecs) * time.Second),
		}}, nil
}
