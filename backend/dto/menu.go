package dto

type MenuRequestDTO struct {
	Name     string `json:"name" binding:"required"`
	Icon     string `json:"icon"`
	Path     string `json:"path"`
	SortCode int    `json:"sort_code"`
	ParentID int    `json:"parent_id"`
	Visible  *bool  `json:"visible"`
	Remark   string `json:"remark"`
	BizCode  string `json:"biz_code"`
	BizName  string `json:"biz_name"`
	BizSort  int    `json:"biz_sort"`
}

type MenuTreeItemDTO struct {
	ID       int               `json:"id"`
	Name     string            `json:"name"`
	Icon     string            `json:"icon"`
	Path     string            `json:"path"`
	SortCode int               `json:"sort_code"`
	ParentID int               `json:"parent_id"`
	Visible  bool              `json:"visible"`
	Remark   string            `json:"remark"`
	BizCode  string            `json:"biz_code"`
	BizName  string            `json:"biz_name"`
	BizSort  int               `json:"biz_sort"`
	Children []MenuTreeItemDTO `json:"children"`
}
