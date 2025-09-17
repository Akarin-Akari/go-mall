package product

import (
	"fmt"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// CategoryService 商品分类服务
type CategoryService struct {
	db *gorm.DB
}

// NewCategoryService 创建商品分类服务
func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{
		db: db,
	}
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=100"`
	Description    string `json:"description"`
	ParentID       uint   `json:"parent_id"`
	Icon           string `json:"icon"`
	Image          string `json:"image"`
	Sort           int    `json:"sort"`
	SEOTitle       string `json:"seo_title"`
	SEOKeywords    string `json:"seo_keywords"`
	SEODescription string `json:"seo_description"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=100"`
	Description    string `json:"description"`
	ParentID       uint   `json:"parent_id"`
	Icon           string `json:"icon"`
	Image          string `json:"image"`
	Sort           int    `json:"sort"`
	Status         string `json:"status"`
	SEOTitle       string `json:"seo_title"`
	SEOKeywords    string `json:"seo_keywords"`
	SEODescription string `json:"seo_description"`
}

// CategoryListRequest 分类列表请求
type CategoryListRequest struct {
	ParentID *uint  `form:"parent_id"`
	Level    *int   `form:"level"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}

// CategoryTreeNode 分类树节点
type CategoryTreeNode struct {
	*model.Category
	Children []*CategoryTreeNode `json:"children"`
}

// CreateCategory 创建分类
func (cs *CategoryService) CreateCategory(req *CreateCategoryRequest) (*model.Category, error) {
	// 检查分类名称是否已存在
	var existingCategory model.Category
	if err := cs.db.Where("name = ? AND parent_id = ?", req.Name, req.ParentID).First(&existingCategory).Error; err == nil {
		return nil, fmt.Errorf("同级分类名称已存在")
	}

	// 计算分类层级和路径
	level := 1
	path := req.Name

	if req.ParentID > 0 {
		var parent model.Category
		if err := cs.db.First(&parent, req.ParentID).Error; err != nil {
			return nil, fmt.Errorf("父分类不存在")
		}
		level = parent.Level + 1
		path = parent.Path + "/" + req.Name
	}

	// 创建分类
	category := &model.Category{
		Name:           req.Name,
		Description:    req.Description,
		ParentID:       req.ParentID,
		Level:          level,
		Path:           path,
		Icon:           req.Icon,
		Image:          req.Image,
		Sort:           req.Sort,
		Status:         model.CategoryStatusActive,
		SEOTitle:       req.SEOTitle,
		SEOKeywords:    req.SEOKeywords,
		SEODescription: req.SEODescription,
	}

	if err := cs.db.Create(category).Error; err != nil {
		return nil, fmt.Errorf("创建分类失败: %v", err)
	}

	return category, nil
}

// UpdateCategory 更新分类
func (cs *CategoryService) UpdateCategory(id uint, req *UpdateCategoryRequest) (*model.Category, error) {
	var category model.Category
	if err := cs.db.First(&category, id).Error; err != nil {
		return nil, fmt.Errorf("分类不存在")
	}

	// 检查是否修改了父分类
	if req.ParentID != category.ParentID {
		// 不能将分类设置为自己的子分类
		if req.ParentID == category.ID {
			return nil, fmt.Errorf("不能将分类设置为自己的子分类")
		}

		// 检查是否会形成循环引用
		if err := cs.checkCircularReference(category.ID, req.ParentID); err != nil {
			return nil, err
		}

		// 重新计算层级和路径
		level := 1
		path := req.Name

		if req.ParentID > 0 {
			var parent model.Category
			if err := cs.db.First(&parent, req.ParentID).Error; err != nil {
				return nil, fmt.Errorf("父分类不存在")
			}
			level = parent.Level + 1
			path = parent.Path + "/" + req.Name
		}

		category.Level = level
		category.Path = path
		category.ParentID = req.ParentID

		// 更新所有子分类的层级和路径
		if err := cs.updateChildrenPath(&category); err != nil {
			return nil, fmt.Errorf("更新子分类路径失败: %v", err)
		}
	}

	// 更新分类信息
	category.Name = req.Name
	category.Description = req.Description
	category.Icon = req.Icon
	category.Image = req.Image
	category.Sort = req.Sort
	category.SEOTitle = req.SEOTitle
	category.SEOKeywords = req.SEOKeywords
	category.SEODescription = req.SEODescription

	if req.Status != "" {
		category.Status = req.Status
	}

	if err := cs.db.Save(&category).Error; err != nil {
		return nil, fmt.Errorf("更新分类失败: %v", err)
	}

	return &category, nil
}

// DeleteCategory 删除分类
func (cs *CategoryService) DeleteCategory(id uint) error {
	var category model.Category
	if err := cs.db.First(&category, id).Error; err != nil {
		return fmt.Errorf("分类不存在")
	}

	// 检查是否有子分类
	var childCount int64
	if err := cs.db.Model(&model.Category{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return fmt.Errorf("检查子分类失败: %v", err)
	}
	if childCount > 0 {
		return fmt.Errorf("存在子分类，无法删除")
	}

	// 检查是否有商品使用此分类
	var productCount int64
	if err := cs.db.Model(&model.Product{}).Where("category_id = ?", id).Count(&productCount).Error; err != nil {
		return fmt.Errorf("检查商品失败: %v", err)
	}
	if productCount > 0 {
		return fmt.Errorf("存在商品使用此分类，无法删除")
	}

	// 删除分类
	if err := cs.db.Delete(&category).Error; err != nil {
		return fmt.Errorf("删除分类失败: %v", err)
	}

	return nil
}

// GetCategory 获取分类详情
func (cs *CategoryService) GetCategory(id uint) (*model.Category, error) {
	var category model.Category
	if err := cs.db.Preload("Parent").Preload("Children").First(&category, id).Error; err != nil {
		return nil, fmt.Errorf("分类不存在")
	}

	return &category, nil
}

// GetCategoryList 获取分类列表
func (cs *CategoryService) GetCategoryList(req *CategoryListRequest) ([]*model.Category, int64, error) {
	query := cs.db.Model(&model.Category{})

	// 条件筛选
	if req.ParentID != nil {
		query = query.Where("parent_id = ?", *req.ParentID)
	}

	if req.Level != nil {
		query = query.Where("level = ?", *req.Level)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分类总数失败: %v", err)
	}

	// 分页查询
	var categories []*model.Category
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("sort ASC, id ASC").Offset(offset).Limit(req.PageSize).Find(&categories).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分类列表失败: %v", err)
	}

	return categories, total, nil
}

// GetCategoryTree 获取分类树
func (cs *CategoryService) GetCategoryTree(parentID uint) ([]*CategoryTreeNode, error) {
	var categories []model.Category
	query := cs.db.Where("status = ?", model.CategoryStatusActive)

	if parentID > 0 {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id = 0")
	}

	if err := query.Order("sort ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询分类失败: %v", err)
	}

	var nodes []*CategoryTreeNode
	for _, category := range categories {
		node := &CategoryTreeNode{
			Category: &category,
		}

		// 递归获取子分类
		children, err := cs.GetCategoryTree(category.ID)
		if err != nil {
			return nil, err
		}
		node.Children = children

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetAllCategoryTree 获取完整分类树
func (cs *CategoryService) GetAllCategoryTree() ([]*CategoryTreeNode, error) {
	// 获取所有分类
	var categories []model.Category
	if err := cs.db.Where("status = ?", model.CategoryStatusActive).Order("sort ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询分类失败: %v", err)
	}

	// 构建分类映射
	categoryMap := make(map[uint]*CategoryTreeNode)
	var rootNodes []*CategoryTreeNode

	// 创建所有节点
	for _, category := range categories {
		node := &CategoryTreeNode{
			Category: &category,
			Children: []*CategoryTreeNode{},
		}
		categoryMap[category.ID] = node
	}

	// 构建树结构
	for _, category := range categories {
		node := categoryMap[category.ID]
		if category.ParentID == 0 {
			rootNodes = append(rootNodes, node)
		} else {
			if parent, exists := categoryMap[category.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return rootNodes, nil
}

// GetCategoryPath 获取分类路径
func (cs *CategoryService) GetCategoryPath(id uint) ([]*model.Category, error) {
	var category model.Category
	if err := cs.db.First(&category, id).Error; err != nil {
		return nil, fmt.Errorf("分类不存在")
	}

	var path []*model.Category
	current := &category

	for current != nil {
		path = append([]*model.Category{current}, path...)

		if current.ParentID == 0 {
			break
		}

		var parent model.Category
		if err := cs.db.First(&parent, current.ParentID).Error; err != nil {
			break
		}
		current = &parent
	}

	return path, nil
}

// checkCircularReference 检查循环引用
func (cs *CategoryService) checkCircularReference(categoryID, parentID uint) error {
	if parentID == 0 {
		return nil
	}

	// 获取父分类的所有祖先分类
	ancestors, err := cs.getAncestors(parentID)
	if err != nil {
		return err
	}

	// 检查是否包含当前分类
	for _, ancestor := range ancestors {
		if ancestor.ID == categoryID {
			return fmt.Errorf("不能形成循环引用")
		}
	}

	return nil
}

// getAncestors 获取祖先分类
func (cs *CategoryService) getAncestors(categoryID uint) ([]*model.Category, error) {
	var ancestors []*model.Category
	currentID := categoryID

	for currentID != 0 {
		var category model.Category
		if err := cs.db.First(&category, currentID).Error; err != nil {
			break
		}

		ancestors = append(ancestors, &category)
		currentID = category.ParentID
	}

	return ancestors, nil
}

// updateChildrenPath 更新子分类路径
func (cs *CategoryService) updateChildrenPath(parent *model.Category) error {
	var children []model.Category
	if err := cs.db.Where("parent_id = ?", parent.ID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		child.Level = parent.Level + 1
		child.Path = parent.Path + "/" + child.Name

		if err := cs.db.Save(&child).Error; err != nil {
			return err
		}

		// 递归更新子分类的子分类
		if err := cs.updateChildrenPath(&child); err != nil {
			return err
		}
	}

	return nil
}

// 全局分类服务实例
var globalCategoryService *CategoryService

// InitGlobalCategoryService 初始化全局分类服务
func InitGlobalCategoryService(db *gorm.DB) {
	globalCategoryService = NewCategoryService(db)
}

// GetGlobalCategoryService 获取全局分类服务
func GetGlobalCategoryService() *CategoryService {
	return globalCategoryService
}
