package main

import (
	"flag"
	"fmt"
	"github.com/jonhanks/ga"
	"log"
	"math"
	"math/rand"
	"time"
)

type Generator struct {
}

func (sg *Generator) Generate() ga.Individual {
	ind := Individual{}
	for i := 0; i < len(ind.Genes); i++ {
		ind.Genes[i] = byte('a') + byte(rand.Intn(26))
	}
	return &ind
}

func (sg *Generator) Evolve(a, b ga.Individual) ga.Individual {
	aInd := a.(*Individual)
	bInd := b.(*Individual)
	newInd := &Individual{}

	splitPoint := rand.Intn(10)
	i := 0
	for ; i < splitPoint; i++ {
		newInd.Genes[i] = aInd.Genes[i]
	}
	for i := splitPoint; i < 10; i++ {
		newInd.Genes[i] = bInd.Genes[i]
	}
	return newInd
}

type Individual struct {
	Genes [10]byte
}

func (ind *Individual) Clone() ga.Individual {
	return &Individual{Genes: ind.Genes}
}

func (ind *Individual) Mutate() ga.Individual {
	i := rand.Intn(10)
	newGenes := ind.Genes
	newGenes[i] = byte('a') + byte(rand.Intn(26))
	return &Individual{Genes: newGenes}
}

func Fitness(phrase string) ga.Fitness {
	if len(phrase) != 10 {
		log.Fatal("The phrase must be 10 letters long and only lower case a-z")
	}
	var reference [10]byte
	for i, v := range []byte(phrase) {
		reference[i] = v
	}
	return func(individual ga.Individual) float32 {
		ind := individual.(*Individual)
		score := 0.
		for i, v := range reference {
			delta := int(v) - int(ind.Genes[i])
			score += 1.0 - math.Abs(float64(delta))/25.0
		}
		return float32(score)
	}
}

func main() {
	searchPhrase := flag.String("phrase", "helloworld", "Phrase to search for, must be 10 characters")
	maxGenerations := flag.Int("generations", 1000, "Max number of generations")
	populationSize := flag.Int("size", 1000, "Number of individuals in the population")
	flag.Parse()

	rand.Seed(time.Now().Unix())
	population := ga.NewPopulation(&Generator{}, *populationSize, Fitness(*searchPhrase))
	matchFound := false
	for generation := 0; generation < *maxGenerations && !matchFound; generation++ {
		population.Test()

		//max := population.Length()
		fmt.Printf("Generation %d\n", generation)
		for i := 0; i < 5; i++ {

			ind := population.Individual(i).(*Individual)
			fitness := population.FitnessAt(i)
			fmt.Printf("\t%s %v\n", string(ind.Genes[:]), fitness)
			if fitness == 10.0 {
				matchFound = true
			}
		}

		population = population.Evolve()
	}
}
