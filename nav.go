package main

type nav struct {
	Name string
	Link string
}

func (w *web) mkNav(navs ...nav) []string {
	strs := make([]string, len(navs))
	for i, n := range navs {
		data := m{"NavItem": n}
		strs[i] = string(w.renderBytes("bit_nav", data))
	}
	return strs
}
