package litematica

import (
	"compress/gzip"
	"fmt"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/nbt"
	"io"
	"log"
	"math/bits"
	"time"
)

var (
	author               = "Author"
	description          = ""
	minecraftDataVersion = 2975
	version              = 6
)

type litematicaProject struct {
	metaData Metadata

	//Region's key
	regionName string

	regionSize Vec3D

	data *bitArray

	palette *blockStatePalette

	entity *entityContainer
}

func NewLitematicaProject(name string, x, y, z int) *litematicaProject {
	return &litematicaProject{
		metaData: Metadata{
			Author:        author,
			Description:   description,
			EnclosingSize: Vec3D{int32(x), int32(y), int32(z)},
			Name:          name,
			RegionCount:   1,
			TimeCreated:   time.Now().UnixMilli(),
			TimeModified:  time.Now().UnixMilli(),
			TotalBlocks:   0,
			TotalVolume:   int32(x * y * z),
		},
		regionName: name,
		data:       newEmptyBitArray(x * y * z),
		palette:    newBlockStatePalette(),
		entity:     newEntityContainer(),
	}
}

func LoadLitematicaFromFile(f io.Reader) (*litematicaProject, error) {
	reader, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	project, err := LoadLitematica(reader)
	if err != nil {
		return nil, err
	}
	reg, regName, err := project.GetRegion()
	if err != nil {
		return nil, err
	}
	return &litematicaProject{
		metaData:   project.Metadata,
		regionName: regName,
		regionSize: reg.Size,
		palette:    newBlockStatePaletteWithData(reg.BlockStatePalette),
		data:       newLitematicaBitArray(bits.Len(uint(len(reg.BlockStatePalette))), int(project.Metadata.TotalVolume), reg.BlockStates),
		entity:     newEntityContainerWithData(reg.Entities),
	}, nil
}

func (p *litematicaProject) getIndex(x, y, z int) int {
	return p.metaData.EnclosingSize.getIndex(x, y, z)
}

func (p *litematicaProject) Palette() []BlockState {
	return p.palette.palette
}

func (p *litematicaProject) GetBlock(x, y, z int) BlockState {
	if !p.metaData.EnclosingSize.outOfSize(x, y, z) {
		return p.palette.value(p.data.getBlock(int64(p.getIndex(x, y, z))))
	} else {
		panic(fmt.Sprintf("GetBlock out of range : enclosingSize: %v, x: %d, y: %d, z: %d", p.metaData.EnclosingSize, x, y, z))
	}
}

func (p *litematicaProject) SetBlock(x, y, z int, b block.Block) {
	if !p.metaData.EnclosingSize.outOfSize(x, y, z) {
		if p.GetBlock(x, y, z).Name == air || b.ID() != air {
			p.metaData.TotalBlocks++
		} else if b.ID() == air {
			p.metaData.TotalBlocks--
		}
		p.data.setBlock(int64(p.getIndex(x, y, z)), p.palette.id(NewBlockState(b)))
	} else {
		panic(fmt.Sprintf("SetBlock out of range : enclosingSize: %v, x: %d, y: %d, z: %d", p.metaData.EnclosingSize, x, y, z))
	}
}

func (p *litematicaProject) Contain(block BlockState) bool {
	return p.palette.contain(block)
}

func (p *litematicaProject) AddEntity(e Entity) {
	p.entity.addEntity(e)
}

func (p *litematicaProject) SetAuthor(author string) {
	p.metaData.Author = author
}

func (p *litematicaProject) SetDescription(description ...any) {
	p.metaData.Description = fmt.Sprint(description...)
}

func (p *litematicaProject) SetName(name string) {
	p.metaData.Name = name
}

func (p *litematicaProject) SetRegionsName(name string) {
	p.regionName = name
}

func (p *litematicaProject) SetMinecraftDataVersion(v int) {
	minecraftDataVersion = v
}

func (p *litematicaProject) SetVersion(v int) {
	version = v
}

func (p *litematicaProject) XRange() int {
	return int(p.metaData.EnclosingSize.X)
}

func (p *litematicaProject) YRange() int {
	return int(p.metaData.EnclosingSize.Y)
}

func (p *litematicaProject) ZRange() int {
	return int(p.metaData.EnclosingSize.Z)
}

func (p *litematicaProject) Size() Vec3D {
	return p.metaData.EnclosingSize
}

func (p *litematicaProject) Encode(w io.Writer) error {
	project := Project{
		Metadata:             p.metaData,
		MinecraftDataVersion: int32(minecraftDataVersion),
		Version:              int32(version),
		Regions:              p.getRegion(),
	}
	project.Metadata.TimeModified = time.Now().UnixMilli()
	gw := gzip.NewWriter(w)
	err := nbt.NewEncoder(gw).Encode(project, "")
	if err != nil {
		return err
	}
	return nil

}

func (p *litematicaProject) getRegion() map[string]Region {
	rs := make(map[string]Region)
	r := Region{
		BlockStatePalette: p.palette.palette,
		TileEntities:      []CompoundTag{},
		Entities:          p.entity.entity,
		Position:          Vec3D{},
		Size:              p.regionSize,
		BlockStates:       p.data.data,
	}
	rs[p.regionName] = r
	return rs
}

const air = "minecraft:air"

type blockStatePalette struct {
	paletteMap map[BlockState]int
	palette    []BlockState
}

func newBlockStatePalette() *blockStatePalette {
	Air := BlockState{Name: air, Properties: block.Air{}}
	m := make(map[BlockState]int)
	m[Air] = 0
	return &blockStatePalette{
		paletteMap: m,
		palette:    []BlockState{Air},
	}
}

func newBlockStatePaletteWithData(data []BlockState) *blockStatePalette {
	m := make(map[BlockState]int, len(data))
	for n, b := range data {
		m[b] = n
	}
	return &blockStatePalette{
		paletteMap: m,
		palette:    data,
	}
}

func (p *blockStatePalette) id(block BlockState) int {
	if _, ok := p.paletteMap[block]; !ok {
		p.paletteMap[block] = len(p.palette)
		p.palette = append(p.palette, block)
	}
	return p.paletteMap[block]
}

func (p *blockStatePalette) value(index int) BlockState {
	return p.palette[index]
}

func (p *blockStatePalette) contain(block BlockState) bool {
	_, ok := p.paletteMap[block]
	return ok
}

type entityContainer struct {
	entity []Entity
}

func newEntityContainer() *entityContainer {
	return &entityContainer{}
}

func newEntityContainerWithData(entity []Entity) *entityContainer {
	return &entityContainer{
		entity: entity,
	}
}

func (e *entityContainer) addEntity(d Entity) {
	e.entity = append(e.entity, d)
}
