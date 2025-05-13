package systems

import (
	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
)

var (
	velPosFilter  = ecs.NewFilter(components.PositionID, components.VelocityID)
	posRendFilter = ecs.NewFilter(components.PositionID, components.RenderableID)
)
