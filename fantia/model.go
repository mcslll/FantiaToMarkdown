package fantia

// Post 表示帖子的基本信息和内容
type Post struct {
	Title    string   `json:"title"`
	Url      string   `json:"url"`
	Content  string   `json:"content"`  // HTML 内容
	Pictures []string `json:"pictures"` // 图片 URL 列表
}

// Cookie 结构体用于解析 cookies.json
type Cookie struct {
	Name  string `json:"name"  binding:"required"`
	Value string `json:"value" binding:"required"`
}
