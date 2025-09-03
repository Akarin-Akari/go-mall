package alipay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"mall-go/pkg/payment/config"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Client 支付宝客户端
type Client struct {
	config     *config.AlipayConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	httpClient *http.Client
}

// NewClient 创建支付宝客户端
func NewClient(cfg *config.AlipayConfig) (*Client, error) {
	client := &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}

	// 解析私钥
	if err := client.parsePrivateKey(); err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	// 解析公钥
	if err := client.parsePublicKey(); err != nil {
		return nil, fmt.Errorf("解析公钥失败: %v", err)
	}

	return client, nil
}

// parsePrivateKey 解析私钥
func (c *Client) parsePrivateKey() error {
	block, _ := pem.Decode([]byte(c.config.PrivateKey))
	if block == nil {
		return fmt.Errorf("私钥格式错误")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS1格式
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("解析私钥失败: %v", err)
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("私钥不是RSA格式")
	}

	c.privateKey = rsaPrivateKey
	return nil
}

// parsePublicKey 解析公钥
func (c *Client) parsePublicKey() error {
	block, _ := pem.Decode([]byte(c.config.PublicKey))
	if block == nil {
		return fmt.Errorf("公钥格式错误")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析公钥失败: %v", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("公钥不是RSA格式")
	}

	c.publicKey = rsaPublicKey
	return nil
}

// CreatePayment 创建支付
func (c *Client) CreatePayment(req *PaymentRequest) (*PaymentResponse, error) {
	logger.Info("创建支付宝支付",
		zap.String("out_trade_no", req.OutTradeNo),
		zap.String("total_amount", req.TotalAmount.String()))

	// 构建请求参数
	params := c.buildPaymentParams(req)

	// 签名
	sign, err := c.sign(params)
	if err != nil {
		return nil, fmt.Errorf("签名失败: %v", err)
	}
	params["sign"] = sign

	// 发送请求
	response, err := c.sendRequest(params)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	// 解析响应
	return c.parsePaymentResponse(response)
}

// buildPaymentParams 构建支付参数
func (c *Client) buildPaymentParams(req *PaymentRequest) map[string]string {
	params := map[string]string{
		"app_id":     c.config.AppID,
		"method":     "alipay.trade.precreate", // 扫码支付
		"format":     c.config.Format,
		"charset":    c.config.Charset,
		"sign_type":  c.config.SignType,
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"notify_url": req.NotifyURL,
	}

	// 业务参数
	bizContent := map[string]interface{}{
		"out_trade_no":    req.OutTradeNo,
		"total_amount":    req.TotalAmount.String(),
		"subject":         req.Subject,
		"store_id":        "001",
		"timeout_express": "30m",
	}

	if req.Body != "" {
		bizContent["body"] = req.Body
	}

	bizContentJSON, _ := json.Marshal(bizContent)
	params["biz_content"] = string(bizContentJSON)

	return params
}

// sign 签名
func (c *Client) sign(params map[string]string) (string, error) {
	// 排序参数
	var keys []string
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}

	// RSA2签名
	hash := sha256.Sum256([]byte(signStr.String()))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// sendRequest 发送请求
func (c *Client) sendRequest(params map[string]string) ([]byte, error) {
	// 构建POST数据
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}

	// 发送请求
	resp, err := c.httpClient.PostForm(c.config.GatewayURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// parsePaymentResponse 解析支付响应
func (c *Client) parsePaymentResponse(data []byte) (*PaymentResponse, error) {
	var response struct {
		AlipayTradePrecreateResponse struct {
			Code       string `json:"code"`
			Msg        string `json:"msg"`
			SubCode    string `json:"sub_code"`
			SubMsg     string `json:"sub_msg"`
			OutTradeNo string `json:"out_trade_no"`
			QRCode     string `json:"qr_code"`
		} `json:"alipay_trade_precreate_response"`
		Sign string `json:"sign"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	resp := response.AlipayTradePrecreateResponse
	if resp.Code != "10000" {
		return nil, fmt.Errorf("支付宝返回错误: %s - %s", resp.SubCode, resp.SubMsg)
	}

	return &PaymentResponse{
		OutTradeNo: resp.OutTradeNo,
		QRCode:     resp.QRCode,
		Success:    true,
	}, nil
}

// QueryPayment 查询支付
func (c *Client) QueryPayment(outTradeNo string) (*QueryResponse, error) {
	logger.Info("查询支付宝支付状态", zap.String("out_trade_no", outTradeNo))

	params := map[string]string{
		"app_id":    c.config.AppID,
		"method":    "alipay.trade.query",
		"format":    c.config.Format,
		"charset":   c.config.Charset,
		"sign_type": c.config.SignType,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0",
	}

	bizContent := map[string]interface{}{
		"out_trade_no": outTradeNo,
	}

	bizContentJSON, _ := json.Marshal(bizContent)
	params["biz_content"] = string(bizContentJSON)

	// 签名
	sign, err := c.sign(params)
	if err != nil {
		return nil, fmt.Errorf("签名失败: %v", err)
	}
	params["sign"] = sign

	// 发送请求
	response, err := c.sendRequest(params)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	// 解析响应
	return c.parseQueryResponse(response)
}

// parseQueryResponse 解析查询响应
func (c *Client) parseQueryResponse(data []byte) (*QueryResponse, error) {
	var response struct {
		AlipayTradeQueryResponse struct {
			Code        string `json:"code"`
			Msg         string `json:"msg"`
			SubCode     string `json:"sub_code"`
			SubMsg      string `json:"sub_msg"`
			OutTradeNo  string `json:"out_trade_no"`
			TradeNo     string `json:"trade_no"`
			TradeStatus string `json:"trade_status"`
			TotalAmount string `json:"total_amount"`
		} `json:"alipay_trade_query_response"`
		Sign string `json:"sign"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	resp := response.AlipayTradeQueryResponse
	if resp.Code != "10000" {
		return nil, fmt.Errorf("支付宝返回错误: %s - %s", resp.SubCode, resp.SubMsg)
	}

	// 转换支付状态
	var status model.PaymentStatus
	switch resp.TradeStatus {
	case "WAIT_BUYER_PAY":
		status = model.PaymentStatusPending
	case "TRADE_SUCCESS":
		status = model.PaymentStatusSuccess
	case "TRADE_FINISHED":
		status = model.PaymentStatusSuccess
	case "TRADE_CLOSED":
		status = model.PaymentStatusCancelled
	default:
		status = model.PaymentStatusFailed
	}

	totalAmount, _ := decimal.NewFromString(resp.TotalAmount)

	return &QueryResponse{
		OutTradeNo:  resp.OutTradeNo,
		TradeNo:     resp.TradeNo,
		TradeStatus: resp.TradeStatus,
		TotalAmount: totalAmount,
		Status:      status,
		Success:     true,
	}, nil
}

// VerifyCallback 验证回调
func (c *Client) VerifyCallback(params map[string]string) error {
	sign := params["sign"]
	delete(params, "sign")
	delete(params, "sign_type")

	// 构建签名字符串
	var keys []string
	for k := range params {
		if params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}

	// 验证签名
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("签名解码失败: %v", err)
	}

	hash := sha256.Sum256([]byte(signStr.String()))
	err = rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, hash[:], signBytes)
	if err != nil {
		return fmt.Errorf("签名验证失败: %v", err)
	}

	return nil
}
