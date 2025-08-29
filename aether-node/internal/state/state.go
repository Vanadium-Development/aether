package state

import (
	"math/rand"

	"github.com/google/uuid"
)

type RGBColor struct {
	R uint8
	G uint8
	B uint8
}

func rgb(r uint8, g uint8, b uint8) RGBColor {
	return RGBColor{R: r, G: g, B: b}
}

var availableColors = []RGBColor{
	rgb(52, 152, 219),  // Peter River
	rgb(231, 76, 60),   // Alizarin
	rgb(26, 188, 156),  // Turquoise
	rgb(155, 89, 182),  // Amethyst
	rgb(153, 128, 250), // Forgotten Purple
	rgb(237, 76, 103),  // Bara Red
	rgb(84, 109, 229),  // Cornflower
}

func RandomNodeColor() RGBColor {
	return availableColors[rand.Intn(len(availableColors))]
}

type Node struct {
	ID    uuid.UUID
	Name  string
	Color RGBColor
}

func (node *Node) NodeInfoMap() map[string]interface{} {
	return map[string]interface{}{
		"id":   node.ID.String(),
		"name": node.Name,
		"color": map[string]interface{}{
			"r": node.Color.R,
			"g": node.Color.G,
			"b": node.Color.B,
		},
	}
}
