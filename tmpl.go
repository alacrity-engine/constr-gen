package main

func getCompPkgSet(comps []ComponentData) map[string]struct{} {
	compPkgs := map[string]struct{}{}

	for _, comp := range comps {
		compPkgs["\""+comp.PkgPath+"\""] = struct{}{}
	}

	return compPkgs
}
