package components

import (
	"github.com/Salvadego/ECS/pkg/ecs"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Renderable struct {
	Color rl.Color
}

func (c Renderable) ID() ecs.ComponentID {
	return RenderableID
}
