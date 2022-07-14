package schematic

import (
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/nbt"
)

type state struct {
	Name       string
	Properties nbt.RawMessage
}

type BlockState struct {
	Name       string
	Properties block.Block `nbt_omitempty:"true"`
}

func NewBlockState(b block.Block) BlockState {
	return BlockState{
		Name:       b.ID(),
		Properties: b,
	}
}

func (s Vec3D) outOfRange(x, y, z int) bool {
	if int(s.X) <= x || int(s.Y) <= y || int(s.Z) <= z {
		return true
	} else if x < 0 || y < 0 || z < 0 {
		return true
	}
	return false
}

func (s Vec3D) getIndex(x, y, z int) int {
	return y*int(s.X)*int(s.Z) + z*int(s.X) + x
}

func parseBlocks(states []state) []BlockState {
	var blockPalette []BlockState
	for _, s := range states {
		b := block.FromID[s.Name]
		if s.Properties.Type != nbt.TagEnd {
			err := s.Properties.Unmarshal(&b)
			if err != nil {
				panic(err)
			}
		}
		blockPalette = append(blockPalette, BlockState{Name: s.Name, Properties: b})
	}
	return blockPalette
}

func parseEntities(entities []nbt.RawMessage) []Entity {
	var e []Entity
	for _, i := range entities {
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
			e = append(e, entity)
		}
	}
	return e
}
