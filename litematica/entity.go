package litematica

// Entity TODO:Add Entity Structure
type Entity interface {
	ID() string
}

type ID struct {
	ID string `nbt:"id"`
}

type GlowItemFrame struct {
	Air            int16
	Facing         uint8
	FallDistance   float32
	Fire           int16
	Fixed          uint8
	Invisible      uint8
	Invulnerable   uint8
	Item           FrameItem `nbt_omitempty:"true"`
	ItemDropChance float32   `nbt_omitempty:"true"`
	ItemRotation   uint8     `nbt_omitempty:"true"`
	Motion         []float64 `nbt_type:"list"`
	OnGround       uint8
	PortalCooldown int32
	Pos            []float64 `nbt_type:"list"`
	Rotation       []float32 `nbt_type:"list"`
	TileX          int32
	TileY          int32
	TileZ          int32
	UUID           []int32 `nbt_type:"list"`
	Id             string  `nbt:"id"`
}

type FrameItem struct {
	Count int32
	ID    string `nbt:"id"`
	Tag   Tag    `nbt:"tag"`
}

type Tag struct {
	Map int32 `nbt:"map"`
}

func (e GlowItemFrame) ID() string { return "minecraft:glow_item_frame" }

var ByID = map[string]Entity{
	"minecraft:glow_item_frame": GlowItemFrame{},
}
