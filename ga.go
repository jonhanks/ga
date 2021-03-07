package ga

import (
	"math/rand"
	"sort"
)

// An Individual is the basic genome/sample/... that is being evolved
type Individual interface {
	Clone() Individual
	Mutate() Individual
}

// Generators create new Individuals.
type Generator interface {
	Generate() Individual
	Evolve(a, b Individual) Individual
}

type Fitness func(Individual) float32

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
	individuals []gradedIndividual
	generator   Generator
	fitness     Fitness
}

// Create a new population
func NewPopulation(generator Generator, size int, f Fitness) *Population {
	individuals := make([]gradedIndividual, size, size)
	for i := 0; i < size; i++ {
		individuals[i] = gradedIndividual{individual: generator.Generate()}
	}
	return &Population{individuals: individuals, generator: generator, fitness: f}
}

// Test the population against the fitness function
func (p *Population) Test() {
	for i := range p.individuals {
		p.individuals[i].fitness = p.fitness(p.individuals[i].individual)
	}
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
	i := 0
	for ; i < copyCount; i++ {
		children[i].individual = p.individuals[i].individual.Clone()
	}
	remaining -= copyCount
	// generate 10% new population
	for ; remaining > 0 && copyCount > 0; i++ {
		remaining--
		copyCount--

		children[i].individual = p.generator.Generate()
	}
	// mutate or evolve the rest
	mutate := true
	for ; remaining > 0; i++ {
		remaining--

		pIndex := rand.Intn(len(p.individuals))
		if mutate {
			children[i].individual = p.individuals[pIndex].individual.Mutate()
		} else {
			pIndex2 := pIndex
			for pIndex2 == pIndex {
				pIndex = rand.Intn(len(p.individuals))
			}
			children[i].individual = p.generator.Evolve(p.individuals[pIndex].individual, p.individuals[pIndex2].individual)
		}
		mutate = !mutate
	}

	return &Population{
		individuals: children,
		generator:   p.generator,
		fitness:     p.fitness,
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
