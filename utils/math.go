package utils

type UnsignedInteger interface {
	byte | uint | uint16 | uint32 | uint64
}

type SignedInteger interface {
	int | int16 | int32 | int64
}

type Float interface {
	float32 | float64
}

type Number interface {
	UnsignedInteger | SignedInteger | Float
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func MinOf[T Number](nums ...T) T {
	var sample T
	ind := 0
	for i := range nums {
		if i == 0 {
			sample = nums[i]
			continue
		} else if nums[i] < sample {
			sample = nums[i]
			ind = i
		}
	}
	return nums[ind]
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func MaxOf[T Number](nums ...T) T {
	var sample T
	ind := 0
	for i := range nums {
		if i == 0 {
			sample = nums[i]
			continue
		} else if nums[i] > sample {
			sample = nums[i]
			ind = i
		}
	}
	return nums[ind]
}
