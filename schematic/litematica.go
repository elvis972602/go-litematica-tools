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

func ReadLitematicaFile(r io.Reader) (*Litematic, error) {
	var l *LitematicWithRawMessage
	reader, err := gzip.NewReader(r)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	_, err = nbt.NewDecoder(reader).Decode(&l)
	if err != nil {
		return nil, err
	}
	return &Litematic{
		Metadata:             l.Metadata,
		MinecraftDataVersion: l.MinecraftDataVersion,
		Version:              l.Version,
		Regions:              parseRegion(l.Regions),
	}, nil
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
		MetaData:             l.Metadata,
		MinecraftDataVersion: l.MinecraftDataVersion,
		Version:              l.Version,
		RegionName:           regName,
		regionSize:           reg.Size,
		palette:              newBlockStatePaletteWithData(reg.BlockStatePalette),
		data:                 NewBitArray(bits.Len(uint(len(reg.BlockStatePalette)-1)), int(l.Metadata.TotalVolume), reg.BlockStates),
		entity:               newEntityContainerWithData(reg.Entities),
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
