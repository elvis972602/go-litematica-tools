package schematic

import (
	"compress/gzip"
	"github.com/Tnze/go-mc/nbt"
	"io"
	"path/filepath"
)

type Nbt struct {
	Blocks      []Blocks     `nbt:"blocks"`
	Entities    []Entity     `nbt:"entities"`
	Palette     []BlockState `nbt:"palette"`
	Size        []int32      `nbt:"size" nbt_type:"list"`
	Author      string       `nbt:"author"`
	DataVersion int32
}

type NbtWithRawMessage struct {
	Blocks      []Blocks         `nbt:"blocks"`
	Entities    []nbt.RawMessage `nbt:"entities"`
	Palette     []state          `nbt:"palette"`
	Size        []int32          `nbt:"size" nbt_type:"list"`
	Author      string           `nbt:"author"`
	DataVersion int32
}

type CompoundTag struct {
	Name string
}

type Blocks struct {
	Pos   []int32 `nbt:"pos" nbt_type:"list"`
	State int32   `nbt:"state"`
}

func ReadNbtFile(r io.Reader) (*Nbt, error) {
	var n *NbtWithRawMessage
	reader, err := gzip.NewReader(r)
	defer reader.Close()
	_, err = nbt.NewDecoder(reader).Decode(&n)
	if err != nil {
		return nil, err
	}
	return &Nbt{
		Blocks:      n.Blocks,
		Entities:    parseEntities(n.Entities),
		Palette:     parseBlocks(n.Palette),
		Size:        n.Size,
		Author:      n.Author,
		DataVersion: n.DataVersion,
	}, nil
}

func (n *Nbt) Encode(w io.Writer) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()
	err := nbt.NewEncoder(gw).Encode(n, "")
	if err != nil {
		return err
	}
	return nil
}

func (n *Nbt) toProject(name string) *Project {
	name = filepath.Base(name)
	name = name[:len(name)-len(filepath.Ext(name))]
	l := NewProject(name, int(n.Size[0]), int(n.Size[1]), int(n.Size[2]))
	for _, v := range n.Blocks {
		l.SetBlock(int(v.Pos[0]), int(v.Pos[1]), int(v.Pos[2]), n.Palette[v.State].Properties)
	}
	l.MetaData.Author = n.Author
	l.MinecraftDataVersion = int32(defaultMinecraftDataVersion)
	l.Version = int32(defaultVersion)
	return l
}
