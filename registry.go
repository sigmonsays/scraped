package scraped

func NewRegistry() *Registry {
	reg := &Registry{}
	return reg
}

type Registry struct {
	items []Scraper
}

func (me *Registry) UnloadAll() {
	me.items = nil
}

func (me *Registry) Register(s Scraper) {
	me.items = append(me.items, s)
	log.Debugf("Register scraper %s", s.GetType().ID)
}

func (me *Registry) Iterate() []Scraper {
	return me.items
}
