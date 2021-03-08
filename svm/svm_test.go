package svm

import "testing"

func Test_NewSVM(t *testing.T) {
	vm := NewSVM(100, 100)
	if vm == nil {
		t.Fatal("Invalid VM")
	}
}

func Test_SVM_Peek(t *testing.T) {
	vm := NewSVM(100, 100)
	for i := 0; i < 100; i++ {
		if vm.PeekMem(i) != 0 {
			t.Fatalf("Unexpected memory value at %v", i)
		}
	}
	if vm.PeekMem(-1) != 0 {
		t.Fatalf("Out of bounds read did not result in a zero value")
	}
}

func Test_SVM_Poke(t *testing.T) {
	vm := NewSVM(100, 100)
	vm.PokeMem(6, 42)
	vm.PokeMem(42, 7)
	if vm.PeekMem(6) != 42 || vm.PeekMem(42) != 7 {
		t.Fatal("PokeMem did not set memory as expected")
	}
}
