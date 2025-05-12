package systems

import (
	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
)

type MovementSystem struct {
	World                     *ecs.World
	screenWidth, screenHeight int
}

var velPosFilter = ecs.With[components.Position](ecs.With[components.Velocity](ecs.NewFilter()))

func NewMovementSystem(world *ecs.World, width, height int) *MovementSystem {
	return &MovementSystem{
		World:        world,
		screenWidth:  width,
		screenHeight: height,
	}
}

func (ms *MovementSystem) SetSize(width, height int){
	ms.screenWidth, ms.screenHeight = width, height
}

func (ms *MovementSystem) Update(dt float64) {
	for _, t := range ecs.Query2[*components.Velocity, *components.Position](velPosFilter, ms.World) {
		vel := t.C1
		pos := t.C2

		pos.X += vel.X * dt
		pos.Y += vel.Y * dt
		if pos.X < 0 || pos.X > float64(ms.screenWidth) {
			vel.X *= -1
		}

		if pos.Y < 0 || pos.Y > float64(ms.screenHeight) {
			vel.Y *= -1
		}
	}
}
