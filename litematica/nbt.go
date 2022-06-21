package litematica

import (
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/nbt"
	"io"
)

type NBTFile struct {
	Blocks      []Blocks      `nbt:"blocks"`
	Entities    []CompoundTag `nbt:"entities"`
	Palette     []BlockState  `nbt:"palette"`
	Size        []int32       `nbt:"size" nbt_type:"list"`
	Author      string        `nbt:"author"`
	DataVersion int32
}

type NBTFileWithRawMessage struct {
	Blocks      []Blocks      `nbt:"blocks"`
	Entities    []CompoundTag `nbt:"entities"`
	Palette     []state       `nbt:"palette"`
	Size        []int32       `nbt:"size" nbt_type:"list"`
	Author      string        `nbt:"author"`
	DataVersion int32
}

type CompoundTag struct {
	Name string
}

type Blocks struct {
	Pos   []int32 `nbt:"pos" nbt_type:"list"`
	State int32   `nbt:"state"`
}

type state struct {
	Name       string
	Properties nbt.RawMessage
}

func LoadNBT(reader io.Reader) (*NBTFile, error) {
	var temp *NBTFileWithRawMessage
	_, err := nbt.NewDecoder(reader).Decode(&temp)
	if err != nil {
		return nil, err
	}
	return &NBTFile{
		Blocks:      temp.Blocks,
		Entities:    temp.Entities,
		Palette:     stateToBlock(temp.Palette),
		Size:        temp.Size,
		Author:      temp.Author,
		DataVersion: temp.DataVersion,
	}, nil
}

func stateToBlock(states []state) []BlockState {
	var blockPalette []BlockState
	for _, state := range states {
		b := block.FromID[state.Name]
		if state.Properties.Type != nbt.TagEnd {
			err := state.Properties.Unmarshal(&b)
			if err != nil {
				panic(err)
			}
		}
		blockPalette = append(blockPalette, BlockState{Name: state.Name, Properties: b})
	}
	return blockPalette
}
