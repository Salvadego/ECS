package components

import "github.com/Salvadego/ECS/pkg/ecs"

type Position Vector2

func (c Position) ID() ecs.ComponentID {
	return PositionID
}
