package ga

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"
)

// An Individual is the basic genome/sample/... that is being evolved
type Individual interface {
	Clone() Individual
	Mutate(r *rand.Rand) Individual
}

// Generators create new Individuals.
type Generator interface {
	Generate() Individual
	Evolve(a, b Individual, r *rand.Rand) Individual
}

type Fitness func(Individual) float32

type FitnessGenerator interface {
	Generate() Fitness
}

type gradedIndividual struct {
	individual Individual
	fitness    float32
}

type gradedIndividualSort struct {
	individuals []gradedIndividual
}

func (gi *gradedIndividualSort) Len() int {
	return len(gi.individuals)
}

func (gi *gradedIndividualSort) Less(i, j int) bool {
	return gi.individuals[i].fitness > gi.individuals[j].fitness
}

func (gi *gradedIndividualSort) Swap(i, j int) {
	tmp := gi.individuals[i]
	gi.individuals[i] = gi.individuals[j]
	gi.individuals[j] = tmp
}

// The Population is a collection of Individuals that is evolved to meet the fitness function
type Population struct {
	individuals      []gradedIndividual
	generator        Generator
	fitnessGenerator FitnessGenerator
}

type singleFitness struct {
	f Fitness
}

func (sf *singleFitness) Generate() Fitness {
	return sf.f
}

// Create a new population
func NewPopulation(generator Generator, size int, f Fitness) *Population {
	return NewPopulationFG(generator, size, &singleFitness{f: f})
}

func NewPopulationFG(generator Generator, size int, fg FitnessGenerator) *Population {
	individuals := make([]gradedIndividual, size, size)
	for i := 0; i < size; i++ {
		individuals[i] = gradedIndividual{individual: generator.Generate()}
	}
	return &Population{individuals: individuals, generator: generator, fitnessGenerator: fg}
}

// Test the population against the fitness function
func (p *Population) Test() {
	cpus := runtime.NumCPU()
	wg := sync.WaitGroup{}
	stride := len(p.individuals) / cpus

	for curCpu := 0; curCpu < cpus; curCpu++ {
		wg.Add(1)
		go func(idx, stride int, f Fitness) {
			defer wg.Done()
			end := idx + stride
			if end > len(p.individuals) {
				end = len(p.individuals)
			}
			for ; idx < end; idx++ {
				p.individuals[idx].fitness = f(p.individuals[idx].individual)
			}
		}(curCpu*stride, stride, p.fitnessGenerator.Generate())
	}
	wg.Wait()
	sort.Sort(&gradedIndividualSort{p.individuals})
}

func (p *Population) Evolve() *Population {

	children := make([]gradedIndividual, len(p.individuals), len(p.individuals))

	remaining := len(p.individuals)
	// first the top 10% get copied over
	copyCount := int(float32(len(p.individuals)) * .1)
	if copyCount > remaining {
		copyCount = remaining
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(count int) {
		defer wg.Done()
		for i := 0; i < count; i++ {
			children[i].individual = p.individuals[i].individual.Clone()
		}
	}(copyCount)
	remaining -= copyCount
	i := copyCount
	// generate 10% new population
	newPopCount := copyCount
	if newPopCount > remaining {
		newPopCount = remaining
	}
	wg.Add(1)
	go func(start, count int) {
		defer wg.Done()
		for offset := 0; offset < count; offset++ {
			children[start+offset].individual = p.generator.Generate()
		}
	}(i, newPopCount)
	remaining -= newPopCount
	i += newPopCount
	// mutate or evolve the rest

	cpus := runtime.NumCPU()
	if cpus > remaining {
		cpus = remaining
	}
	stride := cpus
	for cpu := 0; cpu < cpus; cpu++ {
		wg.Add(1)
		go func(start, max, stride int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			pIndex := r.Intn(len(p.individuals))

			for cur := start; cur < max; cur += stride {
				mutate := (cur % 2) == 0
				if mutate {
					children[cur].individual = p.individuals[pIndex].individual.Mutate(r)
				} else {
					pIndex2 := pIndex
					for pIndex2 == pIndex {
						pIndex = r.Intn(len(p.individuals))
					}
					children[cur].individual = p.generator.Evolve(p.individuals[pIndex].individual, p.individuals[pIndex2].individual, r)
				}
			}
		}(i, len(children), stride)
		i++
	}

	wg.Wait()
	return &Population{
		individuals:      children,
		generator:        p.generator,
		fitnessGenerator: p.fitnessGenerator,
	}
}

func (p *Population) Length() int {
	return len(p.individuals)
}

func (p *Population) FitnessAt(i int) float32 {
	return p.individuals[i].fitness
}

func (p *Population) Individual(i int) Individual {
	return p.individuals[i].individual
}
