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
	ecs.RegisterComponentType[*Renderable](RenderableID)
	ecs.RegisterComponentType[*Vector2](Vector2ID)
	ecs.RegisterComponentType[*Velocity](VelocityID)
}
