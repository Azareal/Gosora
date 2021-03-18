package main

type Prec struct {
	Sets      []map[string]int
	NameToSet map[string]int
}

func NewPrec() *Prec {
	return &Prec{NameToSet: make(map[string]int)}
}

func (p *Prec) AddSet(precs ...string) {
	set := make(map[string]int)
	setIndex, i := len(p.Sets), 0
	for _, prec := range precs {
		set[prec] = i
		p.NameToSet[prec] = setIndex
		i++
	}
	p.Sets = append(p.Sets, set)
}

func (p *Prec) InAnySet(name string) bool {
	_, ok := p.NameToSet[name]
	return ok
}

func (p *Prec) InSameSet(n, n2 string) bool {
	ok, ok2 := p.InAnySet(n), p.InAnySet(n2)
	if !ok || !ok2 {
		return false
	}
	set1, set2 := p.NameToSet[n], p.NameToSet[n2]
	return set1 == set2
}

func (p *Prec) GreaterThan(greater, lesser string) bool {
	if !p.InSameSet(greater, lesser) {
		return false
	}
	set := p.Sets[p.NameToSet[greater]]
	return set[greater] > set[lesser]
}

func (p *Prec) LessThanItem(greater string) (l []string) {
	if len(p.Sets) == 0 {
		return nil
	}

	setIndex := p.NameToSet[greater]
	set := p.Sets[setIndex]
	ref := set[greater]
	for name, value := range set {
		if value < ref {
			l = append(l, name)
		}
	}

	return l
}
