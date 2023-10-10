package main

func getCompPkgSet(comps []ComponentData) map[string]int {
	compPkgs := map[string]int{}

	for i, comp := range comps {
		compPkgs["\""+comp.PkgPath+"\""] = i
	}

	return compPkgs
}
