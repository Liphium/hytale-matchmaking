package players_routes_test

import (
	"testing"

	"github.com/Liphium/hytale-matchmaking/starter"
	"github.com/Liphium/magic/v2"
)

func TestMain(m *testing.M) {
	magic.PrepareTesting(m, starter.BuildMagicConfig())
}
