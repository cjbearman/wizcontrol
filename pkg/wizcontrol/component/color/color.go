package color

import "fmt"

type Colors map[string]*Color

type Color struct {
	name, color string
}

func Build(cfg *[]ConfigYAML) (Colors, error) {
	colors := make(Colors)
	for _, c := range *cfg {
		if _, found := colors[c.Name]; found {
			return nil, fmt.Errorf("color %q is defined more than once", c.Name)
		}
		if other := colors.getByColor(c.Color); other != nil {
			return nil, fmt.Errorf("color %q has same value as color %q", c.Name, other.name)
		}
		colors[c.Name] = &Color{
			name:  c.Name,
			color: c.Color,
		}
	}
	return colors, nil
}

func (c Colors) getByColor(color string) *Color {
	for _, col := range c {
		if col.color == color {
			return col
		}
	}
	return nil
}

func (c *Color) String() string {
	return fmt.Sprintf("%s: %s", c.name, c.color)
}
