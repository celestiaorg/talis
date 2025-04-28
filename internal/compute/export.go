package compute

import (
	"github.com/celestiaorg/talis/internal/types"
)

// Provider returns a new VirtFusion provider instance
func Provider() (types.Provider, error) {
	return NewVirtFusionProvider()
}
