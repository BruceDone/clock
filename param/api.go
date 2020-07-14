package param

// 请求
type (
	Page struct {
		Count int `query:"count" json:"count"`
		Index int `query:"index" json:"index"`
		Total int `json:"total"`
	}

	// 用户信息
	User struct {
		UserName string `json:"user_name"`
		UserPwd  string `json:"user_pwd"`
	}

	// 关系节点查询
	NodeQuery struct {
		Page
		KeyWord string `query:"keyword" json:"keyword"` // 查询关键字
		TaskID  int    `query:"task_id" json:"task_id"`
	}
)

// 返回
type (
	ApiResponse struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}

	// 分页返回的请求体
	ListResponse struct {
		Items     interface{} `json:"items"`
		PageQuery interface{} `json:"page"`
	}

	RelationResponse struct {
		Nodes interface{} `json:"nodes"`
		Links interface{} `json:"links"`
	}
)
