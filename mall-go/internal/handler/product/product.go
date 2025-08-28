package product

import (
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"mall-go/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// List 获取商品列表
// @Summary 获取商品列表
// @Description 分页获取商品列表，支持分类筛选和关键词搜索
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param category_id query int false "分类ID"
// @Param keyword query string false "搜索关键词"
// @Param status query string false "商品状态"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /products [get]
func (h *Handler) List(c *gin.Context) {
	var req model.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建查询
	query := h.db.Model(&model.Product{}).Preload("Category").Preload("Images")

	// 分类筛选
	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询
	var products []model.Product
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&products).Error; err != nil {
		logger.Error("查询商品列表失败", zap.Error(err))
		response.ServerError(c, "查询商品列表失败")
		return
	}

	response.SuccessWithPage(c, "查询成功", products, total, req.Page, req.PageSize)
}

// Get 获取商品详情
// @Summary 获取商品详情
// @Description 根据商品ID获取商品详细信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Success 200 {object} model.Product
// @Failure 404 {object} map[string]interface{}
// @Router /products/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	var product model.Product
	if err := h.db.Preload("Category").Preload("Images").First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "商品不存在")
			return
		}
		logger.Error("查询商品详情失败", zap.Error(err))
		response.ServerError(c, "查询商品详情失败")
		return
	}

	response.SuccessWithData(c, product)
}

// Create 创建商品
// @Summary 创建商品
// @Description 创建新商品（需要管理员权限）
// @Tags 商品管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body model.ProductCreateRequest true "商品信息"
// @Success 200 {object} model.Product
// @Failure 400 {object} map[string]interface{}
// @Router /products [post]
func (h *Handler) Create(c *gin.Context) {
	var req model.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 检查分类是否存在
	var category model.Category
	if err := h.db.First(&category, req.CategoryID).Error; err != nil {
		response.BadRequest(c, "商品分类不存在")
		return
	}

	// 创建商品
	product := model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       decimal.NewFromFloat(req.Price),
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		Status:      "active",
	}

	if err := h.db.Create(&product).Error; err != nil {
		logger.Error("创建商品失败", zap.Error(err))
		response.ServerError(c, "创建商品失败")
		return
	}

	// 创建商品图片
	if len(req.Images) > 0 {
		for i, imageURL := range req.Images {
			productImage := model.ProductImage{
				ProductID: product.ID,
				URL:       imageURL,
				Sort:      i,
			}
			h.db.Create(&productImage)
		}
	}

	// 重新查询商品（包含关联数据）
	h.db.Preload("Category").Preload("Images").First(&product, product.ID)

	response.Success(c, "商品创建成功", product)
}

// Update 更新商品
// @Summary 更新商品
// @Description 更新商品信息（需要管理员权限）
// @Tags 商品管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "商品ID"
// @Param product body model.ProductUpdateRequest true "商品信息"
// @Success 200 {object} model.Product
// @Failure 400 {object} map[string]interface{}
// @Router /products/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	var req model.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	var product model.Product
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "商品不存在")
			return
		}
		response.ServerError(c, "查询商品失败")
		return
	}

	// 更新商品信息
	product.Name = req.Name
	product.Description = req.Description
	product.Price = decimal.NewFromFloat(req.Price)
	product.Stock = req.Stock
	product.CategoryID = req.CategoryID

	if err := h.db.Save(&product).Error; err != nil {
		logger.Error("更新商品失败", zap.Error(err))
		response.ServerError(c, "更新商品失败")
		return
	}

	// 更新商品图片
	if len(req.Images) > 0 {
		// 删除旧图片
		h.db.Where("product_id = ?", product.ID).Delete(&model.ProductImage{})

		// 创建新图片
		for i, imageURL := range req.Images {
			productImage := model.ProductImage{
				ProductID: product.ID,
				URL:       imageURL,
				Sort:      i,
			}
			h.db.Create(&productImage)
		}
	}

	// 重新查询商品（包含关联数据）
	h.db.Preload("Category").Preload("Images").First(&product, product.ID)

	response.Success(c, "商品更新成功", product)
}

// Delete 删除商品
// @Summary 删除商品
// @Description 删除商品（需要管理员权限）
// @Tags 商品管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "商品ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /products/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的商品ID")
		return
	}

	var product model.Product
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "商品不存在")
			return
		}
		response.ServerError(c, "查询商品失败")
		return
	}

	// 删除商品图片
	h.db.Where("product_id = ?", product.ID).Delete(&model.ProductImage{})

	// 删除商品
	if err := h.db.Delete(&product).Error; err != nil {
		logger.Error("删除商品失败", zap.Error(err))
		response.ServerError(c, "删除商品失败")
		return
	}

	response.SuccessWithMessage(c, "商品删除成功")
}
