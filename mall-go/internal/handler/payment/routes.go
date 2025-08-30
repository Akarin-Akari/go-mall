package payment

import (
	"mall-go/pkg/payment"
	"mall-go/pkg/payment/alipay"
	"mall-go/pkg/payment/wechat"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册支付相关路由
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, paymentService *payment.Service, alipayClient *alipay.Client, wechatClient *wechat.Client) {
	handler := NewHandler(db, paymentService)
	callbackHandler := NewCallbackHandler(db, paymentService, alipayClient, wechatClient)

	// 支付相关路由组
	paymentGroup := router.Group("/payments")
	{
		// 支付基础功能
		paymentGroup.POST("", handler.CreatePayment)     // 创建支付
		paymentGroup.GET("", handler.ListPayments)       // 获取支付列表
		paymentGroup.GET("/query", handler.QueryPayment) // 查询支付状态
		paymentGroup.GET("/:id", handler.GetPaymentByID) // 获取支付详情

		// 退款功能
		paymentGroup.POST("/refund", handler.RefundPayment) // 申请退款

		// 支付方式和配置
		paymentGroup.GET("/methods", handler.GetPaymentMethods) // 获取支付方式列表

		// 统计功能
		paymentGroup.GET("/statistics", handler.GetPaymentStatistics) // 获取支付统计

		// 回调处理路由组
		callbackGroup := paymentGroup.Group("/callback")
		{
			callbackGroup.POST("/alipay", callbackHandler.AlipayCallback) // 支付宝回调
			callbackGroup.POST("/wechat", callbackHandler.WechatCallback) // 微信支付回调
		}
	}
}
