package svm

import (
	"fmt"
	"math"
)

type stats struct {
	InstructionsIssued  int
	InvalidInstructions int
}

// A Simple VM
type SVM struct {
	memory []int32
	stack  []int32
	sp     uint
	stats  stats
}

const (
	OpNop    = iota
	OpBitOr  // 2 args on stack
	OpBitAnd // 2 args on stack
	OpBitXor // 2 arg on stack
	OpAdd    // 2 args on stack
	OpSub
	OpMult
	OpDiv
	OpPush          // 1 arg literal
	OpPop           // no args
	OpPushDuplicate // no args
	OpPushMem       // 1 arg literal
	OpPopMem        // 1 arg literal
	OpJumpRel       // 1 arg literal
	OpJumpEq        // 1 arg on stack 1 arg literal
	OpJumpGt
	OpJumpLt
	OpAbort // no args -- must be last opcode
)

type OpCode struct {
	Code    byte
	Literal int32
}

func (o OpCode) String() (out string) {
	switch o.Code {
	case OpNop:
		out = "nop"
	case OpBitOr:
		out = "bit_or"
	case OpBitAnd:
		out = "bit_and"
	case OpBitXor:
		out = "bit_xor"
	case OpAdd:
		out = "add"
	case OpSub:
		out = "sub"
	case OpMult:
		out = "mult"
	case OpDiv:
		out = "div"
	case OpPush:
		out = fmt.Sprintf("push %d", o.Literal)
	case OpPop:
		out = "pop"
	case OpPushDuplicate:
		out = "push_dup"
	case OpPushMem:
		out = fmt.Sprintf("push (%d)", o.Literal)
	case OpPopMem:
		out = fmt.Sprintf("pop_to %d", o.Literal)
	case OpJumpRel:
		out = fmt.Sprintf("jmp %d", o.Literal)
	case OpJumpEq:
		out = fmt.Sprintf("jmp_eq %d", o.Literal)
	case OpJumpLt:
		out = fmt.Sprintf("jmp_lt %d", o.Literal)
	case OpJumpGt:
		out = fmt.Sprintf("jmp_gt %d", o.Literal)
	case OpAbort:
		out = "abort"
	default:
		out = "unknown"
	}
	return
}

func NewSVM(words uint, stackSize uint) *SVM {
	return &SVM{memory: make([]int32, words, words),
		stack: make([]int32, stackSize, stackSize)}
}

func (vm *SVM) PeekMem(address int) int32 {
	if address < 0 || address >= len(vm.memory) {
		return 0
	}
	return vm.memory[address]
}

func (vm *SVM) PokeMem(address int, value int32) {
	if address >= 0 && address < len(vm.memory) {
		vm.memory[address] = value
	}
}

func (vm *SVM) popStack() (val int32) {
	val = 0
	if vm.sp > 0 {
		vm.sp--
		val = vm.stack[vm.sp]
	}
	return
}

func (vm *SVM) pushStack(val int32) {
	if vm.sp < uint(len(vm.stack)) {
		vm.stack[vm.sp] = val
		vm.sp++
	}
}

func (vm *SVM) ResetState() {
	for i := range vm.memory {
		vm.memory[i] = 0
	}
	vm.stats = stats{}
	vm.sp = 0
}

const (
	ExitOnTimeout = iota
	ExitOnAbort
)

type ExitType int

func (vm *SVM) ExecuteProgram(program []OpCode, maxSteps int) ExitType {
	ip := 0
	done := false
	var exitType ExitType

	boundIp := func() {
		if ip < 0 {
			ip = 0
		}
		if ip >= len(program) {
			ip = 0
		}
	}

	for !done && maxSteps > 0 {
		maxSteps--
		boundIp()

		curOp := program[ip]
		ip++
		vm.stats.InstructionsIssued++

		switch curOp.Code {
		case OpNop:
			break
		case OpBitOr:
			vm.pushStack(vm.popStack() | vm.popStack())
		case OpBitAnd:
			vm.pushStack(vm.popStack() & vm.popStack())
		case OpBitXor:
			vm.pushStack(vm.popStack() ^ vm.popStack())
		case OpAdd:
			vm.pushStack(vm.popStack() + vm.popStack())
		case OpSub:
			vm.pushStack(vm.popStack() - vm.popStack())
		case OpMult:
			vm.pushStack(vm.popStack() * vm.popStack())
		case OpDiv:
			{
				dividend := vm.popStack()
				divisor := vm.popStack()
				if divisor != 0 {
					vm.pushStack(dividend / divisor)
				} else {
					vm.pushStack(math.MaxInt32)
				}
			}
		case OpPush:
			vm.pushStack(curOp.Literal)
		case OpPop:
			vm.popStack()
		case OpPushDuplicate:
			{
				val := vm.popStack()
				vm.pushStack(val)
				vm.pushStack(val)
			}
		case OpPushMem:
			vm.pushStack(vm.PeekMem(int(curOp.Literal)))
		case OpPopMem:
			vm.PokeMem(int(curOp.Literal), vm.popStack())
		case OpJumpRel:
			ip += int(curOp.Literal)
		case OpJumpEq:
			if vm.popStack() == 0 {
				ip += int(curOp.Literal)
			}
		case OpJumpGt:
			if vm.popStack() > 0 {
				ip += int(curOp.Literal)
			}
		case OpJumpLt:
			if vm.popStack() < 0 {
				ip += int(curOp.Literal)
			}
		case OpAbort:
			done = true
			exitType = ExitOnAbort
		default:
			vm.stats.InvalidInstructions++
		}
	}
	return exitType
}
