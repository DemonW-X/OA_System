package models

import (
	"time"

	"gorm.io/gorm"
)

// Onboarding 入职管理
type Onboarding struct {
	ID            int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	EmployeeName  string         `json:"employee_name"`  // 员工姓名（文本填写）
	OnboardDate   time.Time      `json:"onboard_date"`
	OnboardType   string         `json:"onboard_type" gorm:"default:'new'"` // new=新员工 rehire=返聘 transfer=调入
	ProbationDays int            `json:"probation_days" gorm:"default:90"`
	ProbationEnd  *time.Time     `json:"probation_end"`
	// 个人详细信息
	IDCard        string         `json:"id_card"`        // 身份证号
	Phone         string         `json:"phone"`          // 联系电话
	Email         string         `json:"email"`          // 电子邮箱
	NativePlace   string         `json:"native_place"`   // 籍贯
	Address       string         `json:"address"`        // 现居住地址
	EmergencyName string         `json:"emergency_name"` // 紧急联系人
	EmergencyPhone string        `json:"emergency_phone"` // 紧急联系电话
	Education     string         `json:"education"`      // 学历：junior/high/college/bachelor/master/doctor
	School        string         `json:"school"`         // 毕业院校
	Major         string         `json:"major"`          // 专业
	WorkYears     int            `json:"work_years"`     // 工作年限
	Remark        string         `json:"remark"`
	ApproveStatus string         `json:"approve_status" gorm:"default:'draft'"` // draft/pending/approved/rejected
	ApprovedBy    string         `json:"approved_by"`
	ApprovedAt    *time.Time     `json:"approved_at"`
	ApproveRemark string         `json:"approve_remark"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
