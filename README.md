# go-litematica-tools

go-litematica-tools is a Go library for reading and writing litematica and NBT files.

## Installation
```shell
go get github.com/elvis972602/go-litematica-tools
```
## Usage

```go
package main

import (
	"github.com/Tnze/go-mc/level/block"
	"github.com/elvis972602/go-litematica-tools/schematic"
	"os"
)

func main() {
	// create project
	project := schematic.NewProject("Test", 16, 16, 16)

	// set block
	project.SetBlock(0, 0, 0, block.Dirt{})

	// get block
	block := project.GetBlock(0, 0, 0)

	// encode
	file, err := os.Create("test.litematic")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	project.Encode(file)

	// read file
	file, err := os.Open("test.litematic")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	project, err := schematic.LoadFromFile(file)
	if err != nil {
		panic(err)
	}
}
```

## API
### func NewProject
```go
func NewProject(name string, x, y, z int) *Project
```
NewProject creates a new Project instance with the given name and dimensions.

### func LoadFromFile
```go
func LoadFromFile(file *os.File) (*Project, error)
```
LoadFromFile reads a litematica or NBT file and returns a Project instance.

### func (p *Project) SetBlock
```go
func (p *Project) SetBlock(x, y, z int, b BlockState)
```
SetBlock sets the block at the given coordinates to the given block.

### func (p *Project) GetBlock
```go
func (p *Project) GetBlock(x, y, z int) BlockState
```
GetBlock returns the block at the given coordinates.

### func (p *Project) Encode
```go
func (p *Project) Encode(w io.Writer) error
```
Encode encodes the project as a litematica file.

## License
This library is released under the MIT license. See [LICENSE](https://github.com/elvis972602/go-litematica-tools/blob/master/LICENSE) for more details.

#### Note: This README was generated with assistance from GPT-3




