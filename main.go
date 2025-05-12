package main

import (
	"fmt"
	"math/rand"

	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/internal/systems"
	"github.com/Salvadego/ECS/pkg/ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	entityCount  = 100_000
	screenWidth  = 800
	screenHeight = 600
)

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(screenWidth, screenHeight, "ECS")
	rl.SetTargetFPS(120)

	world := ecs.NewWorld()
	movementSystem := systems.NewMovementSystem(world, screenWidth, screenHeight)
	renderSystem := systems.NewRenderSystem(world, screenWidth, screenHeight)
	world.AddSystems(movementSystem, renderSystem)

	for range entityCount {
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
				Color: rl.Color{
					R: 100,
					G: 255,
					B: 100,
					A: uint8(rand.Intn(100)),
				},
			},
		)
	}

	lastWidth, lastHeight := screenWidth, screenHeight
	for !rl.WindowShouldClose() {
		currWidth, currHeight := rl.GetScreenWidth(), rl.GetScreenHeight()
		if currWidth != lastWidth || currHeight != lastHeight {
			movementSystem.SetSize(currWidth, currHeight)
			renderSystem.SetSize(currWidth, currHeight)
		}
		lastWidth, lastHeight = currWidth, currHeight

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		world.Update(float64(rl.GetFrameTime()))
		rl.DrawFPS(10, 10)
		rl.DrawText(fmt.Sprintf("Entity Count %d", entityCount), 10, 30, 30, rl.White)
		rl.EndDrawing()
	}

	rl.CloseWindow()
}
