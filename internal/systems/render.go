package systems

import (
	"image/color"

	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type RenderSystem struct {
	world                     *ecs.World
	screenWidth, screenHeight int
	texture                   rl.Texture2D
	framebuffer               []color.RGBA
}

func NewRenderSystem(world *ecs.World, width, height int) *RenderSystem {
	framebuffer := make([]color.RGBA, width*height)
	image := rl.GenImageColor(width, height, rl.Black)
	texture := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)

	return &RenderSystem{
		world:        world,
		screenWidth:  width,
		screenHeight: height,
		texture:      texture,
		framebuffer:  framebuffer,
	}
}

func (rs *RenderSystem) SetSize(width, height int) {
	rs.screenWidth, rs.screenHeight = width, height
	rs.framebuffer = make([]color.RGBA, width*height)
	image := rl.GenImageColor(width, height, rl.Black)
	rs.texture = rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
}

func (rs *RenderSystem) Update(_ float64) {
	for i := range rs.framebuffer {
		rs.framebuffer[i] = color.RGBA{0, 0, 0, 255}
	}

	for _, t := range posRendFilter.Query(rs.world) {
		pos := t[0].(*components.Position)
		rend := t[1].(components.Renderable)

		px := int(pos.X)
		py := int(pos.Y)
		if px >= 0 && px < rs.screenWidth && py >= 0 && py < rs.screenHeight {
			i := py*rs.screenWidth + px
			rs.framebuffer[i] = color.RGBA{
				R: rend.Color.R,
				G: rend.Color.G,
				B: rend.Color.B,
				A: rend.Color.A,
			}
		}
	}

	rl.UpdateTexture(rs.texture, rs.framebuffer)
	rl.DrawTexture(rs.texture, 0, 0, rl.White)
}
