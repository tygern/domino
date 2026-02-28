package coxeter

import "sync"

func AllElements(rank int) []Element {
	var result []Element
	perm := make([]int, rank)
	used := make([]bool, rank+1)
	allElementsBacktrack(perm, used, 0, rank, 0, &result)
	return result
}

func allElementsBacktrack(perm []int, used []bool, pos, rank, negCount int, result *[]Element) {
	if pos == rank {
		if negCount%2 != 0 {
			return
		}
		cp := make([]int, rank)
		copy(cp, perm)
		*result = append(*result, Element{perm: cp})
		return
	}

	for absVal := 1; absVal <= rank; absVal++ {
		if used[absVal] {
			continue
		}
		used[absVal] = true

		for _, sign := range []int{1, -1} {
			newNeg := negCount
			if sign < 0 {
				newNeg++
			}
			if pos == rank-1 && newNeg%2 != 0 {
				continue
			}

			perm[pos] = sign * absVal
			allElementsBacktrack(perm, used, pos+1, rank, newNeg, result)
		}

		used[absVal] = false
	}
}

func BadElements(rank int) []Element {
	if rank < 4 {
		var result []Element
		for _, e := range AllElements(rank) {
			if e.IsBad() {
				result = append(result, e)
			}
		}
		return result
	}

	type work struct {
		absVal int
		sign   int
	}

	var tasks []work
	for absVal := 1; absVal <= rank; absVal++ {
		for _, sign := range []int{1, -1} {
			tasks = append(tasks, work{absVal, sign})
		}
	}

	results := make([][]Element, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, w work) {
			defer wg.Done()
			perm := make([]int, rank)
			inv := make([]int, rank)
			used := make([]bool, rank+1)

			startVal := w.sign * w.absVal
			perm[0] = startVal
			used[w.absVal] = true
			inv[w.absVal-1] = w.sign * 1
			negCount := 0
			if w.sign < 0 {
				negCount = 1
			}

			var local []Element
			badElementsBacktrack(perm, inv, used, 1, rank, negCount, &local)
			results[idx] = local
		}(i, task)
	}

	wg.Wait()

	var combined []Element
	for _, r := range results {
		combined = append(combined, r...)
	}
	return combined
}

func badElementsBacktrack(perm, inv []int, used []bool, pos, rank, negCount int, result *[]Element) {
	if pos == rank {
		if negCount%2 != 0 {
			return
		}
		elem := Element{perm: make([]int, rank)}
		copy(elem.perm, perm)
		if elem.IsBad() {
			*result = append(*result, elem)
		}
		return
	}

	for absVal := 1; absVal <= rank; absVal++ {
		if used[absVal] {
			continue
		}
		used[absVal] = true

		for _, sign := range []int{1, -1} {
			newNeg := negCount
			if sign < 0 {
				newNeg++
			}
			if pos == rank-1 && newNeg%2 != 0 {
				continue
			}

			perm[pos] = sign * absVal
			inv[absVal-1] = sign * (pos + 1)

			if canBeRightBadPartial(perm, pos, rank) && canBeRightBadPartial(inv, absVal-1, rank) {
				badElementsBacktrack(perm, inv, used, pos+1, rank, newNeg, result)
			}

			perm[pos] = 0
			inv[absVal-1] = 0
		}

		used[absVal] = false
	}
}

func canBeRightBadPartial(perm []int, filledPos, rank int) bool {
	if rank >= 3 && perm[0] != 0 && perm[2] != 0 {
		if -perm[0] > perm[2] {
			return false
		}
	}

	for j := 0; j+2 < rank; j++ {
		if perm[j] != 0 && perm[j+2] != 0 && perm[j] > perm[j+2] {
			return false
		}
	}

	return true
}
