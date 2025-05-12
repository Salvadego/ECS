package systems

import (
	"image/color"

	"github.com/Salvadego/ECS/internal/components"
	"github.com/Salvadego/ECS/pkg/ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type RenderSystem struct {
	world         *ecs.World
	renderFilter  ecs.Filter
	imageBuffer   []color.RGBA
	texture       rl.Texture2D
	width, height int32
}

func NewRenderSystem(world *ecs.World, width, height int) *RenderSystem {
	renderFilter := world.CreateFilter(components.Position{}, rl.Color{})

	imageBuffer := make([]color.RGBA, width*height*4)
	img := rl.GenImageColor(width, height, rl.Black)
	texture := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)

	return &RenderSystem{
		world:        world,
		renderFilter: renderFilter,
		imageBuffer:  imageBuffer,
		texture:      texture,
		width:        int32(width),
		height:       int32(height),
	}
}

func (rs *RenderSystem) SetSize(wid, hei int) {
	imageBuffer := make([]color.RGBA, wid*hei*4)
	img := rl.GenImageColor(wid, hei, rl.Black)
	texture := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	rs.width, rs.height = int32(wid), int32(hei)
	rs.texture = texture
	rs.imageBuffer = imageBuffer
}

func (rs *RenderSystem) Update(_ float64) {
	for i := range rs.imageBuffer {
		rs.imageBuffer[i] = color.RGBA{}
	}

	entities := rs.world.Query(rs.renderFilter)

	for _, entity := range entities {
		pos, _ := ecs.GetComponent[components.Position](rs.world, entity)
		col, _ := ecs.GetComponent[rl.Color](rs.world, entity)

		x, y := int32(pos.X), int32(pos.Y)
		if x >= 0 && x < rs.width && y >= 0 && y < rs.height {
			index := (y*rs.width + x)
			rs.imageBuffer[index] = color.RGBA{col.R, col.G, col.B, col.A}
		}
	}

	rl.UpdateTexture(rs.texture, rs.imageBuffer)

	rl.DrawTexture(rs.texture, 0, 0, rl.White)
}

func (rs *RenderSystem) Unload() {
	rl.UnloadTexture(rs.texture)
}
