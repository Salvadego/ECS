package components

import "github.com/Salvadego/ECS/pkg/ecs"

type Vector2 struct{ X, Y float64 }

func (c Vector2) ID() ecs.ComponentID {
	return Vector2ID
}
