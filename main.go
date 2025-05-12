package main

import (
	"math/rand"

	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	entityCount  = 200_000
	screenWidth  = 800
	screenHeight = 600
)

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(screenWidth, screenHeight, "ECS")
	rl.SetTargetFPS(120)

	world := ecs.NewWorld()

	for range entityCount {
		world.CreateEntity(
			&components.Position{X: float64(rand.Intn(screenWidth)), Y: float64(rand.Intn(screenHeight))},
			&components.Velocity{X: 0, Y: 0},
			components.Renderable{
				Width:  2,
				Height: 2,
				Color: rl.Color{
					R: 255,
					G: 255,
					B: 255,
					A: uint8(rand.Intn(100)),
				},
			},
		)
	}

	velPosFilter := ecs.With[components.Position](ecs.With[components.Velocity](ecs.NewFilter()))
	posRendFilter := ecs.With[components.Renderable](ecs.With[components.Position](ecs.NewFilter()))

	for !rl.WindowShouldClose() {
		for _, t := range ecs.Query2[*components.Velocity, *components.Position](velPosFilter, world) {
			vel := t.C1
			pos := t.C2

			vel.X = (rand.Float64()*2 - 1) * 2
			vel.Y = (rand.Float64()*2 - 1) * 2

			pos.X += vel.X
			pos.Y += vel.Y
			if pos.X < 0 {
				pos.X += screenWidth
			} else if pos.X > screenWidth {
				pos.X -= screenWidth
			}
			if pos.Y < 0 {
				pos.Y += screenHeight
			} else if pos.Y > screenHeight {
				pos.Y -= screenHeight
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.DrawFPS(10, 10)
		for _, t := range ecs.Query2[*components.Position, components.Renderable](posRendFilter, world) {
			pos := t.C1
			rend := t.C2
			rl.DrawRectangle(
				int32(pos.X), int32(pos.Y),
				int32(rend.Width), int32(rend.Height),
				rend.Color,
			)
		}

		rl.EndDrawing()
	}

	rl.CloseWindow()
}
