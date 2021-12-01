package main

import (
	"flag"
	"fmt"
	"github.com/jonhanks/ga"
	"github.com/jonhanks/ga/svm"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"strings"
	"time"
)

type Generator struct {
	r *rand.Rand
}

func generateOp(r *rand.Rand) svm.OpCode {
	return svm.OpCode{Code: byte(r.Intn(svm.OpAbort + 1)),
		Literal: r.Int31n(5)}
}

func NewGenerator() *Generator {
	return &Generator{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (g *Generator) Generate() ga.Individual {
	l := 5 + g.r.Intn(20)
	ops := make([]svm.OpCode, l, l)
	for i := 0; i < l; i++ {
		ops[i] = generateOp(g.r)
	}
	return &Individual{ops: ops}
}

func (g *Generator) Evolve(a, b ga.Individual, r *rand.Rand) ga.Individual {
	_ = b
	ind := a.(*Individual)
	return ind.Mutate(r)
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

func (i *Individual) Mutate(r *rand.Rand) ga.Individual {
	action := r.Intn(3)
	var ops []svm.OpCode
	l := len(i.ops)
	if l <= 2 && action == 1 {
		action = 2
	}
	switch action {
	case 0:
		ops = make([]svm.OpCode, l, l)
		copy(ops, i.ops)
		idx := r.Intn(l)
		ops[idx] = generateOp(r)
	case 1:
		ops = make([]svm.OpCode, l-1, l-1)
		splitPoint := r.Intn(l)
		copy(ops[0:splitPoint], i.ops[0:splitPoint])
		copy(ops[splitPoint:], i.ops[splitPoint+1:])
		//ops = append(ops, i.ops[0:splitPoint]...)
		//ops = append(ops, i.ops[splitPoint+1:]...)
	case 2:
		ops = make([]svm.OpCode, l+1, l+1)
		copy(ops[0:l], i.ops)
		ops[l] = generateOp(r)
		//ops = append(ops, i.ops...)
		//ops = append(ops, generateOp(r))
	}
	return &Individual{ops: ops}
}

func getVal(r *rand.Rand) int32 {
	for {
		val := int32(r.Intn(10000))
		if val != 0 {
			return val
		}
	}
}

type FitnessGenerator struct {
}

func (fg *FitnessGenerator) Generate() ga.Fitness {
	vm := svm.NewSVM(100, 100)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(individual ga.Individual) float32 {
		ind := individual.(*Individual)
		vm.ResetState()
		a := getVal(r)
		b := getVal(r)
		c := getVal(r)
		expected := float32((a + c) * b)
		vm.PokeMem(0, a)
		vm.PokeMem(1, b)
		vm.PokeMem(2, c)
		exitType := vm.ExecuteProgram(ind.ops, 25)
		val := float32(vm.PeekMem(3))
		var modifier float32
		if exitType == svm.ExitOnAbort {
			modifier = 1.0
			l := len(ind.ops)
			switch {
			//case l < 6:
			//	modifier += 1.0
			case l < 10:
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
	cpuProfilePath := flag.String("cpu-profile", "", "path to cpu profile, empty is none")
	flag.Parse()
	const matchLevel = 100
	seed := time.Now().Unix()
	_ = seed
	rand.Seed(seed)

	if *cpuProfilePath != "" {
		f, err := os.Create(*cpuProfilePath)
		if err != nil {
			log.Fatal("could not create the cpu profile file, ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start the cpu profiler, ", err)
		}
		defer pprof.StopCPUProfile()
	}

	population := ga.NewPopulationFG(NewGenerator(), 4000000, &FitnessGenerator{})
	needMatches := matchLevel

	var bestIndividual ga.Individual

	for generation := 0; generation < 1000 && needMatches > 0; generation++ {
		population.Test()

		fmt.Printf("Generation %v need %d more consecutive perfect scores\n", generation, needMatches)
		for i := 0; i < 5; i++ {
			fmt.Printf("\t%v\n", population.FitnessAt(i))
		}
		bestIndividual = population.Individual(0)
		if population.FitnessAt(0) >= 0 {
			needMatches--
		} else {
			needMatches = matchLevel
		}
		population = population.Evolve()
	}
	fmt.Printf("%v\n", bestIndividual)
}
