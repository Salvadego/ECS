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
	entityCount  = 50_000
	screenWidth  = 800
	screenHeight = 600
)

var (
	currWidth, currHeight = screenWidth, screenHeight
	lastWidth, lastHeight = screenWidth, screenHeight
)

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(screenWidth, screenHeight, "ECS")
	rl.SetTargetFPS(120)

	world := ecs.NewWorld()
	movementSystem := systems.NewMovementSystem(world, screenWidth, screenHeight)
	renderSystem := systems.NewRenderSystem(world, screenWidth, screenHeight)
	inputSystem := systems.NewInputSystem(world)
	world.AddSystems(movementSystem, renderSystem, inputSystem)

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

	var fps int32 = 120
	var lastFps int32 = fps

	for !rl.WindowShouldClose() {
		currWidth, currHeight = rl.GetScreenWidth(), rl.GetScreenHeight()
		if currWidth != lastWidth || currHeight != lastHeight {
			movementSystem.SetSize(currWidth, currHeight)
			renderSystem.SetSize(currWidth, currHeight)
		}
		lastWidth, lastHeight = currWidth, currHeight

		fps = rl.GetFPS()
		if lastFps != fps {
			fmt.Print("\033[H\033[2J")
			fmt.Println(fps)
		}
		lastFps = fps

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		world.Update(float64(rl.GetFrameTime()))
		rl.DrawFPS(10, 10)
		rl.EndDrawing()
	}

	rl.CloseWindow()
}
