package progress

// ProgressPercentage calculates the percentage that should be shown in
// a progress bar when k of n work items are done.
// A return value of 0 percent means that nothing has been done.
// 1 means that something has been done, but less than 2 %.
// (This is the only case in which the displayed progress is higher than
// the actual progress. It is different from 0 though, to show that
// something has already been done.) All other values are rounded down,
// therefore 100 means that everything has been done.
func ProgressPercentage(k, n int) int {
	var perc int
	switch {
	case k <= 0 || n <= 0:
		perc = 0
	case k >= n:
		perc = 100
	default:
		if k100 := 100 * k; k100/100 == k {
			perc = k100 / n
		} else {
			perc = k / (n / 100)
		}
		if perc == 0 {
			perc = 1
		}
	}
	return perc
}
