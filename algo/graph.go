package algo

// DFS 使用深度优先遍历图
func DFS(graph [][]int) []int {
	vis := make([]bool, len(graph))
	var res []int
	var dfs func(id int)
	dfs = func(id int) {
		vis[id] = true
		res = append(res, id)
		for i := 0; i < len(graph[id]); i++ {
			if graph[id][i] == 1 && !vis[i] {
				dfs(i)
			}
		}
	}
	for i := range vis {
		if !vis[i] {
			dfs(i)
		}
	}
	return res
}
