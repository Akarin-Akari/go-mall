package product

import (
	"fmt"
	"path/filepath"
	"strings"

	"mall-go/internal/model"
	"mall-go/pkg/upload"

	"gorm.io/gorm"
)

// ImageService 商品图片服务
type ImageService struct {
	db          *gorm.DB
	fileManager *upload.FileManager
}

// NewImageService 创建商品图片服务
func NewImageService(db *gorm.DB, fileManager *upload.FileManager) *ImageService {
	return &ImageService{
		db:          db,
		fileManager: fileManager,
	}
}

// UploadImageRequest 上传图片请求
type UploadImageRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	FileID    uint   `json:"file_id" binding:"required"`
	Alt       string `json:"alt"`
	Sort      int    `json:"sort"`
	IsMain    bool   `json:"is_main"`
}

// UpdateImageRequest 更新图片请求
type UpdateImageRequest struct {
	Alt    string `json:"alt"`
	Sort   int    `json:"sort"`
	IsMain bool   `json:"is_main"`
}

// BatchUploadImageRequest 批量上传图片请求
type BatchUploadImageRequest struct {
	ProductID uint     `json:"product_id" binding:"required"`
	FileIDs   []uint   `json:"file_ids" binding:"required,min=1"`
	Alts      []string `json:"alts"`
}

// ImageListRequest 图片列表请求
type ImageListRequest struct {
	ProductID uint `form:"product_id" binding:"required"`
	Page      int  `form:"page" binding:"min=1"`
	PageSize  int  `form:"page_size" binding:"min=1,max=100"`
}

// UploadProductImage 上传商品图片
func (is *ImageService) UploadProductImage(req *UploadImageRequest) (*model.ProductImage, error) {
	// 验证商品是否存在
	var product model.Product
	if err := is.db.First(&product, req.ProductID).Error; err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	// 验证文件是否存在
	fileInfo, err := is.fileManager.GetFileInfo(req.FileID)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %v", err)
	}

	// 验证文件类型
	if !is.isImageFile(fileInfo.OriginalName) {
		return nil, fmt.Errorf("只能上传图片文件")
	}

	// 如果设置为主图，需要将其他图片的主图状态取消
	if req.IsMain {
		if err := is.db.Model(&model.ProductImage{}).
			Where("product_id = ?", req.ProductID).
			Update("is_main", false).Error; err != nil {
			return nil, fmt.Errorf("更新主图状态失败: %v", err)
		}
	}

	// 创建商品图片记录
	image := &model.ProductImage{
		ProductID: req.ProductID,
		URL:       fileInfo.URL(),
		Alt:       req.Alt,
		Sort:      req.Sort,
		IsMain:    req.IsMain,
	}

	if err := is.db.Create(image).Error; err != nil {
		return nil, fmt.Errorf("创建商品图片失败: %v", err)
	}

	return image, nil
}

// BatchUploadProductImages 批量上传商品图片
func (is *ImageService) BatchUploadProductImages(req *BatchUploadImageRequest) ([]*model.ProductImage, error) {
	// 验证商品是否存在
	var product model.Product
	if err := is.db.First(&product, req.ProductID).Error; err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	// 开始事务
	tx := is.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var images []*model.ProductImage

	for i, fileID := range req.FileIDs {
		// 验证文件是否存在
		fileInfo, err := is.fileManager.GetFileInfo(fileID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("文件ID %d 不存在: %v", fileID, err)
		}

		// 验证文件类型
		if !is.isImageFile(fileInfo.OriginalName) {
			tx.Rollback()
			return nil, fmt.Errorf("文件ID %d 不是图片文件", fileID)
		}

		// 获取Alt文本
		alt := ""
		if i < len(req.Alts) {
			alt = req.Alts[i]
		}

		// 创建商品图片记录
		image := &model.ProductImage{
			ProductID: req.ProductID,
			URL:       fileInfo.URL(),
			Alt:       alt,
			Sort:      i,
			IsMain:    i == 0, // 第一张图片设为主图
		}

		if err := tx.Create(image).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("创建商品图片失败: %v", err)
		}

		images = append(images, image)
	}

	// 如果有新的主图，需要将其他图片的主图状态取消
	if len(images) > 0 && images[0].IsMain {
		if err := tx.Model(&model.ProductImage{}).
			Where("product_id = ? AND id != ?", req.ProductID, images[0].ID).
			Update("is_main", false).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新主图状态失败: %v", err)
		}
	}

	tx.Commit()
	return images, nil
}

// UpdateProductImage 更新商品图片
func (is *ImageService) UpdateProductImage(id uint, req *UpdateImageRequest) (*model.ProductImage, error) {
	var image model.ProductImage
	if err := is.db.First(&image, id).Error; err != nil {
		return nil, fmt.Errorf("图片不存在")
	}

	// 如果设置为主图，需要将其他图片的主图状态取消
	if req.IsMain && !image.IsMain {
		if err := is.db.Model(&model.ProductImage{}).
			Where("product_id = ? AND id != ?", image.ProductID, id).
			Update("is_main", false).Error; err != nil {
			return nil, fmt.Errorf("更新主图状态失败: %v", err)
		}
	}

	// 更新图片信息
	image.Alt = req.Alt
	image.Sort = req.Sort
	image.IsMain = req.IsMain

	if err := is.db.Save(&image).Error; err != nil {
		return nil, fmt.Errorf("更新商品图片失败: %v", err)
	}

	return &image, nil
}

// DeleteProductImage 删除商品图片
func (is *ImageService) DeleteProductImage(id uint) error {
	var image model.ProductImage
	if err := is.db.First(&image, id).Error; err != nil {
		return fmt.Errorf("图片不存在")
	}

	// 如果删除的是主图，需要将第一张图片设为主图
	if image.IsMain {
		var firstImage model.ProductImage
		if err := is.db.Where("product_id = ? AND id != ?", image.ProductID, id).
			Order("sort ASC, id ASC").
			First(&firstImage).Error; err == nil {
			is.db.Model(&firstImage).Update("is_main", true)
		}
	}

	// 删除图片记录
	if err := is.db.Delete(&image).Error; err != nil {
		return fmt.Errorf("删除商品图片失败: %v", err)
	}

	return nil
}

// GetProductImage 获取商品图片详情
func (is *ImageService) GetProductImage(id uint) (*model.ProductImage, error) {
	var image model.ProductImage
	if err := is.db.Preload("Product").First(&image, id).Error; err != nil {
		return nil, fmt.Errorf("图片不存在")
	}

	return &image, nil
}

// GetProductImages 获取商品图片列表
func (is *ImageService) GetProductImages(req *ImageListRequest) ([]*model.ProductImage, int64, error) {
	query := is.db.Model(&model.ProductImage{}).Where("product_id = ?", req.ProductID)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询图片总数失败: %v", err)
	}

	// 分页查询
	var images []*model.ProductImage
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("is_main DESC, sort ASC, id ASC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&images).Error; err != nil {
		return nil, 0, fmt.Errorf("查询图片列表失败: %v", err)
	}

	return images, total, nil
}

// GetAllProductImages 获取商品所有图片
func (is *ImageService) GetAllProductImages(productID uint) ([]*model.ProductImage, error) {
	var images []*model.ProductImage
	if err := is.db.Where("product_id = ?", productID).
		Order("is_main DESC, sort ASC, id ASC").
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("查询商品图片失败: %v", err)
	}

	return images, nil
}

// SetMainImage 设置主图
func (is *ImageService) SetMainImage(id uint) error {
	var image model.ProductImage
	if err := is.db.First(&image, id).Error; err != nil {
		return fmt.Errorf("图片不存在")
	}

	// 开始事务
	tx := is.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 将该商品的所有图片设为非主图
	if err := tx.Model(&model.ProductImage{}).
		Where("product_id = ?", image.ProductID).
		Update("is_main", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新主图状态失败: %v", err)
	}

	// 将指定图片设为主图
	if err := tx.Model(&image).Update("is_main", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("设置主图失败: %v", err)
	}

	tx.Commit()
	return nil
}

// SortImages 图片排序
func (is *ImageService) SortImages(productID uint, imageIDs []uint) error {
	// 验证商品是否存在
	var product model.Product
	if err := is.db.First(&product, productID).Error; err != nil {
		return fmt.Errorf("商品不存在")
	}

	// 验证图片是否都属于该商品
	var count int64
	if err := is.db.Model(&model.ProductImage{}).
		Where("product_id = ? AND id IN ?", productID, imageIDs).
		Count(&count).Error; err != nil {
		return fmt.Errorf("验证图片失败: %v", err)
	}

	if int(count) != len(imageIDs) {
		return fmt.Errorf("存在不属于该商品的图片")
	}

	// 开始事务
	tx := is.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新图片排序
	for i, imageID := range imageIDs {
		if err := tx.Model(&model.ProductImage{}).
			Where("id = ?", imageID).
			Update("sort", i).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新图片排序失败: %v", err)
		}
	}

	tx.Commit()
	return nil
}

// BatchDeleteImages 批量删除图片
func (is *ImageService) BatchDeleteImages(ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("图片ID列表不能为空")
	}

	// 查询要删除的图片
	var images []model.ProductImage
	if err := is.db.Where("id IN ?", ids).Find(&images).Error; err != nil {
		return fmt.Errorf("查询图片失败: %v", err)
	}

	if len(images) == 0 {
		return fmt.Errorf("没有找到要删除的图片")
	}

	// 开始事务
	tx := is.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 按商品分组处理
	productMainImages := make(map[uint]bool)
	for _, image := range images {
		if image.IsMain {
			productMainImages[image.ProductID] = true
		}
	}

	// 删除图片
	if err := tx.Where("id IN ?", ids).Delete(&model.ProductImage{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除图片失败: %v", err)
	}

	// 为删除了主图的商品重新设置主图
	for productID := range productMainImages {
		var firstImage model.ProductImage
		if err := tx.Where("product_id = ?", productID).
			Order("sort ASC, id ASC").
			First(&firstImage).Error; err == nil {
			tx.Model(&firstImage).Update("is_main", true)
		}
	}

	tx.Commit()
	return nil
}

// GetImageStatistics 获取图片统计信息
func (is *ImageService) GetImageStatistics(productID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := is.db.Model(&model.ProductImage{})
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}

	// 总图片数
	var totalCount int64
	query.Count(&totalCount)
	stats["total_images"] = totalCount

	// 主图数量
	var mainImageCount int64
	query.Where("is_main = ?", true).Count(&mainImageCount)
	stats["main_image_count"] = mainImageCount

	return stats, nil
}

// isImageFile 检查是否为图片文件
func (is *ImageService) isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

	for _, imageExt := range imageExts {
		if ext == imageExt {
			return true
		}
	}

	return false
}

// 全局图片服务实例
var globalImageService *ImageService

// InitGlobalImageService 初始化全局图片服务
func InitGlobalImageService(db *gorm.DB, fileManager *upload.FileManager) {
	globalImageService = NewImageService(db, fileManager)
}

// GetGlobalImageService 获取全局图片服务
func GetGlobalImageService() *ImageService {
	return globalImageService
}
