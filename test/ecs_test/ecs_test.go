package ecstest

import (
	"math/rand"
	"testing"

	"github.com/Salvadego/ECS/pkg/ecs"
)

type Position struct {
	X, Y, Z float64
}

func (p Position) ID() ecs.ComponentID { return 1 }

type Velocity struct {
	X, Y, Z float64
}

func (v Velocity) ID() ecs.ComponentID { return 2 }

type Health struct {
	Current, Max float64
}

func (h Health) ID() ecs.ComponentID { return 3 }

type Sprite struct {
	TextureID uint32
	Width     float64
	Height    float64
}

func (s Sprite) ID() ecs.ComponentID { return 4 }

type AI struct {
	State uint8
	Path  []Position
}

func (a AI) ID() ecs.ComponentID { return 5 }

type MovementSystem struct {
	world *ecs.World
}

func NewMovementSystem(world *ecs.World) *MovementSystem {
	return &MovementSystem{world: world}
}

func (s *MovementSystem) Update(dt float64) {
	filter := ecs.NewFilter(1, 2)
	results := filter.Query(s.world)

	for _, row := range results {
		pos := row[0].(Position)
		vel := row[1].(Velocity)

		pos.X += vel.X * dt
		pos.Y += vel.Y * dt
		pos.Z += vel.Z * dt

	}
}

func createEntities(_ *testing.B, world *ecs.World, count int, componentMix []float64) []ecs.EntityID {
	entities := make([]ecs.EntityID, count)

	for i := range count {
		var components []ecs.Component

		if rand.Float64() < componentMix[0] {
			components = append(components, Position{
				X: rand.Float64() * 100,
				Y: rand.Float64() * 100,
				Z: rand.Float64() * 100,
			})
		}

		if rand.Float64() < componentMix[1] {
			components = append(components, Velocity{
				X: (rand.Float64() - 0.5) * 10,
				Y: (rand.Float64() - 0.5) * 10,
				Z: (rand.Float64() - 0.5) * 10,
			})
		}

		if rand.Float64() < componentMix[2] {
			components = append(components, Health{
				Current: 100,
				Max:     100,
			})
		}

		if rand.Float64() < componentMix[3] {
			components = append(components, Sprite{
				TextureID: uint32(rand.Intn(100)),
				Width:     rand.Float64() * 2,
				Height:    rand.Float64() * 2,
			})
		}

		if rand.Float64() < componentMix[4] {
			path := make([]Position, 0, 10)
			for range 10 {
				path = append(path, Position{
					X: rand.Float64() * 100,
					Y: rand.Float64() * 100,
					Z: rand.Float64() * 100,
				})
			}

			components = append(components, AI{
				State: uint8(rand.Intn(5)),
				Path:  path,
			})
		}

		entities[i] = world.CreateEntity(components...)
	}

	return entities
}

// Benchmark entity creation
func BenchmarkEntityCreation(b *testing.B) {
	benchmarks := []struct {
		name         string
		entityCount  int
		componentMix []float64
	}{
		{"Small_FewComponents", 100, []float64{0.3, 0.3, 0.0, 0.0, 0.0}},
		{"Small_ManyComponents", 100, []float64{0.8, 0.8, 0.8, 0.8, 0.8}},
		{"Medium_FewComponents", 1000, []float64{0.3, 0.3, 0.0, 0.0, 0.0}},
		{"Medium_ManyComponents", 1000, []float64{0.8, 0.8, 0.8, 0.8, 0.8}},
		{"Large_FewComponents", 10000, []float64{0.3, 0.3, 0.0, 0.0, 0.0}},
		{"Large_ManyComponents", 10000, []float64{0.8, 0.8, 0.8, 0.8, 0.8}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Reset the timer for setup
			b.StopTimer()

			// Run the benchmark
			b.StartTimer()
			for b.Loop() {
				world := ecs.NewWorld()
				createEntities(b, world, bm.entityCount, bm.componentMix)
			}
			b.StopTimer()
		})
	}
}

// Benchmark component access
func BenchmarkComponentAccess(b *testing.B) {
	benchmarks := []struct {
		name        string
		entityCount int
	}{
		{"Small", 100},
		{"Medium", 1000},
		{"Large", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Setup
			b.StopTimer()
			world := ecs.NewWorld()
			entities := createEntities(b, world, bm.entityCount, []float64{1.0, 1.0, 1.0, 1.0, 1.0})

			// Run the benchmark
			b.StartTimer()
			for b.Loop() {
				// Access a component from each entity
				for _, ent := range entities {
					_ = ecs.GetComponent[Position](world, ent)
				}
			}
			b.StopTimer()
		})
	}
}

// Benchmark query performance
func BenchmarkQuery(b *testing.B) {
	benchmarks := []struct {
		name           string
		entityCount    int
		componentMix   []float64
		queryComponent []ecs.ComponentID
	}{
		{"Simple_Small", 100, []float64{0.7, 0.7, 0.7, 0.7, 0.7}, []ecs.ComponentID{1}},
		{"Simple_Medium", 1000, []float64{0.7, 0.7, 0.7, 0.7, 0.7}, []ecs.ComponentID{1}},
		{"Simple_Large", 10000, []float64{0.7, 0.7, 0.7, 0.7, 0.7}, []ecs.ComponentID{1}},
		{"Complex_Small", 100, []float64{0.7, 0.7, 0.7, 0.7, 0.7}, []ecs.ComponentID{1, 2, 3}},
		{"Complex_Medium", 1000, []float64{0.7, 0.7, 0.7, 0.7, 0.7}, []ecs.ComponentID{1, 2, 3}},
		{"Complex_Large", 10000, []float64{0.7, 0.7, 0.7, 0.7, 0.7}, []ecs.ComponentID{1, 2, 3}},
		{"Rare_Small", 100, []float64{0.7, 0.7, 0.1, 0.1, 0.1}, []ecs.ComponentID{3, 4, 5}},
		{"Rare_Medium", 1000, []float64{0.7, 0.7, 0.1, 0.1, 0.1}, []ecs.ComponentID{3, 4, 5}},
		{"Rare_Large", 10000, []float64{0.7, 0.7, 0.1, 0.1, 0.1}, []ecs.ComponentID{3, 4, 5}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Setup
			b.StopTimer()
			world := ecs.NewWorld()
			createEntities(b, world, bm.entityCount, bm.componentMix)
			filter := ecs.NewFilter(bm.queryComponent...)

			// Run the benchmark
			b.StartTimer()
			for b.Loop() {
				results := filter.Query(world)
				// Just accessing the results array ensures the compiler
				// doesn't optimize away the query
				if len(results) > 0 {
					_ = results[0]
				}
			}
			b.StopTimer()
		})
	}
}

// Benchmark system update
func BenchmarkSystemUpdate(b *testing.B) {
	benchmarks := []struct {
		name        string
		entityCount int
		systemCount int
	}{
		{"Small_OneSys", 100, 1},
		{"Medium_OneSys", 1000, 1},
		{"Large_OneSys", 10000, 1},
		{"Small_ManySys", 100, 10},
		{"Medium_ManySys", 1000, 10},
		{"Large_ManySys", 10000, 10},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Setup
			b.StopTimer()
			world := ecs.NewWorld()

			// Create entities with position and velocity
			componentMix := []float64{1.0, 1.0, 0.5, 0.5, 0.5}
			createEntities(b, world, bm.entityCount, componentMix)

			// Add systems
			for range bm.systemCount {
				world.AddSystems(NewMovementSystem(world))
			}

			// Run the benchmark
			b.StartTimer()
			for b.Loop() {
				world.Update(0.016) // ~60 FPS
			}
			b.StopTimer()
		})
	}
}

// Benchmark mixed operations (more realistic usage pattern)
func BenchmarkMixedOperations(b *testing.B) {
	benchmarks := []struct {
		name        string
		entityCount int
		createOps   int
		queryOps    int
		accessOps   int
	}{
		{"Small", 100, 10, 5, 20},
		{"Medium", 1000, 50, 10, 100},
		{"Large", 10000, 100, 20, 200},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Setup
			b.StopTimer()
			world := ecs.NewWorld()
			componentMix := []float64{0.7, 0.7, 0.7, 0.7, 0.7}
			entities := createEntities(b, world, bm.entityCount, componentMix)
			filter := ecs.NewFilter(1, 2)

			// Add a system
			world.AddSystems(NewMovementSystem(world))

			// Run the benchmark
			b.StartTimer()
			for b.Loop() {
				// Perform some entity creations
				newEntities := createEntities(b, world, bm.createOps, componentMix)
				entities = append(entities, newEntities...)

				// Perform some queries
				for range bm.queryOps {
					results := filter.Query(world)
					if len(results) > 0 {
						_ = results[0]
					}
				}

				// Perform some component accesses
				for range bm.accessOps {
					if len(entities) > 0 {
						idx := rand.Intn(len(entities))
						_ = ecs.GetComponent[Position](world, entities[idx])
					}
				}

				// Update systems
				world.Update(0.016)
			}
			b.StopTimer()
		})
	}
}

// Memory allocation benchmark
func BenchmarkMemoryUsage(b *testing.B) {
	benchmarks := []struct {
		name         string
		entityCount  int
		componentMix []float64
	}{
		{"Small_FewComponents", 100, []float64{0.3, 0.3, 0.0, 0.0, 0.0}},
		{"Small_ManyComponents", 100, []float64{1.0, 1.0, 1.0, 1.0, 1.0}},
		{"Medium_FewComponents", 1000, []float64{0.3, 0.3, 0.0, 0.0, 0.0}},
		{"Medium_ManyComponents", 1000, []float64{1.0, 1.0, 1.0, 1.0, 1.0}},
		{"Large_FewComponents", 10000, []float64{0.3, 0.3, 0.0, 0.0, 0.0}},
		{"Large_ManyComponents", 10000, []float64{1.0, 1.0, 1.0, 1.0, 1.0}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				world := ecs.NewWorld()
				createEntities(b, world, bm.entityCount, bm.componentMix)

				// Perform a typical query
				filter := ecs.NewFilter(1, 2)
				results := filter.Query(world)
				if len(results) > 0 {
					_ = results[0]
				}

				// Update the world
				world.Update(0.016)
			}
		})
	}
}

// Benchmark for mutex contention
func BenchmarkMutexContention(b *testing.B) {
	benchmarks := []struct {
		name           string
		entityCount    int
		goroutineCount int
	}{
		{"Low_Contention", 1000, 2},
		{"Medium_Contention", 1000, 4},
		{"High_Contention", 1000, 8},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Setup
			b.StopTimer()
			world := ecs.NewWorld()
			componentMix := []float64{0.7, 0.7, 0.7, 0.7, 0.7}
			entities := createEntities(b, world, bm.entityCount, componentMix)

			b.StartTimer()
			for b.Loop() {
				// Create a wait group for goroutines
				done := make(chan bool)

				// Launch goroutines that all try to access the world
				for g := range bm.goroutineCount {
					go func(id int) {
						if id%3 == 0 {
							// Create some entities
							createEntities(b, world, 10, componentMix)
						} else if id%3 == 1 {
							// Query some components
							filter := ecs.NewFilter(1, 2)
							results := filter.Query(world)
							if len(results) > 0 {
								_ = results[0]
							}
						} else {
							// Access some components
							for range 20 {
								if len(entities) > 0 {
									idx := rand.Intn(len(entities))
									_ = ecs.GetComponent[Position](world, entities[idx])
								}
							}
						}
						done <- true
					}(g)
				}

				// Wait for all goroutines to complete
				for range bm.goroutineCount {
					<-done
				}
			}
			b.StopTimer()
		})
	}
}

