package coxeter

import (
	"math/bits"
	"runtime"
	"sync"
)

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

type badWorkItem struct {
	perm     []int
	inv      []int
	used     uint64
	negCount int
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

	items := generateBadWorkItems(rank)

	ch := make(chan int, len(items))
	for i := range items {
		ch <- i
	}
	close(ch)

	numWorkers := runtime.NumCPU()
	results := make([][]Element, numWorkers)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var local []Element
			for i := range ch {
				item := &items[i]
				badElementsSearch(item.perm, item.inv, item.used, 2, rank, item.negCount, &local)
			}
			results[idx] = local
		}(w)
	}
	wg.Wait()

	var combined []Element
	for _, r := range results {
		combined = append(combined, r...)
	}
	return combined
}

func generateBadWorkItems(rank int) []badWorkItem {
	var items []badWorkItem

	numEvenAbove := 0
	for p := 2; p < rank; p += 2 {
		numEvenAbove++
	}

	for absVal0 := 1; absVal0 <= rank; absVal0++ {
		if rank-absVal0 < numEvenAbove {
			continue
		}

		signs0 := []int{1, -1}
		if absVal0 >= 3 && absVal0%2 == 1 {
			signs0 = []int{1}
		}

		for _, sign0 := range signs0 {
			negCount0 := 0
			if sign0 < 0 {
				negCount0 = 1
			}

			for absVal1 := 1; absVal1 <= rank; absVal1++ {
				if absVal1 == absVal0 {
					continue
				}

				signs1 := []int{1, -1}
				if absVal1 >= 3 && absVal1%2 == 1 {
					signs1 = []int{1}
				}

				for _, sign1 := range signs1 {
					negCount1 := negCount0
					if sign1 < 0 {
						negCount1++
					}

					perm := make([]int, rank)
					inv := make([]int, rank)
					var used uint64

					perm[0] = sign0 * absVal0
					perm[1] = sign1 * absVal1
					inv[absVal0-1] = sign0 * 1
					inv[absVal1-1] = sign1 * 2
					used |= 1 << uint(absVal0)
					used |= 1 << uint(absVal1)

					if !checkInvPlacement(inv, absVal0, rank) || !checkInvPlacement(inv, absVal1, rank) {
						continue
					}

					items = append(items, badWorkItem{perm: perm, inv: inv, used: used, negCount: negCount1})
				}
			}
		}
	}

	return items
}

func badElementsSearch(perm, inv []int, used uint64, pos, rank, negCount int, result *[]Element) {
	if pos == rank {
		if negCount%2 != 0 {
			return
		}
		if !isCommutingProductSlice(perm) {
			cp := make([]int, rank)
			copy(cp, perm)
			*result = append(*result, Element{perm: cp})
		}
		return
	}

	if pos%2 == 0 {
		var minAbsVal int
		if pos == 2 {
			minAbsVal = abs(perm[0]) + 1
		} else {
			minAbsVal = perm[pos-2] + 1
		}

		remainingEven := (rank - pos - 1) / 2

		for absVal := minAbsVal; absVal <= rank; absVal++ {
			if used&(1<<uint(absVal)) != 0 {
				continue
			}

			if remainingEven > 0 {
				unused := ^used & ((1 << uint(rank+1)) - 1)
				available := bits.OnesCount64(unused >> uint(absVal+1))
				if available < remainingEven {
					break
				}
			}

			perm[pos] = absVal
			inv[absVal-1] = pos + 1
			used |= 1 << uint(absVal)

			if checkInvPlacement(inv, absVal, rank) {
				badElementsSearch(perm, inv, used, pos+1, rank, negCount, result)
			}

			perm[pos] = 0
			inv[absVal-1] = 0
			used &^= 1 << uint(absVal)
		}
	} else {
		minSigned := perm[pos-2] + 1
		isLast := pos == rank-1

		if !isLast || negCount%2 == 0 {
			startPos := 1
			if minSigned > 1 {
				startPos = minSigned
			}
			for absVal := startPos; absVal <= rank; absVal++ {
				if used&(1<<uint(absVal)) != 0 {
					continue
				}

				perm[pos] = absVal
				inv[absVal-1] = pos + 1
				used |= 1 << uint(absVal)

				if checkInvPlacement(inv, absVal, rank) {
					badElementsSearch(perm, inv, used, pos+1, rank, negCount, result)
				}

				perm[pos] = 0
				inv[absVal-1] = 0
				used &^= 1 << uint(absVal)
			}
		}

		if minSigned <= 0 {
			newNeg := negCount + 1
			if !isLast || newNeg%2 == 0 {
				maxNegAbs := -minSigned
				if maxNegAbs > rank {
					maxNegAbs = rank
				}
				for absVal := 1; absVal <= maxNegAbs; absVal++ {
					if absVal >= 3 && absVal&1 == 1 {
						continue
					}
					if used&(1<<uint(absVal)) != 0 {
						continue
					}

					perm[pos] = -absVal
					inv[absVal-1] = -(pos + 1)
					used |= 1 << uint(absVal)

					if checkInvPlacement(inv, absVal, rank) {
						badElementsSearch(perm, inv, used, pos+1, rank, newNeg, result)
					}

					perm[pos] = 0
					inv[absVal-1] = 0
					used &^= 1 << uint(absVal)
				}
			}
		}
	}
}

func checkInvPlacement(inv []int, absVal, rank int) bool {
	k := absVal - 1

	if k >= 2 && inv[k-2] != 0 && inv[k] <= inv[k-2] {
		return false
	}

	if k+2 < rank && inv[k+2] != 0 && inv[k] >= inv[k+2] {
		return false
	}

	if k == 0 && rank >= 3 && inv[2] != 0 && -inv[0] > inv[2] {
		return false
	}
	if k == 2 && inv[0] != 0 && -inv[0] > inv[2] {
		return false
	}

	return true
}

func isCommutingProductSlice(perm []int) bool {
	n := len(perm)
	j := 0

	switch {
	case perm[0] == 1:
		j = 1
	case perm[0] == -1:
		if n < 2 || perm[1] != -2 {
			return false
		}
		j = 2
	case perm[0] == 2:
		if n < 2 || perm[1] != 1 {
			return false
		}
		j = 2
	case perm[0] == -2:
		if n < 2 || perm[1] != -1 {
			return false
		}
		j = 2
	default:
		return false
	}

	for j < n-2 {
		if perm[j] > j+2 {
			return false
		} else if perm[j] == j+1 {
			j++
		} else {
			if perm[j+1] != j+1 {
				return false
			}
			j += 2
		}
	}

	return true
}
