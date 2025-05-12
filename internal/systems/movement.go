package systems

import (
	"fmt"

	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
)

type MovementSystem struct {
	world        *ecs.World
	moveFilter   ecs.Filter
	screenWidth  int32
	screenHeight int32
}

func NewMovementSystem(world *ecs.World, width, height int32) *MovementSystem {
	moveFilter := world.CreateFilter(components.Position{}, components.Velocity{})
	return &MovementSystem{
		world:        world,
		moveFilter:   moveFilter,
		screenWidth:  width,
		screenHeight: height,
	}
}

func (s *MovementSystem) SetSize(wid, hei int) {
	fmt.Println("Updated")
	s.screenWidth = int32(wid)
	s.screenHeight = int32(hei)
}

func (s *MovementSystem) Update(dt float64) {
	entities := s.world.Query(s.moveFilter)

	for _, entity := range entities {
		pos, _ := ecs.GetComponent[components.Position](s.world, entity)
		vel, _ := ecs.GetComponent[components.Velocity](s.world, entity)

		newX := pos.X + vel.DX*dt
		newY := pos.Y + vel.DY*dt

		bounceX := !isWithinBounds(newX, float64(s.screenWidth))
		bounceY := !isWithinBounds(newY, float64(s.screenHeight))

		ecs.UpdateComponent(s.world, entity, func(p *components.Position) {
			p.X = newX
			p.Y = newY
		})

		if bounceX || bounceY {
			ecs.UpdateComponent(s.world, entity, func(v *components.Velocity) {
				if bounceX {
					v.DX *= -1
				}
				if bounceY {
					v.DY *= -1
				}
			})
		}
	}
}

func isWithinBounds(pos float64, boundary float64) bool {
	return pos >= 0 && pos <= boundary
}
