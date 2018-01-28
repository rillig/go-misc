package progress

import "testing"

func TestProgressPercentage(t *testing.T) {
	var n int

	check := func(k, percentage int) {
		actual := ProgressPercentage(k, n)
		if actual != percentage {
			t.Errorf("Progress %d/%d should be %d %%, is %d %%.", k, n, percentage, actual)
		}
	}

	n = 7

	check(-100, 0)
	check(0, 0)
	check(1, 14)
	check(2, 28)
	check(3, 42)
	check(4, 57)
	check(5, 71)
	check(6, 85)
	check(7, 100)
	check(8, 100)

	n = 1000

	check(0, 0)
	check(1, 1)
	check(10, 1)
	check(19, 1)
	check(20, 2)
	check(29, 2)
	check(30, 3)
	check(623, 62)
	check(984, 98)
	check(985, 98)
	check(999, 99)
	check(1000, 100)

	n = int(^uint(0) >> 1)

	check(0, 0)
	check(n/3, 33)
	check(n/2, 50)
	check(n, 100)
}
