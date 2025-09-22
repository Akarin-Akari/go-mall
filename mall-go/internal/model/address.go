package model

import (
	"time"

	"gorm.io/gorm"
)

// Address 地址模型
type Address struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	ReceiverName  string         `gorm:"type:varchar(50);not null" json:"receiver_name"`      // 收货人姓名
	ReceiverPhone string         `gorm:"type:varchar(20);not null" json:"receiver_phone"`     // 收货人电话
	Province      string         `gorm:"type:varchar(50);not null" json:"province"`           // 省份
	City          string         `gorm:"type:varchar(50);not null" json:"city"`               // 城市
	District      string         `gorm:"type:varchar(50);not null" json:"district"`           // 区/县
	DetailAddress string         `gorm:"type:varchar(200);not null" json:"detail_address"`    // 详细地址
	PostalCode    string         `gorm:"type:varchar(10)" json:"postal_code"`                 // 邮政编码
	IsDefault     bool           `gorm:"default:false" json:"is_default"`                     // 是否默认地址
	AddressType   AddressType    `gorm:"type:varchar(20);default:'home'" json:"address_type"` // 地址类型
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// AddressType 地址类型
type AddressType string

const (
	AddressTypeHome    AddressType = "home"    // 家庭地址
	AddressTypeCompany AddressType = "company" // 公司地址
	AddressTypeOther   AddressType = "other"   // 其他地址
)

// TableName 指定表名
func (Address) TableName() string {
	return "addresses"
}

// BeforeCreate 创建前钩子
func (a *Address) BeforeCreate(tx *gorm.DB) error {
	// 如果设置为默认地址，需要将该用户的其他地址设为非默认
	if a.IsDefault {
		return tx.Model(&Address{}).
			Where("user_id = ? AND is_default = ?", a.UserID, true).
			Update("is_default", false).Error
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (a *Address) BeforeUpdate(tx *gorm.DB) error {
	// 如果设置为默认地址，需要将该用户的其他地址设为非默认
	if a.IsDefault {
		return tx.Model(&Address{}).
			Where("user_id = ? AND id != ? AND is_default = ?", a.UserID, a.ID, true).
			Update("is_default", false).Error
	}
	return nil
}

// GetFullAddress 获取完整地址
func (a *Address) GetFullAddress() string {
	return a.Province + a.City + a.District + a.DetailAddress
}

// AddressCreateRequest 创建地址请求
type AddressCreateRequest struct {
	ReceiverName  string      `json:"receiver_name" binding:"required,max=50" example:"张三"`
	ReceiverPhone string      `json:"receiver_phone" binding:"required,max=20" example:"13800138000"`
	Province      string      `json:"province" binding:"required,max=50" example:"广东省"`
	City          string      `json:"city" binding:"required,max=50" example:"深圳市"`
	District      string      `json:"district" binding:"required,max=50" example:"南山区"`
	DetailAddress string      `json:"detail_address" binding:"required,max=200" example:"科技园南区深南大道"`
	PostalCode    string      `json:"postal_code" binding:"omitempty,len=6" example:"518000"`
	IsDefault     bool        `json:"is_default" example:"false"`
	AddressType   AddressType `json:"address_type" binding:"omitempty,oneof=home company other" example:"home"`
}

// AddressUpdateRequest 更新地址请求
type AddressUpdateRequest struct {
	ReceiverName  string      `json:"receiver_name" binding:"omitempty,max=50"`
	ReceiverPhone string      `json:"receiver_phone" binding:"omitempty,max=20"`
	Province      string      `json:"province" binding:"omitempty,max=50"`
	City          string      `json:"city" binding:"omitempty,max=50"`
	District      string      `json:"district" binding:"omitempty,max=50"`
	DetailAddress string      `json:"detail_address" binding:"omitempty,max=200"`
	PostalCode    string      `json:"postal_code" binding:"omitempty,len=6"`
	IsDefault     *bool       `json:"is_default"` // 使用指针允许明确设置false
	AddressType   AddressType `json:"address_type" binding:"omitempty,oneof=home company other"`
}

// AddressResponse 地址响应
type AddressResponse struct {
	ID            uint        `json:"id"`
	ReceiverName  string      `json:"receiver_name"`
	ReceiverPhone string      `json:"receiver_phone"`
	Province      string      `json:"province"`
	City          string      `json:"city"`
	District      string      `json:"district"`
	DetailAddress string      `json:"detail_address"`
	PostalCode    string      `json:"postal_code"`
	IsDefault     bool        `json:"is_default"`
	AddressType   AddressType `json:"address_type"`
	FullAddress   string      `json:"full_address"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// ToResponse 转换为响应格式
func (a *Address) ToResponse() *AddressResponse {
	return &AddressResponse{
		ID:            a.ID,
		ReceiverName:  a.ReceiverName,
		ReceiverPhone: a.ReceiverPhone,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		DetailAddress: a.DetailAddress,
		PostalCode:    a.PostalCode,
		IsDefault:     a.IsDefault,
		AddressType:   a.AddressType,
		FullAddress:   a.GetFullAddress(),
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

// AddressListRequest 地址列表请求
type AddressListRequest struct {
	UserID      uint        `json:"user_id" binding:"omitempty,min=1"`
	AddressType AddressType `json:"address_type" binding:"omitempty,oneof=home company other"`
	IsDefault   *bool       `json:"is_default"`
	Page        int         `json:"page" binding:"omitempty,min=1"`
	PageSize    int         `json:"page_size" binding:"omitempty,min=1,max=100"`
}

// RegionResponse 地区响应
type RegionResponse struct {
	Code     string           `json:"code"`
	Name     string           `json:"name"`
	Level    int              `json:"level"` // 1:省份 2:城市 3:区县
	Children []RegionResponse `json:"children,omitempty"`
}

// ValidateAddressFields 验证地址字段
func ValidateAddressFields(req *AddressCreateRequest) error {
	if req.ReceiverName == "" {
		return ErrInvalidParam
	}
	if req.ReceiverPhone == "" {
		return ErrInvalidParam
	}
	if req.Province == "" {
		return ErrInvalidParam
	}
	if req.City == "" {
		return ErrInvalidParam
	}
	if req.District == "" {
		return ErrInvalidParam
	}
	if req.DetailAddress == "" {
		return ErrInvalidParam
	}
	return nil
}
