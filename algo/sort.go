package algo

// 快速排序函数
func QuickSort(nums []int) {
	if len(nums) <= 1 {
		return
	}
	piviot := nums[0]
	left, right := 0, len(nums)-1
	for left < right {
		for nums[left] < piviot {
			left++
		}
		for nums[right] > piviot {
			right--
		}
		if left <= right {
			nums[left], nums[right] = nums[right], nums[left]
			left++
			right--
		}
	}
	QuickSort(nums[:right+1])
	QuickSort(nums[left:])
}
