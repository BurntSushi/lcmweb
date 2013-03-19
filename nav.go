package main

type nav struct {
	Name string
	Link string
}

func (c *controller) mkNav(navs ...nav) []string {
	strs := make([]string, len(navs))
	for i, n := range navs {
		strs[i] = c.renderString("bit_nav", n)
	}
	return strs
}
