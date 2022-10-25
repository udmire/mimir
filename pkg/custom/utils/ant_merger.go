package utils

func MergeTenants(expects, permitted []string) []string {
	var merged []string
	if len(expects) == 0 || len(permitted) == 0 {
		return merged
	}

	expects = normalize(expects)
	permitted = normalize(permitted)

	if len(expects) == 1 && expects[0] == "*" {
		return permitted
	} else if len(permitted) == 1 && permitted[0] == "*" {
		return expects
	} else {
		for _, expect := range expects {
			for _, p := range permitted {
				if expect == p {
					merged = append(merged, expect)
				}
			}
		}
	}
	return merged
}

func normalize(tenants []string) []string {
	var result []string
	for _, policy := range tenants {
		if policy == "*" {
			return []string{"*"}
		} else {
			result = append(result, policy)
		}
	}
	return result
}
