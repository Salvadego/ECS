package ecs_test

import (
	"math/rand"
	"testing"

	"github.com/Salvadego/ECS/pkg/ecs"
)

// This file contains profiling-specific benchmarks to identify specific bottlenecks
// and performance characteristics of the ECS implementation

// Benchmark different archetype operations
// Create some basic component types for testing
type TestComp1 struct{}

func (t TestComp1) ID() ecs.ComponentID { return 10 }

type TestComp2 struct{}

func (t TestComp2) ID() ecs.ComponentID { return 11 }

type TestComp3 struct{}

func (t TestComp3) ID() ecs.ComponentID { return 12 }

type TestComp4 struct{}

func (t TestComp4) ID() ecs.ComponentID { return 13 }
func BenchmarkArchetypeOperations(b *testing.B) {

	// Benchmark finding archetypes
	b.Run("ArchetypeMatching", func(b *testing.B) {
		world := ecs.NewWorld()

		// Create some different archetype combinations
		for i := 0; i < 50; i++ {
			var comps []ecs.Component
			if i%2 == 0 {
				comps = append(comps, TestComp1{})
			}
			if i%3 == 0 {
				comps = append(comps, TestComp2{})
			}
			if i%5 == 0 {
				comps = append(comps, TestComp3{})
			}
			if i%7 == 0 {
				comps = append(comps, TestComp4{})
			}
			// Create at least one entity per archetype
			world.CreateEntity(comps...)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Test query that should match several archetypes
			filter := ecs.NewFilter(10, 11) // TestComp1 and TestComp2
			results := filter.Query(world)
			_ = results
		}
	})

	// Benchmark BitSet operations
	b.Run("BitSetOperations", func(b *testing.B) {
		a := ecs.BitSet{}
		bs := ecs.BitSet{}

		// Set some bits
		a.Set(1)
		a.Set(10)
		a.Set(30)
		a.Set(63)

		bs.Set(1)
		bs.Set(40)
		bs.Set(63)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Test all BitSet operations
			_ = a.ContainsAll(bs)
			_ = a.Intersects(bs)
			_ = a.Equals(bs)
			_ = a.Hash()
			_ = a.Indices()
		}
	})
}

// Benchmark entity operations (creation, deletion if implemented)
// Define some small, medium, and large components
type SmallComponent struct {
	X, Y float32
}

func (s SmallComponent) ID() ecs.ComponentID { return 20 }

type MediumComponent struct {
	A, B, C, D, E, F, G, H float64
}

func (m MediumComponent) ID() ecs.ComponentID { return 21 }

type LargeComponent struct {
	Data [100]float32
}

func (l LargeComponent) ID() ecs.ComponentID { return 22 }
func BenchmarkEntityOperations(b *testing.B) {

	// Benchmark creation with different component sizes
	b.Run("CreateWithSmallComponent", func(b *testing.B) {
		world := ecs.NewWorld()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			world.CreateEntity(SmallComponent{float32(i), float32(i)})
		}
	})

	b.Run("CreateWithMediumComponent", func(b *testing.B) {
		world := ecs.NewWorld()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			world.CreateEntity(MediumComponent{
				float64(i), float64(i), float64(i), float64(i),
				float64(i), float64(i), float64(i), float64(i),
			})
		}
	})

	b.Run("CreateWithLargeComponent", func(b *testing.B) {
		world := ecs.NewWorld()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			var data LargeComponent
			world.CreateEntity(data)
		}
	})
}

// Benchmark specific performance bottlenecks
// Setup some component types
type Position struct {
	X, Y, Z float64
}

func (p Position) ID() ecs.ComponentID { return 30 }

type Velocity struct {
	X, Y, Z float64
}

func (v Velocity) ID() ecs.ComponentID { return 31 }
func BenchmarkBottlenecks(b *testing.B) {

	// Benchmark pure component access overhead
	b.Run("ComponentAccessOverhead", func(b *testing.B) {
		world := ecs.NewWorld()

		// Create a single entity with components
		entityID := world.CreateEntity(
			Position{1, 2, 3},
			Velocity{4, 5, 6},
		)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Access the component repeatedly to measure overhead
			pos := ecs.GetComponent[Position](world, entityID)
			_ = pos
		}
	})

	// Benchmark query with type assertion overhead
	b.Run("QueryTypeAssertionOverhead", func(b *testing.B) {
		world := ecs.NewWorld()

		// Create entities with the same component types
		for i := 0; i < 1000; i++ {
			world.CreateEntity(
				Position{float64(i), float64(i), float64(i)},
				Velocity{float64(i), float64(i), float64(i)},
			)
		}

		filter := ecs.NewFilter(30, 31) // Position and Velocity

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			results := filter.Query(world)

			// Force type assertions
			for _, row := range results {
				pos := row[0].(Position)
				vel := row[1].(Velocity)
				_ = pos
				_ = vel
			}
		}
	})
}

// Benchmark performance scalability
// Define a simple component
type TestComponent struct {
	Value int
}

func (t TestComponent) ID() ecs.ComponentID { return 40 }

// Benchmark memory locality effects
// Define component types
type ScalarComponent struct {
	Value float64
}

func (s ScalarComponent) ID() ecs.ComponentID { return 60 }

type ArrayComponent struct {
	Values [16]float64
}

func (a ArrayComponent) ID() ecs.ComponentID { return 61 }
func BenchmarkMemoryLocality(b *testing.B) {

	// Test sequential access vs. random access
	sequentialTest := func(b *testing.B, componentType string) {
		world := ecs.NewWorld()
		entityCount := 10000

		// Create entities based on component type
		var entities []ecs.EntityID
		if componentType == "scalar" {
			for i := 0; i < entityCount; i++ {
				entityID := world.CreateEntity(ScalarComponent{float64(i)})
				entities = append(entities, entityID)
			}
		} else {
			for i := 0; i < entityCount; i++ {
				var values [16]float64
				for j := 0; j < 16; j++ {
					values[j] = float64(i * j)
				}
				entityID := world.CreateEntity(ArrayComponent{values})
				entities = append(entities, entityID)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Sequential access
			var sum float64
			for _, entityID := range entities {
				if componentType == "scalar" {
					comp := ecs.GetComponent[ScalarComponent](world, entityID)
					sum += comp.Value
				} else {
					comp := ecs.GetComponent[ArrayComponent](world, entityID)
					for j := 0; j < 16; j++ {
						sum += comp.Values[j]
					}
				}
			}
			// Use sum to prevent optimization
			_ = sum
		}
	}

	randomTest := func(b *testing.B, componentType string) {
		world := ecs.NewWorld()
		entityCount := 10000

		// Create entities based on component type
		var entities []ecs.EntityID
		if componentType == "scalar" {
			for i := 0; i < entityCount; i++ {
				entityID := world.CreateEntity(ScalarComponent{float64(i)})
				entities = append(entities, entityID)
			}
		} else {
			for i := 0; i < entityCount; i++ {
				var values [16]float64
				for j := 0; j < 16; j++ {
					values[j] = float64(i * j)
				}
				entityID := world.CreateEntity(ArrayComponent{values})
				entities = append(entities, entityID)
			}
		}

		// Create a random access pattern
		accessPattern := make([]int, entityCount)
		for i := 0; i < entityCount; i++ {
			// Fisher-Yates shuffle
			j := i + int(rand.Intn(entityCount-i))
			accessPattern[i] = j
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Random access
			var sum float64
			for _, idx := range accessPattern {
				entityID := entities[idx]
				if componentType == "scalar" {
					comp := ecs.GetComponent[ScalarComponent](world, entityID)
					sum += comp.Value
				} else {
					comp := ecs.GetComponent[ArrayComponent](world, entityID)
					for j := 0; j < 16; j++ {
						sum += comp.Values[j]
					}
				}
			}
			// Use sum to prevent optimization
			_ = sum
		}
	}

	b.Run("SequentialAccess_Scalar", func(b *testing.B) {
		sequentialTest(b, "scalar")
	})

	b.Run("RandomAccess_Scalar", func(b *testing.B) {
		randomTest(b, "scalar")
	})

	b.Run("SequentialAccess_Array", func(b *testing.B) {
		sequentialTest(b, "array")
	})

	b.Run("RandomAccess_Array", func(b *testing.B) {
		randomTest(b, "array")
	})
}

// Helper function for dynamic components
type DynamicComponent struct {
	Value int
	id    ecs.ComponentID
}

func (d DynamicComponent) ID() ecs.ComponentID {
	return d.id
}
