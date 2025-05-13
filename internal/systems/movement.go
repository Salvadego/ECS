package systems

import (
	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
)

type MovementSystem struct {
	world                     *ecs.World
	screenWidth, screenHeight int
}

func NewMovementSystem(world *ecs.World, width, height int) *MovementSystem {
	return &MovementSystem{
		world:        world,
		screenWidth:  width,
		screenHeight: height,
	}
}

func (ms *MovementSystem) SetSize(width, height int) {
	ms.screenWidth, ms.screenHeight = width, height
}

func (ms *MovementSystem) Update(dt float64) {
	for _, t := range velPosFilter.Query(ms.world) {
		pos := t[0].(*components.Position)
		vel := t[1].(*components.Velocity)

		pos.X += vel.X * dt
		pos.Y += vel.Y * dt
		if pos.X <= 0 || pos.X >= float64(ms.screenWidth) {
			vel.X *= -1
		}

		if pos.Y <= 0 || pos.Y >= float64(ms.screenHeight) {
			vel.Y *= -1
		}
	}
}
