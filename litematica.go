package go_litematica

import (
	"compress/gzip"
	"fmt"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/nbt"
	"io"
	"log"
)

type ProjectWithRawMessage struct {
	Metadata             Metadata
	MinecraftDataVersion int32
	Version              int32
	Regions              map[string]RegionWithRawMessage
}

type Project struct {
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

type BlockState struct {
	Name       string
	Properties block.Block
}

func LoadLitematica(r io.Reader) (*Project, error) {
	var project *ProjectWithRawMessage
	reader, err := gzip.NewReader(r)
	if err != nil {
		log.Fatal(err)
	}
	_, err = nbt.NewDecoder(reader).Decode(&project)
	return &Project{
		Metadata:             project.Metadata,
		MinecraftDataVersion: project.MinecraftDataVersion,
		Version:              project.Version,
		Regions:              toRegion(project.Regions),
	}, err
}

func toRegion(rr map[string]RegionWithRawMessage) map[string]Region {
	var m = make(map[string]Region)
	for s, r := range rr {
		m[s] = Region{
			BlockStatePalette: stateToBlock(r.BlockStatePalette),
			TileEntities:      r.TileEntities,
			Entities:          toEntity(r.Entities),
			Position:          r.Position,
			Size:              r.Size,
			BlockStates:       r.BlockStates,
		}
	}
	return m
}

func toEntity(e []nbt.RawMessage) []Entity {
	var entities []Entity
	for _, i := range e {
		if i.Type != nbt.TagEnd {
			id := ID{}
			err := i.Unmarshal(&id)
			if err != nil {
				panic(err)
			}
			entity := ByID[id.ID]
			err = i.Unmarshal(&entity)
			if err != nil {
				panic(err)
			}
			entities = append(entities, entity)
		}
	}
	return entities
}

func (p *Project) GetRegion() (Region, string, error) {
	for k, r := range p.Regions {
		return r, k, nil //there is always at least one region theoretically
	}
	return Region{}, "", fmt.Errorf("there is no region in this priject") //empty
}
