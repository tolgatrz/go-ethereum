package vm

import (
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
)

type progStatus int32

const (
	progUnknown progStatus = iota
	progCompile
	progReady
	progError
)

var (
	programMu sync.RWMutex
	programs  = map[common.Hash]*Program{}
)

func GetProgram(addr common.Hash) *Program {
	programMu.RLock()
	defer programMu.RUnlock()
	return programs[addr]
}

func GetProgramStatus(addr common.Hash) progStatus {
	program := GetProgram(addr)
	if program != nil {
		return progStatus(atomic.LoadInt32(&program.status))
	}

	return progUnknown
}

func LinkProgram(addr common.Hash, program *Program) {
	programMu.Lock()
	defer programMu.Unlock()
	programs[addr] = program
}

// Program compiled program
type Program struct {
	status int32 // status should be accessed atomically

	context *Context

	instructions []instruction
	mapping      map[uint64]int
	destinations map[uint64]struct{}
}

func NewProgram() *Program {
	return &Program{mapping: make(map[uint64]int), destinations: make(map[uint64]struct{})}
}

func (p *Program) addInstr(op OpCode, pc uint64, fn instrFn, data *big.Int) {
	p.instructions = append(p.instructions, instruction{op, pc, fn, data})
	p.mapping[pc] = len(p.instructions) - 1
}

func AttachProgram(program *Program, code []byte) (err error) {
	if progStatus(atomic.LoadInt32(&program.status)) == progCompile {
		return nil
	}
	atomic.StoreInt32(&program.status, int32(progCompile))
	defer func() {
		if err != nil {
			atomic.StoreInt32(&program.status, int32(progError))
		} else {
			atomic.StoreInt32(&program.status, int32(progReady))
		}
	}()

	for pc := uint64(0); pc < uint64(len(code)); pc++ {
		switch op := OpCode(code[pc]); op {
		case ADD:
			program.addInstr(op, pc, opAdd, nil)
		case SUB:
			program.addInstr(op, pc, opSub, nil)
		case MUL:
			program.addInstr(op, pc, opMul, nil)
		case DIV:
			program.addInstr(op, pc, opDiv, nil)
		case SDIV:
			program.addInstr(op, pc, opSdiv, nil)
		case MOD:
			program.addInstr(op, pc, opMod, nil)
		case SMOD:
			program.addInstr(op, pc, opSmod, nil)
		case EXP:
			program.addInstr(op, pc, opExp, nil)
		case SIGNEXTEND:
			program.addInstr(op, pc, opSignExtend, nil)
		case NOT:
			program.addInstr(op, pc, opNot, nil)
		case LT:
			program.addInstr(op, pc, opLt, nil)
		case GT:
			program.addInstr(op, pc, opGt, nil)
		case SLT:
			program.addInstr(op, pc, opSlt, nil)
		case SGT:
			program.addInstr(op, pc, opSgt, nil)
		case EQ:
			program.addInstr(op, pc, opEq, nil)
		case ISZERO:
			program.addInstr(op, pc, opIszero, nil)
		case AND:
			program.addInstr(op, pc, opAnd, nil)
		case OR:
			program.addInstr(op, pc, opOr, nil)
		case XOR:
			program.addInstr(op, pc, opXor, nil)
		case BYTE:
			program.addInstr(op, pc, opByte, nil)
		case ADDMOD:
			program.addInstr(op, pc, opAddmod, nil)
		case MULMOD:
			program.addInstr(op, pc, opMulmod, nil)
		case SHA3:
			program.addInstr(op, pc, opSha3, nil)
		case ADDRESS:
			program.addInstr(op, pc, opAddress, nil)
		case BALANCE:
			program.addInstr(op, pc, opBalance, nil)
		case ORIGIN:
			program.addInstr(op, pc, opOrigin, nil)
		case CALLER:
			program.addInstr(op, pc, opCaller, nil)
		case CALLVALUE:
			program.addInstr(op, pc, opCallValue, nil)
		case CALLDATALOAD:
			program.addInstr(op, pc, opCalldataLoad, nil)
		case CALLDATASIZE:
			program.addInstr(op, pc, opCalldataSize, nil)
		case CALLDATACOPY:
			program.addInstr(op, pc, opCalldataCopy, nil)
		case CODESIZE:
			program.addInstr(op, pc, opCodeSize, nil)
		case EXTCODESIZE:
			program.addInstr(op, pc, opExtCodeSize, nil)
		case CODECOPY:
			program.addInstr(op, pc, opCodeCopy, nil)
		case EXTCODECOPY:
			program.addInstr(op, pc, opExtCodeCopy, nil)
		case GASPRICE:
			program.addInstr(op, pc, opGasprice, nil)
		case BLOCKHASH:
			program.addInstr(op, pc, opBlockhash, nil)
		case COINBASE:
			program.addInstr(op, pc, opCoinbase, nil)
		case TIMESTAMP:
			program.addInstr(op, pc, opTimestamp, nil)
		case NUMBER:
			program.addInstr(op, pc, opNumber, nil)
		case DIFFICULTY:
			program.addInstr(op, pc, opDifficulty, nil)
		case GASLIMIT:
			program.addInstr(op, pc, opGasLimit, nil)
		case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16, PUSH17, PUSH18, PUSH19, PUSH20, PUSH21, PUSH22, PUSH23, PUSH24, PUSH25, PUSH26, PUSH27, PUSH28, PUSH29, PUSH30, PUSH31, PUSH32:
			size := uint64(op - PUSH1 + 1)
			bytes := getData([]byte(code), new(big.Int).SetUint64(pc+1), new(big.Int).SetUint64(size))

			program.addInstr(op, pc, opPush, common.Bytes2Big(bytes))

			pc += size

		case POP:
			program.addInstr(op, pc, opPop, nil)
		case DUP1, DUP2, DUP3, DUP4, DUP5, DUP6, DUP7, DUP8, DUP9, DUP10, DUP11, DUP12, DUP13, DUP14, DUP15, DUP16:
			program.addInstr(op, pc, opDup, big.NewInt(int64(op-DUP1+1)))
		case SWAP1, SWAP2, SWAP3, SWAP4, SWAP5, SWAP6, SWAP7, SWAP8, SWAP9, SWAP10, SWAP11, SWAP12, SWAP13, SWAP14, SWAP15, SWAP16:
			program.addInstr(op, pc, opSwap, big.NewInt(int64(op-SWAP1+2)))
		case LOG0, LOG1, LOG2, LOG3, LOG4:
			program.addInstr(op, pc, opLog, big.NewInt(int64(op-LOG0)))
		case MLOAD:
			program.addInstr(op, pc, opMload, nil)
		case MSTORE:
			program.addInstr(op, pc, opMstore, nil)
		case MSTORE8:
			program.addInstr(op, pc, opMstore8, nil)
		case SLOAD:
			program.addInstr(op, pc, opSload, nil)
		case SSTORE:
			program.addInstr(op, pc, opSstore, nil)
		case JUMP:
			program.addInstr(op, pc, opJump, nil)
		case JUMPI:
			program.addInstr(op, pc, opJumpi, nil)
		case JUMPDEST:
			program.addInstr(op, pc, opJumpdest, nil)
			program.destinations[pc] = struct{}{}
		case PC:
			program.addInstr(op, pc, opPc, big.NewInt(int64(pc)))
		case MSIZE:
			program.addInstr(op, pc, opMsize, nil)
		case GAS:
			program.addInstr(op, pc, opGas, nil)
		case CREATE:
			program.addInstr(op, pc, opCreate, nil)
		case CALL:
			program.addInstr(op, pc, opCall, nil)
		case CALLCODE:
			program.addInstr(op, pc, opCallCode, nil)
		case RETURN:
			program.addInstr(op, pc, opReturn, nil)
		case SUICIDE:
			program.addInstr(op, pc, opSuicide, nil)
		case STOP: // Stop the context
			program.addInstr(op, pc, opStop, nil)
		default:
			program.addInstr(op, pc, nil, nil)
		}
	}

	return nil
}

func RunProgram(program *Program, env Environment, context *Context, input []byte) ([]byte, error) {
	context.Input = input

	var (
		caller      = context.caller
		statedb     = env.State()
		mem         = NewMemory()
		stack       = newstack()
		pc      int = 0

		jump = func(to *big.Int) error {
			if !validDest(program.destinations, to) {
				nop := context.GetOp(to.Uint64())
				return fmt.Errorf("invalid jump destination (%v) %v", nop, to)
			}

			pc = program.mapping[to.Uint64()]

			return nil
		}
	)

	iterations := 0
	//defer func() { fmt.Println("instructions", iterations) }()
	for pc < len(program.instructions) {
		iterations++

		instr := program.instructions[pc]

		// calculate the new memory size and gas price for the current executing opcode
		newMemSize, cost, err := calculateGasAndSize(env, context, caller, instr.op, statedb, mem, stack)
		if err != nil {
			return nil, err
		}

		// Use the calculated gas. When insufficient gas is present, use all gas and return an
		// Out Of Gas error
		if !context.UseGas(cost) {
			return nil, OutOfGasError
		}
		// Resize the memory calculated previously
		mem.Resize(newMemSize.Uint64())

		switch instr.op {
		case JUMP:
			if err := jump(stack.pop()); err != nil {
				return nil, err
			}
			continue
		case JUMPI:
			pos, cond := stack.pop(), stack.pop()

			if cond.Cmp(common.BigTrue) >= 0 {
				if err := jump(pos); err != nil {
					return nil, err
				}
				continue
			}
		case RETURN:
			offset, size := stack.pop(), stack.pop()
			ret := mem.GetPtr(offset.Int64(), size.Int64())

			return context.Return(ret), nil
		case SUICIDE:
			instr.fn(instr.data, env, context, mem, stack)

			return context.Return(nil), nil
		case STOP:
			return context.Return(nil), nil
		default:
			if instr.fn == nil {
				return nil, fmt.Errorf("Invalid opcode %x", instr.op)
			}

			instr.fn(instr.data, env, context, mem, stack)
		}

		pc++
	}

	return context.Return(nil), nil
}

func validDest(dests map[uint64]struct{}, dest *big.Int) bool {
	// PC cannot go beyond len(code) and certainly can't be bigger than 64bits.
	// Don't bother checking for JUMPDEST in that case.
	if dest.Cmp(bigMaxUint64) > 0 {
		return false
	}
	_, ok := dests[dest.Uint64()]
	return ok
}
