package main

import (
	"fmt"
	"github.com/jonhanks/ga"
	"github.com/jonhanks/ga/svm"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Generator struct {
}

func generateOp() svm.OpCode {
	return svm.OpCode{Code: byte(rand.Intn(svm.OpAbort + 1)),
		Literal: rand.Int31n(100)}
}

func (g *Generator) Generate() ga.Individual {
	l := 5 + rand.Intn(20)
	ops := make([]svm.OpCode, l, l)
	for i := 0; i < l; i++ {
		ops[i] = generateOp()
	}
	return &Individual{ops: ops}
}

func (g *Generator) Evolve(a, b ga.Individual) ga.Individual {
	_ = b
	ind := a.(*Individual)
	return ind.Mutate()
}

type Individual struct {
	ops []svm.OpCode
}

func (i Individual) String() string {
	l := len(i.ops)
	//if l > 5 {
	//	l = 5
	//}
	ops := make([]string, l, l)
	for idx := 0; idx < l; idx++ {
		ops[idx] = i.ops[idx].String()
	}
	return strings.Join(ops, "\n")
}

func (i *Individual) Clone() ga.Individual {
	ops := make([]svm.OpCode, len(i.ops), len(i.ops))
	copy(ops, i.ops)
	return &Individual{ops: ops}
}

func (i *Individual) Mutate() ga.Individual {
	action := rand.Intn(3)
	var ops []svm.OpCode
	l := len(i.ops)
	if l <= 2 && action == 1 {
		action = 2
	}
	switch action {
	case 0:
		ops = make([]svm.OpCode, l, l)
		copy(ops, i.ops)
		idx := rand.Intn(l)
		ops[idx] = generateOp()
	case 1:
		ops = make([]svm.OpCode, 0, l-1)
		splitPoint := rand.Intn(l)
		ops = append(ops, i.ops[0:splitPoint]...)
		ops = append(ops, i.ops[splitPoint+1:]...)
	case 2:
		ops = make([]svm.OpCode, 0, l+1)
		ops = append(ops, i.ops...)
		ops = append(ops, generateOp())
	}
	return &Individual{ops: ops}
}

func getVal() int32 {
	for {
		val := int32(rand.Intn(10000))
		if val != 0 {
			return val
		}
	}
}

func Fitness() ga.Fitness {
	vm := svm.NewSVM(100, 100)

	return func(individual ga.Individual) float32 {
		ind := individual.(*Individual)
		vm.ResetState()
		a := getVal()
		b := getVal()
		c := getVal()
		expected := float32(a + c)
		vm.PokeMem(0, a)
		vm.PokeMem(1, b)
		vm.PokeMem(2, c)
		exitType := vm.ExecuteProgram(ind.ops, 25)
		val := float32(vm.PeekMem(33))
		var modifier float32
		if exitType == svm.ExitOnAbort {
			modifier = 1.0
			l := len(ind.ops)
			switch {
			//case l < 6:
			//	modifier += 1.0
			case l < 8:
				modifier += 1.0
			case l < 15:
				modifier += 0.8
			case l < 20:
				modifier += 0.7
			case l < 25:
				modifier += 0.6
			case l < 30:
				modifier += 0.5
			}
		}
		return modifier - float32(math.Abs(float64(expected-val)))
	}
}

func main() {
	const matchLevel = 100
	seed := time.Now().Unix()
	_ = seed
	rand.Seed(seed)

	population := ga.NewPopulation(&Generator{}, 10000, Fitness())
	needMatches := matchLevel

	var bestIndividual ga.Individual

	for generation := 0; generation < 10000 && needMatches > 0; generation++ {
		population.Test()

		fmt.Printf("Generation %v need %d more consecutive perfect scores\n", generation, needMatches)
		for i := 0; i < 5; i++ {
			fmt.Printf("\t%v\n", population.FitnessAt(i))
		}
		bestIndividual = population.Individual(0)
		if population.FitnessAt(0) == 2.0 {
			needMatches--
		} else {
			needMatches = matchLevel
		}
		population = population.Evolve()
	}
	fmt.Printf("%v\n", bestIndividual)
}
