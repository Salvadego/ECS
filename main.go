package main

import (
	"flag"
	"fmt"
	"math/rand"

	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/internal/systems"
	"github.com/Salvadego/ECS/pkg/ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 600
	screenHeight = 450
)

var (
	currWidth, currHeight = screenWidth, screenHeight
	lastWidth, lastHeight = screenWidth, screenHeight
)

func main() {
	entityCount := flag.Int64("n", 10000, "Entity count")
	flag.Parse()

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(screenWidth, screenHeight, "ECS")
	rl.SetTargetFPS(120)

	world := ecs.NewWorld()
	movementSystem := systems.NewMovementSystem(world, screenWidth, screenHeight)
	renderSystem := systems.NewRenderSystem(world, screenWidth, screenHeight)
	inputSystem := systems.NewInputSystem(world)
	world.AddSystems(movementSystem, renderSystem, inputSystem)

	for range *entityCount {
		world.CreateEntity(
			&components.Position{
				X: float64(rand.Intn(screenWidth)),
				Y: float64(rand.Intn(screenHeight)),
			},
			&components.Velocity{
				X: (rand.Float64()*10 - 1) * 10,
				Y: (rand.Float64()*10 - 1) * 10,
			},
			components.Renderable{
				// Width: 20,
				// Height: 20,
				Color: rl.Color{
					R: 100,
					G: 255,
					B: 100,
					A: uint8(rand.Intn(100) + 100),
				},
			},
		)
	}

	for !rl.WindowShouldClose() {
		currWidth, currHeight = rl.GetScreenWidth(), rl.GetScreenHeight()
		if currWidth != lastWidth || currHeight != lastHeight {
			movementSystem.SetSize(currWidth, currHeight)
			renderSystem.SetSize(currWidth, currHeight)
		}
		lastWidth, lastHeight = currWidth, currHeight

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		world.Update(float64(rl.GetFrameTime()))

		rl.DrawFPS(10, 10)
		rl.DrawText(fmt.Sprintf("Entity count: %d", *entityCount), 10, 30, 20, rl.White)

		rl.EndDrawing()
	}

	rl.CloseWindow()
}
