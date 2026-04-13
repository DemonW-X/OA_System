package dto

// DataDictionaryRequestDTO 数据字典主表请求
type DataDictionaryRequestDTO struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
}

// DataDictionaryItemRequestDTO 数据字典子表请求
type DataDictionaryItemRequestDTO struct {
	ExtField string `json:"extfield"`
	Remark   string `json:"remark"`
}
