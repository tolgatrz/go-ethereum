package core

import "github.com/ethereum/go-ethereum/pow"

type Finaliser interface {
	Verify(block pow.Block) bool
}
