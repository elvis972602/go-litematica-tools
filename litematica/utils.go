package litematica

import "github.com/Tnze/go-mc/level/block"

func NewBlockState(b block.Block) BlockState {
	return BlockState{
		Name:       b.ID(),
		Properties: b,
	}
}

func (s Vec3D) outOfSize(x, y, z int) bool {
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
