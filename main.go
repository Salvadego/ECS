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
	screenWidth  = 800
	screenHeight = 600
	entityCount  = 20000
)

func main() {

	rl.InitWindow(screenWidth, screenHeight, "ECS Particle Simulation - 100K Entities")
	defer rl.CloseWindow()
	rl.SetTargetFPS(120)

	world := ecs.NewWorld()
	movementSystem := systems.NewMovementSystem(world, screenWidth, screenHeight)
	renderSystem := systems.NewRenderSystem(world, screenWidth, screenHeight)
	world.AddSystems(movementSystem, renderSystem)

	for range entityCount {
		e := world.NewEntity()

		pos := components.Position{
			X: float64(screenWidth / 2),
			Y: float64(screenHeight / 2),
		}

		vel := components.Velocity{
			DX: float64(rand.Float64()*100 - 50),
			DY: float64(rand.Float64()*100 - 50),
		}

		col := rl.Color{
			R: uint8(rand.Intn(256)),
			G: uint8(255),
			B: uint8(rand.Intn(256)),
			A: 255,
		}

		world.AddComponent(e, pos)
		world.AddComponent(e, vel)
		world.AddComponent(e, col)
	}

	lastScreenWid, lastScreenHei := screenWidth, screenHeight
	for !rl.WindowShouldClose() {

		currWidth, currHeight := rl.GetScreenWidth(), rl.GetScreenHeight()
		if currWidth != lastScreenWid || lastScreenHei != currHeight {
			renderSystem.SetSize(currWidth, currHeight)
			movementSystem.SetSize(currWidth, currHeight)
		}
		lastScreenWid, lastScreenHei = currWidth, currHeight

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		world.Update(float64(rl.GetFrameTime()))

		rl.DrawFPS(10, 10)
		rl.DrawText(fmt.Sprintf("Entities: %d", entityCount), 10, 30, 20, rl.White)

		rl.EndDrawing()
	}

}
