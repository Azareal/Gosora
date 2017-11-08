package main

type RouteSubset struct {
	RouteList []*RouteImpl
}

func (set *RouteSubset) Before(line string, literal ...bool) *RouteSubset {
	var litItem bool
	if len(literal) > 0 {
		litItem = literal[0]
	}
	for _, route := range set.RouteList {
		route.RunBefore = append(route.RunBefore, Runnable{line, litItem})
	}
	return set
}

func (set *RouteSubset) Not(path ...string) *RouteSubset {
	for i, route := range set.RouteList {
		if inStringList(route.Path, path) {
			set.RouteList = append(set.RouteList[:i], set.RouteList[i+1:]...)
		}
	}
	return set
}
