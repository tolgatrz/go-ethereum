package common

import "math/big"

type IntPool struct {
	pool chan *big.Int
}

// NewPool creates a new pool of big.Int
func NewIntPool(max int) *IntPool {
	return &IntPool{
		pool: make(chan *big.Int, max),
	}
}

// Borrow an Int from the pool.
func (p *IntPool) Get() *big.Int {
	var c *big.Int
	select {
	case c = <-p.pool:
	default:
		c = new(big.Int)
	}
	return c
}

// Return returns an Intto the pool.
func (p *IntPool) Put(i ...*big.Int) {
	for _, c := range i {
		select {
		case p.pool <- c:
		default:
			// let it go, let it go...
			return
		}
	}
}
