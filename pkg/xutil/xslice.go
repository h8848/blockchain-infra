package xutil

// 切片差集
func FindDifference(slice1 []string, slice2 []string) []string {
	var difference []string
	// 遍历第一个切片
	for _, elem1 := range slice1 {
		found := false
		// 遍历第二个切片，查找是否存在相同的元素
		for _, elem2 := range slice2 {
			if elem1 == elem2 {
				found = true
				break
			}
		}
		// 如果在第二个切片中找不到相同的元素，将其添加到差异切片
		if !found {
			difference = append(difference, elem1)
		}
	}
	return difference
}

func InArray[T comparable](element T, array []T) bool {
	for _, v := range array {
		if v == element {
			return true
		}
	}
	return false
}
