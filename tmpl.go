package main

func getCompPkgSet(comps []ComponentData) map[string]int {
	compPkgs := map[string]int{}
	count := 0

	for _, comp := range comps {
		compPkg := "\"" + comp.PkgPath + "\""

		if _, ok := compPkgs[compPkg]; !ok {
			compPkgs[compPkg] = count
			count++
		}
	}

	return compPkgs
}
