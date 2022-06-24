package schematic

import (
	"compress/gzip"
	"fmt"
	"github.com/Tnze/go-mc/nbt"
	"io"
	"log"
	"math/bits"
)

type LitematicWithRawMessage struct {
	Metadata             Metadata
	MinecraftDataVersion int32
	Version              int32
	Regions              map[string]RegionWithRawMessage
}

type Litematic struct {
	Metadata             Metadata
	MinecraftDataVersion int32
	Version              int32
	Regions              map[string]Region
}

type Metadata struct {
	Author        string
	Description   string
	EnclosingSize Vec3D
	Name          string
	RegionCount   int32
	TimeCreated   int64
	TimeModified  int64
	TotalBlocks   int32
	TotalVolume   int32
}

type RegionWithRawMessage struct {
	BlockStatePalette []state
	TileEntities      []CompoundTag
	Entities          []nbt.RawMessage
	Position          Vec3D
	Size              Vec3D
	BlockStates       []int64
}

type Region struct {
	BlockStatePalette []BlockState
	TileEntities      []CompoundTag
	Entities          []Entity
	Position          Vec3D
	Size              Vec3D
	BlockStates       []int64
}

type Vec3D struct {
	X int32 `nbt:"x"`
	Y int32 `nbt:"y"`
	Z int32 `nbt:"z"`
}

func ReadLitematica(r io.Reader) (*Litematic, error) {
	var project *LitematicWithRawMessage
	reader, err := gzip.NewReader(r)
	if err != nil {
		log.Fatal(err)
	}
	_, err = nbt.NewDecoder(reader).Decode(&project)
	return &Litematic{
		Metadata:             project.Metadata,
		MinecraftDataVersion: project.MinecraftDataVersion,
		Version:              project.Version,
		Regions:              parseRegion(project.Regions),
	}, err
}

func parseRegion(rr map[string]RegionWithRawMessage) map[string]Region {
	var m = make(map[string]Region)
	for s, r := range rr {
		m[s] = Region{
			BlockStatePalette: parseBlocks(r.BlockStatePalette),
			TileEntities:      r.TileEntities,
			Entities:          parseEntities(r.Entities),
			Position:          r.Position,
			Size:              r.Size,
			BlockStates:       r.BlockStates,
		}
	}
	return m
}

func (l *Litematic) toProject() *Project {
	reg, regName, err := l.GetRegion()
	if err != nil {
		log.Fatal(err)
	}
	return &Project{
		metaData:   l.Metadata,
		regionName: regName,
		regionSize: reg.Size,
		palette:    newBlockStatePaletteWithData(reg.BlockStatePalette),
		data:       NewLitematicaBitArray(bits.Len(uint(len(reg.BlockStatePalette))), int(l.Metadata.TotalVolume), reg.BlockStates),
		entity:     newEntityContainerWithData(reg.Entities),
	}
}

func (l *Litematic) Encode(w io.Writer) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()
	err := nbt.NewEncoder(gw).Encode(l, "")
	if err != nil {
		return err
	}
	return nil
}

func (l *Litematic) GetRegion() (Region, string, error) {
	for k, r := range l.Regions {
		return r, k, nil
	}
	return Region{}, "", fmt.Errorf("there is no region in this litematic file") //empty
}
