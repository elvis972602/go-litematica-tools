package schematic

import (
	"compress/gzip"
	"fmt"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/nbt"
	"io"
	"os"
	"strings"
	"time"
)

var (
	author               = "Author"
	description          = ""
	minecraftDataVersion = 2975
	version              = 6
)

type Project struct {
	metaData Metadata

	//Region's key
	regionName string

	regionSize Vec3D

	data *bitArray

	palette *blockStatePalette

	entity *entityContainer
}

func NewProject(name string, x, y, z int) *Project {
	return &Project{
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
		regionSize: Vec3D{int32(x), int32(y), int32(z)},
		data:       NewEmptyBitArray(x * y * z),
		palette:    newBlockStatePalette(),
		entity:     newEntityContainer(),
	}
}

func Load(file *os.File) (*Project, error) {
	if strings.HasSuffix(file.Name(), ".litematic") {
		return LoadFromLitematic(file)

	} else if strings.HasSuffix(file.Name(), ".nbt") {
		return LoadFromNbt(file.Name(), file)
	} else {
		return nil, fmt.Errorf("unsuppot file format")
	}
}

func LoadFromLitematic(f io.Reader) (*Project, error) {
	l, err := LoadLitematica(f)
	if err != nil {
		return nil, err
	}
	return l.toProject(), nil
}

func LoadFromNbt(name string, f io.Reader) (*Project, error) {
	n, err := LoadNBT(f)
	if err != nil {
		return nil, err
	}
	return n.toProject(name), nil
}

func (p *Project) Index(x, y, z int) int {
	return p.metaData.EnclosingSize.getIndex(x, y, z)
}

func (p *Project) Data() []int64 {
	return p.data.data
}

func (p *Project) Palette() []BlockState {
	return p.palette.palette
}

func (p *Project) GetBlock(x, y, z int) BlockState {
	if !p.metaData.EnclosingSize.outOfSize(x, y, z) {
		return p.palette.value(p.data.getBlock(int64(p.Index(x, y, z))))
	} else {
		panic(fmt.Sprintf("GetBlock out of range : enclosingSize: %v,Pos: %d, %d, %d", p.metaData.EnclosingSize, x, y, z))
	}
}

func (p *Project) SetBlock(x, y, z int, b block.Block) {
	if !p.metaData.EnclosingSize.outOfSize(x, y, z) {
		if p.GetBlock(x, y, z).Name == air || b.ID() != air {
			p.metaData.TotalBlocks++
		} else if b.ID() == air {
			p.metaData.TotalBlocks--
		}
		p.data.setBlock(int64(p.Index(x, y, z)), p.palette.id(NewBlockState(b)))
	} else {
		panic(fmt.Sprintf("SetBlock out of range : enclosingSize: %v,Pos: %d, %d, %d", p.metaData.EnclosingSize, x, y, z))
	}
}

func (p *Project) Contain(block BlockState) bool {
	return p.palette.contain(block)
}

func (p *Project) AddEntity(e Entity) {
	p.entity.addEntity(e)
}

func (p *Project) SetAuthor(author string) {
	p.metaData.Author = author
}

func (p *Project) SetDescription(description ...any) {
	p.metaData.Description = fmt.Sprint(description...)
}

func (p *Project) SetName(name string) {
	p.metaData.Name = name
}

func (p *Project) SetRegionsName(name string) {
	p.regionName = name
}

func (p *Project) SetMinecraftDataVersion(v int) {
	minecraftDataVersion = v
}

func (p *Project) SetVersion(v int) {
	version = v
}

func (p *Project) XRange() int {
	return int(p.metaData.EnclosingSize.X)
}

func (p *Project) YRange() int {
	return int(p.metaData.EnclosingSize.Y)
}

func (p *Project) ZRange() int {
	return int(p.metaData.EnclosingSize.Z)
}

func (p *Project) Size() Vec3D {
	return p.metaData.EnclosingSize
}

func (p *Project) Encode(w io.Writer) error {
	project := Litematic{
		Metadata:             p.metaData,
		MinecraftDataVersion: int32(minecraftDataVersion),
		Version:              int32(version),
		Regions:              p.region(),
	}
	project.Metadata.TimeModified = time.Now().UnixMilli()

	gw := gzip.NewWriter(w)
	defer gw.Close()
	err := nbt.NewEncoder(gw).Encode(project, "")
	if err != nil {
		return err
	}
	return nil
}

func (p *Project) region() map[string]Region {
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

func (p *Project) Litematic() *Litematic {
	project := &Litematic{
		Metadata:             p.metaData,
		MinecraftDataVersion: int32(minecraftDataVersion),
		Version:              int32(version),
		Regions:              p.region(),
	}
	project.Metadata.TimeModified = time.Now().UnixMilli()
	return project
}

func (p *Project) Nbt() *Nbt {
	var b []Blocks
	for x := 0; x < p.XRange(); x++ {
		for y := 0; y < p.YRange(); y++ {
			for z := 0; z < p.ZRange(); z++ {
				s := int32(p.data.getBlock(int64(p.Index(x, y, z))))
				if s != 0 {
					b = append(b, Blocks{Pos: []int32{int32(x), int32(y), int32(z)}, State: s - 1})
				}
			}
		}
	}
	return &Nbt{
		Blocks:      b,
		Entities:    p.entity.entity,
		Palette:     p.Palette()[1:],
		Size:        []int32{p.regionSize.X, p.regionSize.Y, p.regionSize.Z},
		Author:      p.metaData.Author,
		DataVersion: int32(minecraftDataVersion),
	}
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
