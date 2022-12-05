package schematic

import (
	"github.com/Tnze/go-mc/level/block"
	"math/rand"
	"os"
	"testing"
	"time"
)

type Pos struct {
	x, y, z int
}

func TestNewLitematicProject(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	size := Pos{randInt(1, 1000), randInt(1, 1000), randInt(1, 1000)}
	blocksCount := randInt(1000, 10000)
	t.Logf("Size: %v, blocks count: %v", size, blocksCount)
	blockMap := make(map[Pos]block.Block)
	project := NewProject("test", size.x, size.y, size.z)
	for i := 0; i < blocksCount; i++ {
		b := block.StateList[rand.Intn(len(block.StateList))]
		p := Pos{rand.Intn(size.x), rand.Intn(size.y), rand.Intn(size.z)}
		blockMap[p] = b
		project.SetBlock(p.x, p.y, p.z, b)
	}
	for k, v := range blockMap {
		if project.GetBlock(k.x, k.y, k.z).Properties != v {
			t.Fatalf("Error, pos: %v, wrong block: %s, correct blocks: %s", k, project.GetBlock(k.x, k.y, k.z).Properties.ID(), v.ID())
		}
	}
}

func TestEncodeNbt(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	size := Pos{randInt(1, 1000), randInt(1, 1000), randInt(1, 1000)}
	blocksCount := randInt(1000, 10000)
	t.Logf("Size: %v, blocks count: %v", size, blocksCount)
	blockMap := make(map[Pos]block.Block)
	project := NewProject("test", size.x, size.y, size.z)
	for i := 0; i < blocksCount; i++ {
		b := block.StateList[rand.Intn(len(block.StateList))]
		p := Pos{rand.Intn(size.x), rand.Intn(size.y), rand.Intn(size.z)}
		blockMap[p] = b
		project.SetBlock(p.x, p.y, p.z, b)
	}
	fw, err := os.Create("test.nbt")
	if err != nil {
		t.Fatal(err)
	}
	project.Nbt().Encode(fw)

	f, err := os.Open("test.nbt")
	if err != nil {
		t.Fatal(err)
	}
	p, err := LoadFromFile(f)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range blockMap {
		if p.GetBlock(k.x, k.y, k.z).Properties != v {
			t.Fatalf("Error, pos: %v, wrong block: %s, correct blocks: %s", k, p.GetBlock(k.x, k.y, k.z).Properties.ID(), v.ID())
		}
	}

}

func TestEncodeLitematic(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	size := Pos{randInt(1, 1000), randInt(1, 1000), randInt(1, 1000)}
	blocksCount := randInt(1000, 10000)
	t.Logf("Size: %v, blocks count: %v", size, blocksCount)
	blockMap := make(map[Pos]block.Block)
	project := NewProject("test", size.x, size.y, size.z)
	for i := 0; i < blocksCount; i++ {
		b := block.StateList[rand.Intn(len(block.StateList))]
		p := Pos{rand.Intn(size.x), rand.Intn(size.y), rand.Intn(size.z)}
		blockMap[p] = b
		project.SetBlock(p.x, p.y, p.z, b)
	}
	fw, err := os.Create("test.litematic")
	if err != nil {
		t.Fatal(err)
	}
	project.Encode(fw)

	f, err := os.Open("test.litematic")
	if err != nil {
		t.Fatal(err)
	}
	p, err := LoadFromFile(f)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range blockMap {
		if p.GetBlock(k.x, k.y, k.z).Properties != v {
			t.Fatalf("Error, pos: %v, wrong block: %s, correct blocks: %s", k, p.GetBlock(k.x, k.y, k.z).Properties.ID(), v.ID())
		}
	}

}

func randInt(min, max int) int {
	return min + rand.Intn(max-min)
}
