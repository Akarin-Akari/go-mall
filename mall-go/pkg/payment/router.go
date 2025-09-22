package payment

import (
	"fmt"
	"math/rand"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// PaymentRouter 支付路由器 - 智能选择支付方式
type PaymentRouter struct {
	config   *PaymentConfig
	alipay   AlipayClientInterface
	wechat   WechatClientInterface
	unionpay UnionPayClientInterface
	metrics  *PaymentMetrics
}

// AlipayClientInterface 支付宝客户端接口
type AlipayClientInterface interface {
	CreatePayment(req *CreatePaymentRequest) (*CreatePaymentResponse, error)
	QueryPayment(outTradeNo string) (*QueryPaymentResponse, error)
	VerifyCallback(params map[string]string) error
}

// WechatClientInterface 微信支付客户端接口
type WechatClientInterface interface {
	CreatePayment(req *CreatePaymentRequest) (*CreatePaymentResponse, error)
	QueryPayment(outTradeNo string) (*QueryPaymentResponse, error)
	VerifyCallback(params map[string]string) error
}

// UnionPayClientInterface 银联支付客户端接口
type UnionPayClientInterface interface {
	CreatePayment(req *CreatePaymentRequest) (*CreatePaymentResponse, error)
	QueryPayment(outTradeNo string) (*QueryPaymentResponse, error)
	VerifyCallback(params map[string]string) error
}

// NewPaymentRouter 创建支付路由器
func NewPaymentRouter(config *PaymentConfig) *PaymentRouter {
	return &PaymentRouter{
		config:  config,
		metrics: NewPaymentMetrics(),
	}
}

// CreatePaymentRequest 统一创建支付请求
type CreatePaymentRequest struct {
	OutTradeNo string                 `json:"out_trade_no" binding:"required"`
	Amount     decimal.Decimal        `json:"amount" binding:"required"`
	Subject    string                 `json:"subject" binding:"required"`
	Body       string                 `json:"body,omitempty"`
	Method     model.PaymentMethod    `json:"method" binding:"required,oneof=alipay wechat unionpay"`
	UserID     uint                   `json:"user_id" binding:"required"`
	Currency   string                 `json:"currency,omitempty"`
	ExpireTime *time.Time             `json:"expire_time,omitempty"`
	NotifyURL  string                 `json:"notify_url,omitempty"`
	ReturnURL  string                 `json:"return_url,omitempty"`
	ClientIP   string                 `json:"client_ip,omitempty"`
	DeviceInfo string                 `json:"device_info,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// CreatePaymentResponse 统一创建支付响应
type CreatePaymentResponse struct {
	OutTradeNo string                 `json:"out_trade_no"`
	Method     model.PaymentMethod    `json:"method"`
	QRCode     string                 `json:"qr_code,omitempty"`
	CodeURL    string                 `json:"code_url,omitempty"`
	PayURL     string                 `json:"pay_url,omitempty"`
	PrepayID   string                 `json:"prepay_id,omitempty"`
	Success    bool                   `json:"success"`
	Message    string                 `json:"message,omitempty"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// QueryPaymentResponse 统一查询支付响应
type QueryPaymentResponse struct {
	OutTradeNo    string                 `json:"out_trade_no"`
	TransactionID string                 `json:"transaction_id"`
	Method        model.PaymentMethod    `json:"method"`
	Status        model.PaymentStatus    `json:"status"`
	Amount        decimal.Decimal        `json:"amount"`
	PaidAt        *time.Time             `json:"paid_at,omitempty"`
	Success       bool                   `json:"success"`
	Message       string                 `json:"message,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// CreatePayment 创建支付订单
func (pr *PaymentRouter) CreatePayment(req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	startTime := time.Now()

	logger.Info("开始创建支付订单",
		zap.String("out_trade_no", req.OutTradeNo),
		zap.String("method", string(req.Method)),
		zap.String("amount", req.Amount.String()))

	// 预处理请求
	if err := pr.preprocessRequest(req); err != nil {
		pr.metrics.RecordCreatePayment(req.Method, "preprocess_error", time.Since(startTime))
		return nil, fmt.Errorf("预处理请求失败: %v", err)
	}

	// 验证支付限额
	if err := pr.validatePaymentLimits(req); err != nil {
		pr.metrics.RecordCreatePayment(req.Method, "limit_error", time.Since(startTime))
		return nil, fmt.Errorf("支付限额验证失败: %v", err)
	}

	// 根据支付方式路由
	var response *CreatePaymentResponse
	var err error

	switch req.Method {
	case model.PaymentMethodAlipay:
		if !pr.config.Alipay.Enabled {
			return nil, fmt.Errorf("支付宝支付未启用")
		}
		response, err = pr.createAlipayPayment(req)

	case model.PaymentMethodWechat:
		if !pr.config.Wechat.Enabled {
			return nil, fmt.Errorf("微信支付未启用")
		}
		response, err = pr.createWechatPayment(req)

	case model.PaymentMethodUnionPay:
		if !pr.config.UnionPay.Enabled {
			return nil, fmt.Errorf("银联支付未启用")
		}
		response, err = pr.createUnionPayPayment(req)

	default:
		err = fmt.Errorf("不支持的支付方式: %s", req.Method)
	}

	// 记录指标
	status := "success"
	if err != nil {
		status = "error"
	}
	pr.metrics.RecordCreatePayment(req.Method, status, time.Since(startTime))

	if err != nil {
		logger.Error("创建支付订单失败",
			zap.String("out_trade_no", req.OutTradeNo),
			zap.String("method", string(req.Method)),
			zap.Error(err))
		return nil, err
	}

	logger.Info("支付订单创建成功",
		zap.String("out_trade_no", req.OutTradeNo),
		zap.String("method", string(req.Method)),
		zap.Duration("duration", time.Since(startTime)))

	return response, nil
}

// preprocessRequest 预处理请求
func (pr *PaymentRouter) preprocessRequest(req *CreatePaymentRequest) error {
	// 设置默认币种
	if req.Currency == "" {
		req.Currency = pr.config.DefaultCurrency
	}

	// 设置默认过期时间（30分钟）
	if req.ExpireTime == nil {
		expireTime := time.Now().Add(30 * time.Minute)
		req.ExpireTime = &expireTime
	}

	// 设置默认回调地址
	if req.NotifyURL == "" {
		switch req.Method {
		case model.PaymentMethodAlipay:
			req.NotifyURL = pr.config.Alipay.NotifyURL
		case model.PaymentMethodWechat:
			req.NotifyURL = pr.config.Wechat.NotifyURL
		case model.PaymentMethodUnionPay:
			req.NotifyURL = pr.config.UnionPay.NotifyURL
		}
	}

	// 金额验证
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("支付金额必须大于0")
	}

	// 金额精度验证（保留2位小数）
	if req.Amount.Exponent() < -2 {
		return fmt.Errorf("支付金额精度不能超过2位小数")
	}

	return nil
}

// validatePaymentLimits 验证支付限额
func (pr *PaymentRouter) validatePaymentLimits(req *CreatePaymentRequest) error {
	methodLimit := pr.config.GetMethodLimit(req.Method)

	// 单笔限额验证
	if req.Amount.LessThan(methodLimit.MinAmount) {
		return fmt.Errorf("支付金额不能小于 %s", methodLimit.MinAmount.String())
	}

	if req.Amount.GreaterThan(methodLimit.MaxAmount) {
		return fmt.Errorf("支付金额不能大于 %s", methodLimit.MaxAmount.String())
	}

	// TODO: 实现日限额和月限额验证
	// 这里需要查询数据库统计用户当日和当月的支付金额

	return nil
}

// createAlipayPayment 创建支付宝支付
func (pr *PaymentRouter) createAlipayPayment(req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// TODO: 调用具体的支付宝客户端
	// 这里返回模拟响应
	return &CreatePaymentResponse{
		OutTradeNo: req.OutTradeNo,
		Method:     model.PaymentMethodAlipay,
		QRCode:     pr.generateMockQRCode("alipay", req.OutTradeNo),
		Success:    true,
		ExpiresAt:  req.ExpireTime,
	}, nil
}

// createWechatPayment 创建微信支付
func (pr *PaymentRouter) createWechatPayment(req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// TODO: 调用具体的微信支付客户端
	// 这里返回模拟响应
	return &CreatePaymentResponse{
		OutTradeNo: req.OutTradeNo,
		Method:     model.PaymentMethodWechat,
		CodeURL:    pr.generateMockQRCode("wechat", req.OutTradeNo),
		Success:    true,
		ExpiresAt:  req.ExpireTime,
	}, nil
}

// createUnionPayPayment 创建银联支付
func (pr *PaymentRouter) createUnionPayPayment(req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// TODO: 调用具体的银联支付客户端
	// 这里返回模拟响应
	return &CreatePaymentResponse{
		OutTradeNo: req.OutTradeNo,
		Method:     model.PaymentMethodUnionPay,
		PayURL:     fmt.Sprintf("https://gateway.95516.com/pay?orderNo=%s", req.OutTradeNo),
		Success:    true,
		ExpiresAt:  req.ExpireTime,
	}, nil
}

// QueryPayment 查询支付状态
func (pr *PaymentRouter) QueryPayment(outTradeNo string, method model.PaymentMethod) (*QueryPaymentResponse, error) {
	startTime := time.Now()

	logger.Info("查询支付状态",
		zap.String("out_trade_no", outTradeNo),
		zap.String("method", string(method)))

	var response *QueryPaymentResponse
	var err error

	switch method {
	case model.PaymentMethodAlipay:
		if !pr.config.Alipay.Enabled {
			return nil, fmt.Errorf("支付宝支付未启用")
		}
		response, err = pr.queryAlipayPayment(outTradeNo)

	case model.PaymentMethodWechat:
		if !pr.config.Wechat.Enabled {
			return nil, fmt.Errorf("微信支付未启用")
		}
		response, err = pr.queryWechatPayment(outTradeNo)

	case model.PaymentMethodUnionPay:
		if !pr.config.UnionPay.Enabled {
			return nil, fmt.Errorf("银联支付未启用")
		}
		response, err = pr.queryUnionPayPayment(outTradeNo)

	default:
		err = fmt.Errorf("不支持的支付方式: %s", method)
	}

	// 记录指标
	status := "success"
	if err != nil {
		status = "error"
	}
	pr.metrics.RecordQueryPayment(method, status, time.Since(startTime))

	if err != nil {
		logger.Error("查询支付状态失败",
			zap.String("out_trade_no", outTradeNo),
			zap.String("method", string(method)),
			zap.Error(err))
		return nil, err
	}

	return response, nil
}

// queryAlipayPayment 查询支付宝支付状态
func (pr *PaymentRouter) queryAlipayPayment(outTradeNo string) (*QueryPaymentResponse, error) {
	// TODO: 调用具体的支付宝客户端
	// 这里返回模拟响应
	return &QueryPaymentResponse{
		OutTradeNo:    outTradeNo,
		TransactionID: fmt.Sprintf("alipay_%d", time.Now().Unix()),
		Method:        model.PaymentMethodAlipay,
		Status:        model.PaymentStatusSuccess,
		Amount:        decimal.NewFromFloat(99.99),
		Success:       true,
	}, nil
}

// queryWechatPayment 查询微信支付状态
func (pr *PaymentRouter) queryWechatPayment(outTradeNo string) (*QueryPaymentResponse, error) {
	// TODO: 调用具体的微信支付客户端
	// 这里返回模拟响应
	return &QueryPaymentResponse{
		OutTradeNo:    outTradeNo,
		TransactionID: fmt.Sprintf("wechat_%d", time.Now().Unix()),
		Method:        model.PaymentMethodWechat,
		Status:        model.PaymentStatusSuccess,
		Amount:        decimal.NewFromFloat(88.88),
		Success:       true,
	}, nil
}

// queryUnionPayPayment 查询银联支付状态
func (pr *PaymentRouter) queryUnionPayPayment(outTradeNo string) (*QueryPaymentResponse, error) {
	// TODO: 调用具体的银联支付客户端
	// 这里返回模拟响应
	return &QueryPaymentResponse{
		OutTradeNo:    outTradeNo,
		TransactionID: fmt.Sprintf("unionpay_%d", time.Now().Unix()),
		Method:        model.PaymentMethodUnionPay,
		Status:        model.PaymentStatusSuccess,
		Amount:        decimal.NewFromFloat(77.77),
		Success:       true,
	}, nil
}

// generateMockQRCode 生成模拟二维码
func (pr *PaymentRouter) generateMockQRCode(method, outTradeNo string) string {
	// 生成模拟的支付二维码URL
	return fmt.Sprintf("https://qr.%s.com/pay?orderNo=%s&t=%d&r=%d",
		method, outTradeNo, time.Now().Unix(), rand.Intn(10000))
}

// GetSupportedMethods 获取支持的支付方式
func (pr *PaymentRouter) GetSupportedMethods() []model.PaymentMethod {
	var methods []model.PaymentMethod

	if pr.config.Alipay.Enabled {
		methods = append(methods, model.PaymentMethodAlipay)
	}
	if pr.config.Wechat.Enabled {
		methods = append(methods, model.PaymentMethodWechat)
	}
	if pr.config.UnionPay.Enabled {
		methods = append(methods, model.PaymentMethodUnionPay)
	}

	return methods
}

// GetRecommendedMethod 根据金额和用户偏好推荐支付方式
func (pr *PaymentRouter) GetRecommendedMethod(amount decimal.Decimal, userID uint) model.PaymentMethod {
	// 获取支持的支付方式
	supportedMethods := pr.GetSupportedMethods()
	if len(supportedMethods) == 0 {
		return ""
	}

	// 根据金额推荐
	if amount.GreaterThan(decimal.NewFromFloat(5000)) {
		// 大额推荐银联
		for _, method := range supportedMethods {
			if method == model.PaymentMethodUnionPay {
				return method
			}
		}
	}

	if amount.LessThan(decimal.NewFromFloat(100)) {
		// 小额推荐微信
		for _, method := range supportedMethods {
			if method == model.PaymentMethodWechat {
				return method
			}
		}
	}

	// 默认推荐支付宝
	for _, method := range supportedMethods {
		if method == model.PaymentMethodAlipay {
			return method
		}
	}

	// 返回第一个可用的支付方式
	return supportedMethods[0]
}
