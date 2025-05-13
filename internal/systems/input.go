package systems

import (
	"math"

	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type InputSystem struct {
	world *ecs.World
}

func NewInputSystem(world *ecs.World) *InputSystem {
	return &InputSystem{
		world: world,
	}
}

func (is *InputSystem) Update(dt float64) {
	if !rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		return
	}

	for _, t := range velPosFilter.Query(is.world) {
		pos := t[0].(*components.Position)
		vel := t[1].(*components.Velocity)

		mouseVector := components.Vector2{
			X: float64(rl.GetMouseX()),
			Y: float64(rl.GetMouseY()),
		}

		dir := components.Vector2{
			X: mouseVector.X - pos.X,
			Y: mouseVector.Y - pos.Y,
		}

		length := math.Hypot(dir.X, dir.Y)
		if length != 0 {
			dir.X /= length
			dir.Y /= length
		}

		speed := 100.0
		vel.X = dir.X * speed
		vel.Y = dir.Y * speed
	}
}
