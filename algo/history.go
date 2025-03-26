package algo

import "sort"

func NextPermutation(nums []int) {
	n := len(nums)
	i := n - 2
	for i >= 0 && nums[i] >= nums[i+1] {
		i--
	}
	if i >= 0 {
		j := n - 1
		for j >= 0 && nums[i] >= nums[j] {
			j--
		}
		nums[i], nums[j] = nums[j], nums[i]
	}
	reverse(nums[i+1:])
}

func reverse(a []int) {
	for i, n := 0, len(a); i < n/2; i++ {
		a[i], a[n-1-i] = a[n-1-i], a[i]
	}
}

func Melons(watermelon []int, expire []int) int {
	// 记录每一天可用的西瓜数量
	available := make([]int, len(watermelon))
	// 记录已吃的西瓜数量
	eaten := 0

	for i := 0; i < len(watermelon); i++ {
		// 计算第 i 天可以吃的西瓜数量
		canEat := min(watermelon[i], available[i])
		eaten += canEat
		// 更新可用数量数组
		for j := i + 1; j < min(i+expire[i]+1, len(available)); j++ {
			available[j] += watermelon[i]
		}
	}

	return eaten
}

func MinMalwareSpread(graph [][]int, initial []int) int {
	n := len(graph)
	dsu := NewDSU(n)
	// 合并所有的感染节点
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if graph[i][j] == 1 {
				dsu.Union(i, j)
			}
		}
	}

	// 统计每个连通分量中的初始感染节点数量
	count := make([]int, n)
	for _, node := range initial {
		count[dsu.Find(node)]++
	}

	// 统计每个连通分量中的节点数量
	ans := -1
	for _, node := range initial {
		root := dsu.Find(node)
		if count[root] == 1 {
			if ans == -1 || dsu.Size(root) > dsu.Size(dsu.Find(ans)) {
				ans = node
			} else if dsu.Size(root) == dsu.Size(dsu.Find(ans)) && node < ans {
				ans = node
			}
		}
	}

	// 如果没有只感染一个节点的节点，返回最小的节点
	if ans == -1 {
		ans = initial[0]
		for _, node := range initial {
			if node < ans {
				ans = node
			}
		}
	}

	return ans
}

func JobScheduling(startTime []int, endTime []int, profit []int) int {
	// 按照endTime排序
	indices := make([]int, len(endTime))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return endTime[indices[i]] < endTime[indices[j]]
	})

	sortedStartTime := make([]int, len(startTime))
	sortedEndTime := make([]int, len(endTime))
	sortedProfit := make([]int, len(profit))

	for i, index := range indices {
		sortedStartTime[i] = startTime[index]
		sortedEndTime[i] = endTime[index]
		sortedProfit[i] = profit[index]
	}

	dp := make([]int, len(sortedStartTime))
	dp[0] = sortedProfit[0]
	for i := 1; i < len(sortedStartTime); i++ {
		if findIndex(sortedEndTime, sortedStartTime[i]) != -1 {
			temp := findIndex(sortedEndTime, sortedStartTime[i])
			dp[i] = max(dp[i-1], dp[temp]+sortedProfit[i])
		} else {
			dp[i] = max(dp[i-1], sortedProfit[i])
		}
	}
	return dp[len(sortedStartTime)-1]
}
