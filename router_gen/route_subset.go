package main

type RouteSubset struct {
	RouteList []*RouteImpl
}

func (set *RouteSubset) Before(lines ...string) *RouteSubset {
	for _, line := range lines {
		for _, route := range set.RouteList {
			route.RunBefore = append(route.RunBefore, Runnable{line, false})
		}
	}
	return set
}

func (set *RouteSubset) LitBefore(lines ...string) *RouteSubset {
	for _, line := range lines {
		for _, route := range set.RouteList {
			route.RunBefore = append(route.RunBefore, Runnable{line, true})
		}
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
