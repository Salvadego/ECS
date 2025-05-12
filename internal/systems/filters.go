package systems

import (
	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
)

var (
	velPosFilter  = ecs.With[components.Position](ecs.With[components.Velocity](ecs.NewFilter()))
	posRendFilter = ecs.With[components.Renderable](ecs.With[components.Position](ecs.NewFilter()))
)
