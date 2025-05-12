package entities

import (
	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
)

func Player(world *ecs.World) ecs.Entity {
	player := world.NewEntity()
	ecs.AddComponent(world, player, components.Position{X: 0, Y: 0})
	ecs.AddComponent(world, player, components.Velocity{DX: 0, DY: 0})
	return player
}
