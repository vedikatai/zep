package embeddings

// MaximalMarginalRelevance selects diverse, relevant indices using cosine similarity
// on float16-stored vectors (promoted to float32 for the kernel only).
func MaximalMarginalRelevance(query Vector, list []Vector, lambdaMult float32, k int) ([]int, error) {
	if k <= 0 || len(list) == 0 {
		return []int{}, nil
	}
	if len(query) == 0 {
		return nil, ErrEmptyVector
	}
	if len(query) != len(list[0]) {
		return nil, ErrDimMismatch
	}
	simQuery := make([]float32, len(list))
	for i, v := range list {
		s, err := CosineSimilarity(query, v)
		if err != nil {
			return nil, err
		}
		simQuery[i] = s
	}
	best := 0
	for i := 1; i < len(simQuery); i++ {
		if simQuery[i] > simQuery[best] {
			best = i
		}
	}
	idxs := []int{best}
	selected := []Vector{list[best]}

	for len(idxs) < k && len(idxs) < len(list) {
		var bestScore float32 = -1e30
		idxToAdd := -1
		for i, qScore := range simQuery {
			already := false
			for _, j := range idxs {
				if j == i {
					already = true
					break
				}
			}
			if already {
				continue
			}
			var redun float32 = -1e30
			for _, s := range selected {
				sc, err := CosineSimilarity(list[i], s)
				if err != nil {
					return nil, err
				}
				if sc > redun {
					redun = sc
				}
			}
			score := lambdaMult*qScore - (1-lambdaMult)*redun
			if score > bestScore {
				bestScore = score
				idxToAdd = i
			}
		}
		if idxToAdd < 0 {
			break
		}
		idxs = append(idxs, idxToAdd)
		selected = append(selected, list[idxToAdd])
	}
	return idxs, nil
}
