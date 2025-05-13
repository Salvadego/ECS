package components

import "github.com/Salvadego/ECS/pkg/ecs"

const (
	PositionID ecs.ComponentID = 1 << iota
	RenderableID
	Vector2ID
	VelocityID
)

func init() {
	ecs.RegisterComponentType[*Position](PositionID)
	ecs.RegisterComponentType[*Velocity](VelocityID)
	ecs.RegisterComponentType[Renderable](RenderableID)
}
