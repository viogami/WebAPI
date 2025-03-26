package algo

type DSU struct {
	parent []int
	size   []int
}

func NewDSU(n int) *DSU {
	parent := make([]int, n)
	size := make([]int, n)
	for i := 0; i < n; i++ {
		parent[i] = i
		size[i] = 1
	}
	return &DSU{parent, size}
}

// 查找根节点
func (dsu *DSU) Find(x int) int {
	if dsu.parent[x] != x {
		dsu.parent[x] = dsu.Find(dsu.parent[x])
	}
	return dsu.parent[x]
}

// 合并两个节点
func (dsu *DSU) Union(x, y int) {
	xr := dsu.Find(x)
	yr := dsu.Find(y)
	if xr != yr {
		dsu.parent[xr] = yr
		dsu.size[yr] += dsu.size[xr]
	}
}

// 获取连通分量的大小
func (dsu *DSU) Size(x int) int {
	return dsu.size[dsu.Find(x)]
}
