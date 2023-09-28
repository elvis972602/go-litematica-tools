package schematic

import (
	"compress/gzip"
	"fmt"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/nbt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	defaultAuthor               = "Author"
	defaultDescription          = ""
	defaultMinecraftDataVersion = 2975
	defaultVersion              = 6
)

func SetDefaultAuthor(s string) {
	defaultAuthor = s
}

func SetDefaultDescription(s string) {
	defaultDescription = s
}

func SetDefaultMinecraftDataVersion(v int) {
	defaultMinecraftDataVersion = v
}

func SetDefaultVersion(v int) {
	defaultVersion = v
}

type Project struct {
	MetaData Metadata

	MinecraftDataVersion int32

	Version int32

	//Region's key
	RegionName string

	regionSize Vec3D

	data *BitArray

	palette *blockStatePalette

	entity *entityContainer
}

func NewProject(name string, x, y, z int) *Project {
	return &Project{
		MetaData: Metadata{
			Author:        defaultAuthor,
			Description:   defaultDescription,
			EnclosingSize: Vec3D{int32(x), int32(y), int32(z)},
			Name:          name,
			RegionCount:   1,
			TimeCreated:   time.Now().UnixMilli(),
			TimeModified:  time.Now().UnixMilli(),
			TotalBlocks:   0,
			TotalVolume:   int32(x * y * z),
		},
		MinecraftDataVersion: int32(defaultMinecraftDataVersion),
		Version:              int32(defaultVersion),
		RegionName:           name,
		regionSize:           Vec3D{int32(x), int32(y), int32(z)},
		data:                 NewEmptyBitArray(x * y * z),
		palette:              newBlockStatePalette(),
		entity:               newEntityContainer(),
	}
}

func LoadFromFile(file *os.File) (*Project, error) {
	ext := filepath.Ext(file.Name())
	if ext == ".litematic" {
		return LoadFromLitematic(file)
	} else if ext == ".nbt" {
		return LoadFromNbt(file.Name(), file)
	} else {
		return nil, fmt.Errorf("unsuppot file format: %s", ext)
	}
}

func LoadFromLitematic(f io.Reader) (*Project, error) {
	l, err := ReadLitematicaFile(f)
	if err != nil {
		return nil, err
	}
	return l.toProject()
}

func LoadFromNbt(name string, f io.Reader) (*Project, error) {
	n, err := ReadNbtFile(f)
	if err != nil {
		return nil, err
	}
	return n.toProject(name), nil
}

func (p *Project) index(x, y, z int) int {
	return p.MetaData.EnclosingSize.getIndex(x, y, z)
}

func (p *Project) Data() []int64 {
	return p.data.data
}

func (p *Project) Palette() []BlockState {
	return p.palette.palette
}

func (p *Project) GetBlock(x, y, z int) BlockState {
	if p.MetaData.EnclosingSize.outOfRange(x, y, z) {
		panic(fmt.Sprintf("GetBlock out of range : enclosingSize: %v,Pos: %d, %d, %d", p.MetaData.EnclosingSize, x, y, z))
	}
	return p.palette.value(p.data.getBlock(int64(p.index(x, y, z))))
}

func (p *Project) SetBlock(x, y, z int, b block.Block) {
	if p.MetaData.EnclosingSize.outOfRange(x, y, z) {
		panic(fmt.Sprintf("SetBlock out of range : enclosingSize: %v,Pos: %d, %d, %d", p.MetaData.EnclosingSize, x, y, z))
	}
	if p.GetBlock(x, y, z).Name == air || b.ID() != air {
		p.MetaData.TotalBlocks++
	} else if b.ID() == air {
		p.MetaData.TotalBlocks--
	}
	p.data.setBlock(int64(p.index(x, y, z)), p.palette.id(NewBlockState(b)))
}

func (p *Project) Contain(block BlockState) bool {
	return p.palette.contain(block)
}

func (p *Project) AddEntity(e Entity) {
	p.entity.addEntity(e)
}

func (p *Project) XRange() int {
	return int(p.MetaData.EnclosingSize.X)
}

func (p *Project) YRange() int {
	return int(p.MetaData.EnclosingSize.Y)
}

func (p *Project) ZRange() int {
	return int(p.MetaData.EnclosingSize.Z)
}

func (p *Project) Size() Vec3D {
	return p.MetaData.EnclosingSize
}

func (p *Project) ChangeMaterial(from, to block.Block) {
	p.palette.changeMaterial(NewBlockState(from), NewBlockState(to))
}

// Encode default encode litematic file
func (p *Project) Encode(w io.Writer) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()
	err := nbt.NewEncoder(gw).Encode(p.Litematic(), "")
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
	rs[p.RegionName] = r
	return rs
}

func (p *Project) Litematic() *Litematic {
	project := &Litematic{
		Metadata:             p.MetaData,
		MinecraftDataVersion: p.MinecraftDataVersion,
		Version:              p.Version,
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
				s := int32(p.data.getBlock(int64(p.index(x, y, z))))
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
		Author:      p.MetaData.Author,
		DataVersion: int32(defaultMinecraftDataVersion),
	}
}

const air = "minecraft:air"

type blockStatePalette struct {
	paletteMap map[BlockState]int
	palette    []BlockState
	sync.RWMutex
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
	p.Lock()
	defer p.Unlock()
	if _, ok := p.paletteMap[block]; !ok {
		p.paletteMap[block] = len(p.palette)
		p.palette = append(p.palette, block)
	}
	return p.paletteMap[block]
}

func (p *blockStatePalette) value(index int) BlockState {
	p.Lock()
	defer p.Unlock()
	return p.palette[index]
}

func (p *blockStatePalette) contain(block BlockState) bool {
	p.Lock()
	defer p.Unlock()
	_, ok := p.paletteMap[block]
	return ok
}

func (p *blockStatePalette) changeMaterial(from, to BlockState) {
	p.Lock()
	defer p.Unlock()
	index, ok := p.paletteMap[from]
	if !ok {
		return
	}
	p.palette[index] = to
	delete(p.paletteMap, from)
	p.paletteMap[to] = index
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
