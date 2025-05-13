package components

import "github.com/Salvadego/ECS/pkg/ecs"

type Velocity Vector2

func (c Velocity) ID() ecs.ComponentID {
	return VelocityID
}
