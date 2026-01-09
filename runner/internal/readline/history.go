package readline

// TODO: implement history persistence
type History struct {
	// config
	file string

	// state
	entries [][]rune
	index   int
}

func NewHistory(file string) *History {
	return &History{
		file:    file,
		entries: [][]rune{},
		index:   0,
	}
}

func (h *History) Add(entry []rune) {
	h.entries = append(h.entries, entry)
	h.index = len(h.entries)
}

func (h *History) Prev() []rune {
	if h.index > 0 {
		h.index--
		return h.entries[h.index]
	}
	return nil
}

func (h *History) Next() []rune {
	if h.index < len(h.entries)-1 {
		h.index++
		return h.entries[h.index]
	} else {
		h.index = len(h.entries)
		return []rune{}
	}
}

func (h *History) Save() error {
	return nil
}
