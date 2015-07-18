package vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

type instrFn func(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack)

type instruction struct {
	op   OpCode
	pc   uint64
	fn   instrFn
	data *big.Int
}

func opAdd(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	stack.push(U256(new(big.Int).Add(x, y)))
}

func opSub(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	stack.push(U256(new(big.Int).Sub(x, y)))
}

func opMul(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	stack.push(U256(new(big.Int).Mul(x, y)))
}

func opDiv(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	base := new(big.Int)
	x, y := stack.pop(), stack.pop()

	if y.Cmp(common.Big0) != 0 {
		base.Div(x, y)
	}

	// pop result back on the stack
	stack.push(U256(base))
}

func opSdiv(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	base := new(big.Int)
	x, y := S256(stack.pop()), S256(stack.pop())

	if y.Cmp(common.Big0) == 0 {
		base.Set(common.Big0)
	} else {
		n := new(big.Int)
		if new(big.Int).Mul(x, y).Cmp(common.Big0) < 0 {
			n.SetInt64(-1)
		} else {
			n.SetInt64(1)
		}

		base.Div(x.Abs(x), y.Abs(y)).Mul(base, n)

		U256(base)
	}

	stack.push(base)
}

func opMod(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	base := new(big.Int)
	x, y := stack.pop(), stack.pop()

	if y.Cmp(common.Big0) == 0 {
		base.Set(common.Big0)
	} else {
		base.Mod(x, y)
	}

	U256(base)

	stack.push(base)
}

func opSmod(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	base := new(big.Int)
	x, y := S256(stack.pop()), S256(stack.pop())

	if y.Cmp(common.Big0) == 0 {
		base.Set(common.Big0)
	} else {
		n := new(big.Int)
		if x.Cmp(common.Big0) < 0 {
			n.SetInt64(-1)
		} else {
			n.SetInt64(1)
		}

		base.Mod(x.Abs(x), y.Abs(y)).Mul(base, n)

		U256(base)
	}

	stack.push(base)
}

func opExp(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	base := new(big.Int)
	x, y := stack.pop(), stack.pop()

	base.Exp(x, y, Pow256)

	U256(base)

	stack.push(base)
}

func opSignExtend(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	back := stack.pop()
	if back.Cmp(big.NewInt(31)) < 0 {
		bit := uint(back.Uint64()*8 + 7)
		num := stack.pop()
		mask := new(big.Int).Lsh(common.Big1, bit)
		mask.Sub(mask, common.Big1)
		if common.BitTest(num, int(bit)) {
			num.Or(num, mask.Not(mask))
		} else {
			num.And(num, mask)
		}

		num = U256(num)

		stack.push(num)
	}
}

func opNot(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(U256(new(big.Int).Not(stack.pop())))
}

func opLt(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	// x < y
	if x.Cmp(y) < 0 {
		stack.push(common.BigTrue)
	} else {
		stack.push(common.BigFalse)
	}
}

func opGt(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	// x > y
	if x.Cmp(y) > 0 {
		stack.push(common.BigTrue)
	} else {
		stack.push(common.BigFalse)
	}
}

func opSlt(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := S256(stack.pop()), S256(stack.pop())

	// x < y
	if x.Cmp(S256(y)) < 0 {
		stack.push(common.BigTrue)
	} else {
		stack.push(common.BigFalse)
	}
}

func opSgt(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := S256(stack.pop()), S256(stack.pop())

	// x > y
	if x.Cmp(y) > 0 {
		stack.push(common.BigTrue)
	} else {
		stack.push(common.BigFalse)
	}
}

func opEq(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	// x == y
	if x.Cmp(y) == 0 {
		stack.push(common.BigTrue)
	} else {
		stack.push(common.BigFalse)
	}
}

func opIszero(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x := stack.pop()
	if x.Cmp(common.BigFalse) > 0 {
		stack.push(common.BigFalse)
	} else {
		stack.push(common.BigTrue)
	}
}

func opAnd(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	stack.push(new(big.Int).And(x, y))
}
func opOr(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	stack.push(new(big.Int).Or(x, y))
}
func opXor(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	x, y := stack.pop(), stack.pop()

	stack.push(new(big.Int).Xor(x, y))
}
func opByte(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	base := new(big.Int)
	th, val := stack.pop(), stack.pop()

	if th.Cmp(big.NewInt(32)) < 0 {
		byt := big.NewInt(int64(common.LeftPadBytes(val.Bytes(), 32)[th.Int64()]))

		base.Set(byt)
	} else {
		base.Set(common.BigFalse)
	}

	stack.push(base)
}
func opAddmod(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	base := new(big.Int)
	x := stack.pop()
	y := stack.pop()
	z := stack.pop()

	if z.Cmp(Zero) > 0 {
		add := new(big.Int).Add(x, y)
		base.Mod(add, z)

		base = U256(base)
	}

	stack.push(base)
}
func opMulmod(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	base := new(big.Int)
	x := stack.pop()
	y := stack.pop()
	z := stack.pop()

	if z.Cmp(Zero) > 0 {
		mul := new(big.Int).Mul(x, y)
		base.Mod(mul, z)

		U256(base)
	}

	stack.push(base)
}

func opSha3(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	offset, size := stack.pop(), stack.pop()
	hash := crypto.Sha3(memory.Get(offset.Int64(), size.Int64()))

	stack.push(common.BigD(hash))
}

func opAddress(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(common.Bytes2Big(context.Address().Bytes()))
}

func opBalance(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	addr := common.BigToAddress(stack.pop())
	balance := env.State().GetBalance(addr)

	stack.push(new(big.Int).Set(balance))
}

func opOrigin(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(env.Origin().Big())
}

func opCaller(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(common.Bytes2Big(context.caller.Address().Bytes()))
}

func opCallValue(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(context.value))
}

func opCalldataLoad(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(common.Bytes2Big(getData(context.Input, stack.pop(), common.Big32)))
}

func opCalldataSize(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(big.NewInt(int64(len(context.Input))))
}

func opCalldataCopy(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	var (
		mOff = stack.pop()
		cOff = stack.pop()
		l    = stack.pop()
	)
	memory.Set(mOff.Uint64(), l.Uint64(), getData(context.Input, cOff, l))
}

func opExtCodeSize(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	addr := common.BigToAddress(stack.pop())
	l := big.NewInt(int64(len(env.State().GetCode(addr))))
	stack.push(l)
}

func opCodeSize(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	l := big.NewInt(int64(len(context.Code)))
	stack.push(l)
}

func opCodeCopy(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	var (
		mOff = stack.pop()
		cOff = stack.pop()
		l    = stack.pop()
	)
	codeCopy := getData(context.Code, cOff, l)

	memory.Set(mOff.Uint64(), l.Uint64(), codeCopy)
}

func opExtCodeCopy(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	var (
		addr = common.BigToAddress(stack.pop())
		mOff = stack.pop()
		cOff = stack.pop()
		l    = stack.pop()
	)
	codeCopy := getData(env.State().GetCode(addr), cOff, l)

	memory.Set(mOff.Uint64(), l.Uint64(), codeCopy)
}

func opGasprice(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(context.Price))
}

func opBlockhash(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	num := stack.pop()

	n := new(big.Int).Sub(env.BlockNumber(), common.Big257)
	if num.Cmp(n) > 0 && num.Cmp(env.BlockNumber()) < 0 {
		stack.push(env.GetHash(num.Uint64()).Big())
	} else {
		stack.push(common.Big0)
	}
}

func opCoinbase(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(env.Coinbase().Big())
}

func opTimestamp(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).SetUint64(env.Time()))
}

func opNumber(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(U256(env.BlockNumber()))
}

func opDifficulty(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(env.Difficulty()))
}

func opGasLimit(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(env.GasLimit()))
}

func opPop(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.pop()
}

func opPush(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(data))
}

func opDup(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.dup(int(data.Int64()))
}

func opSwap(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.swap(int(data.Int64()))
}

func opLog(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	n := int(data.Int64())
	topics := make([]common.Hash, n)
	mStart, mSize := stack.pop(), stack.pop()
	for i := 0; i < n; i++ {
		topics[i] = common.BigToHash(stack.pop())
	}

	d := memory.Get(mStart.Int64(), mSize.Int64())
	log := state.NewLog(context.Address(), topics, d, env.BlockNumber().Uint64())
	env.AddLog(log)
}

func opMload(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	offset := stack.pop()
	val := common.BigD(memory.Get(offset.Int64(), 32))
	stack.push(val)
}

func opMstore(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	// pop value of the stack
	mStart, val := stack.pop(), stack.pop()
	memory.Set(mStart.Uint64(), 32, common.BigToBytes(val, 256))
}

func opMstore8(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	off, val := stack.pop().Int64(), stack.pop().Int64()
	memory.store[off] = byte(val & 0xff)
}

func opSload(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	loc := common.BigToHash(stack.pop())
	val := env.State().GetState(context.Address(), loc).Big()
	stack.push(val)
}

func opSstore(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	loc := common.BigToHash(stack.pop())
	val := stack.pop()

	env.State().SetState(context.Address(), loc, common.BigToHash(val))
}

func opJump(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
}
func opJumpi(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
}
func opJumpdest(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
}

func opPc(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(data)
}

func opMsize(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(big.NewInt(int64(memory.Len())))
}

func opGas(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	stack.push(new(big.Int).Set(context.Gas))
}

func opCreate(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	var (
		value        = stack.pop()
		offset, size = stack.pop(), stack.pop()
		input        = memory.Get(offset.Int64(), size.Int64())
		gas          = new(big.Int).Set(context.Gas)
		addr         common.Address
	)

	context.UseGas(context.Gas)
	ret, suberr, ref := env.Create(context, input, gas, context.Price, value)
	if suberr != nil {
		stack.push(common.BigFalse)

	} else {
		// gas < len(ret) * CreateDataGas == NO_CODE
		dataGas := big.NewInt(int64(len(ret)))
		dataGas.Mul(dataGas, params.CreateDataGas)
		if context.UseGas(dataGas) {
			ref.SetCode(ret)
		}
		addr = ref.Address()

		stack.push(addr.Big())

	}
}

func opCall(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	gas := stack.pop()
	// pop gas and value of the stack.
	addr, value := stack.pop(), stack.pop()
	value = U256(value)
	// pop input size and offset
	inOffset, inSize := stack.pop(), stack.pop()
	// pop return size and offset
	retOffset, retSize := stack.pop(), stack.pop()

	address := common.BigToAddress(addr)

	// Get the arguments from the memory
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if len(value.Bytes()) > 0 {
		gas.Add(gas, params.CallStipend)
	}

	ret, err := env.Call(context, address, args, gas, context.Price, value)

	if err != nil {
		stack.push(common.BigFalse)

	} else {
		stack.push(common.BigTrue)

		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
}

func opCallCode(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	gas := stack.pop()
	// pop gas and value of the stack.
	addr, value := stack.pop(), stack.pop()
	value = U256(value)
	// pop input size and offset
	inOffset, inSize := stack.pop(), stack.pop()
	// pop return size and offset
	retOffset, retSize := stack.pop(), stack.pop()

	address := common.BigToAddress(addr)

	// Get the arguments from the memory
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if len(value.Bytes()) > 0 {
		gas.Add(gas, params.CallStipend)
	}

	ret, err := env.CallCode(context, address, args, gas, context.Price, value)

	if err != nil {
		stack.push(common.BigFalse)

	} else {
		stack.push(common.BigTrue)

		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
}

func opReturn(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {}
func opStop(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack)   {}

func opSuicide(data *big.Int, env Environment, context *Context, memory *Memory, stack *stack) {
	receiver := env.State().GetOrNewStateObject(common.BigToAddress(stack.pop()))
	balance := env.State().GetBalance(context.Address())

	receiver.AddBalance(balance)

	env.State().Delete(context.Address())
}
