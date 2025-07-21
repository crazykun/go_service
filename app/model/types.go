package model

// ServiceListRequest 服务列表请求参数
type ServiceListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Name     string `json:"name" form:"name"`
	Status   *int   `json:"status" form:"status"` // 使用指针以区分0值和未设置
}

// ServiceListResponse 服务列表响应
type ServiceListResponse struct {
	List  []ServiceStatusModel `json:"list"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	ServiceIds []int64 `json:"service_ids" binding:"required"`
	Operation  string  `json:"operation" binding:"required,oneof=start stop restart kill"`
}

// BatchOperationResponse 批量操作响应
type BatchOperationResponse struct {
	TotalCount   int                      `json:"total_count"`
	SuccessCount int                      `json:"success_count"`
	FailedCount  int                      `json:"failed_count"`
	Results      []map[string]interface{} `json:"results"`
}

// LogListRequest 日志列表请求参数
type LogListRequest struct {
	Page      int    `json:"page" form:"page"`
	PageSize  int    `json:"page_size" form:"page_size"`
	ServiceId *int64 `json:"service_id" form:"service_id"`
	Operation string `json:"operation" form:"operation"`
	Status    string `json:"status" form:"status"`
}

// LogListResponse 日志列表响应
type LogListResponse struct {
	List  []ServiceLog `json:"list"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}
