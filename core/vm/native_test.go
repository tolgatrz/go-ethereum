package vm

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
)

const maxRun = 1000

func TestNative(t *testing.T) {
	db, _ := ethdb.NewMemDatabase()
	sender := state.NewStateObject(common.Address{}, db)

	var (
		env   = NewEnv()
		input = []byte{0, 0}
	)

	tstart := time.Now()
	program := NewProgram()
	code := common.Hex2Bytes("600a600a01600a600a01600a600a01600a600a01600a600a01600a600a01600a600a01600a600a01600a600a01600a600a01")
	err := AttachProgram(program, code)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for i := 0; i < maxRun; i++ {
		context := NewContext(sender, sender, big.NewInt(100), big.NewInt(10000), big.NewInt(0))
		_, err = RunProgram(program, env, context, input)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
	fmt.Println("native", time.Since(tstart))

	tstart = time.Now()
	DisableJit = true
	for i := 0; i < maxRun; i++ {
		context := NewContext(sender, sender, big.NewInt(100), big.NewInt(10000), big.NewInt(0))
		context.Code = code
		context.CodeAddr = &common.Address{}
		_, err := New(env).Run(context, input)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

	}
	fmt.Println("vm", time.Since(tstart))
}

type Env struct {
	gasLimit *big.Int
	depth    int
}

func NewEnv() *Env {
	return &Env{big.NewInt(10000), 0}
}

func (self *Env) Origin() common.Address { return common.Address{} }
func (self *Env) BlockNumber() *big.Int  { return big.NewInt(0) }
func (self *Env) AddStructLog(log StructLog) {
}
func (self *Env) StructLogs() []StructLog {
	return nil
}

//func (self *Env) PrevHash() []byte      { return self.parent }
func (self *Env) Coinbase() common.Address { return common.Address{} }
func (self *Env) Time() uint64             { return uint64(time.Now().Unix()) }
func (self *Env) Difficulty() *big.Int     { return big.NewInt(0) }
func (self *Env) State() *state.StateDB    { return nil }
func (self *Env) GasLimit() *big.Int       { return self.gasLimit }
func (self *Env) VmType() Type             { return StdVmTy }
func (self *Env) GetHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Sha3([]byte(big.NewInt(int64(n)).String())))
}
func (self *Env) AddLog(log *state.Log) {
}
func (self *Env) Depth() int     { return self.depth }
func (self *Env) SetDepth(i int) { self.depth = i }
func (self *Env) Transfer(from, to Account, amount *big.Int) error {
	return nil
}
func (self *Env) Call(caller ContextRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return nil, nil
}
func (self *Env) CallCode(caller ContextRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return nil, nil
}
func (self *Env) Create(caller ContextRef, data []byte, gas, price, value *big.Int) ([]byte, error, ContextRef) {
	return nil, nil, nil
}
