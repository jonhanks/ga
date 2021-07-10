package ga

import (
	"math/rand"
	"testing"
)

type dummyIndividual struct {
	value float32
}

func (d *dummyIndividual) Clone() Individual {
	return &dummyIndividual{value: d.value}
}

func (d *dummyIndividual) Mutate(r *rand.Rand) Individual {
	newVal := d.value
	for newVal == d.value {
		newVal = r.Float32()
	}
	return &dummyIndividual{value: newVal}
}

type dummyGenerator struct {
}

func (d dummyGenerator) Generate() Individual {
	return &dummyIndividual{value: rand.Float32()}
}

func (d dummyGenerator) Evolve(a, b Individual, r *rand.Rand) Individual {
	parentA := a.(*dummyIndividual)
	parentB := b.(*dummyIndividual)
	return &dummyIndividual{value: parentA.value + parentB.value}
}

func dummyFitness(individual Individual) float32 {
	return individual.(*dummyIndividual).value
}

func TestNewPopulation(t *testing.T) {

	type args struct {
		generator Generator
		size      int
		f         Fitness
	}
	tests := []struct {
		name string
		args args
		want *Population
	}{
		{
			name: "empty",
			args: args{
				generator: &dummyGenerator{},
				size:      100,
				f:         dummyFitness,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPopulation(tt.args.generator, tt.args.size, tt.args.f)
			if got.Length() != tt.args.size {
				t.Fatalf("Wrong size for the population")
			}
			got.Test()
			score := got.FitnessAt(0)
			for i := 1; i < tt.args.size; i++ {
				if got.FitnessAt(i) > score {
					t.Fatalf("fitness scores not sorted properly")
				}
				score = got.FitnessAt(i)
			}
		})
	}
}
