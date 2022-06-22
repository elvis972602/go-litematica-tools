package litematica

import (
	"compress/gzip"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/nbt"
	"io"
)

type NBTFile struct {
	Blocks      []Blocks     `nbt:"blocks"`
	Entities    []Entity     `nbt:"entities"`
	Palette     []BlockState `nbt:"palette"`
	Size        []int32      `nbt:"size" nbt_type:"list"`
	Author      string       `nbt:"author"`
	DataVersion int32
}

type NBTFileWithRawMessage struct {
	Blocks      []Blocks `nbt:"blocks"`
	Entities    []Entity `nbt:"entities"`
	Palette     []state  `nbt:"palette"`
	Size        []int32  `nbt:"size" nbt_type:"list"`
	Author      string   `nbt:"author"`
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

func LoadNBT(r io.Reader) (*NBTFile, error) {
	var temp *NBTFileWithRawMessage
	reader, err := gzip.NewReader(r)
	_, err = nbt.NewDecoder(reader).Decode(&temp)
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

func (n *NBTFile) Encode(w io.Writer) error {

	gw := gzip.NewWriter(w)
	defer gw.Close()
	err := nbt.NewEncoder(gw).Encode(n, "")
	if err != nil {
		return err
	}
	return nil
}

func (n *NBTFile) ToLitematica(name string) *litematicaProject {
	l := NewLitematicaProject(name, int(n.Size[0]), int(n.Size[1]), int(n.Size[2]))
	for _, v := range n.Blocks {
		l.SetBlock(int(v.Pos[0]), int(v.Pos[1]), int(v.Pos[2]), n.Palette[v.State].Properties)
	}
	l.metaData.Author = n.Author
	return l
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
