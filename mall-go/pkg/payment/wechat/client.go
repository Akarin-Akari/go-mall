package wechat

import (
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// WechatConfig 微信支付配置
type WechatConfig struct {
	AppID      string        `json:"app_id"`
	MchID      string        `json:"mch_id"`
	APIKey     string        `json:"api_key"`
	GatewayURL string        `json:"gateway_url"`
	Timeout    time.Duration `json:"timeout"`
}

// Client 微信支付客户端
type Client struct {
	config     *WechatConfig
	httpClient *http.Client
}

// NewClient 创建微信支付客户端
func NewClient(config *WechatConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// CreatePayment 创建支付
func (c *Client) CreatePayment(req *PaymentRequest) (*PaymentResponse, error) {
	logger.Info("创建微信支付",
		zap.String("out_trade_no", req.OutTradeNo),
		zap.String("total_fee", req.TotalFee.String()))

	// 构建请求参数
	params := c.buildPaymentParams(req)

	// 签名
	params["sign"] = c.sign(params)

	// 构建XML请求
	xmlData, err := c.buildXMLRequest(params)
	if err != nil {
		return nil, fmt.Errorf("构建XML请求失败: %v", err)
	}

	// 发送请求
	response, err := c.sendRequest(c.config.GatewayURL+"/pay/unifiedorder", xmlData)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	// 解析响应
	return c.parsePaymentResponse(response)
}

// buildPaymentParams 构建支付参数
func (c *Client) buildPaymentParams(req *PaymentRequest) map[string]string {
	// 将金额转换为分
	totalFee := req.TotalFee.Mul(decimal.NewFromInt(100)).IntPart()

	params := map[string]string{
		"appid":            c.config.AppID,
		"mch_id":           c.config.MchID,
		"nonce_str":        c.generateNonceStr(),
		"body":             req.Body,
		"out_trade_no":     req.OutTradeNo,
		"total_fee":        strconv.FormatInt(totalFee, 10),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       req.NotifyURL,
		"trade_type":       "NATIVE", // 扫码支付
	}

	if req.Detail != "" {
		params["detail"] = req.Detail
	}

	if req.Attach != "" {
		params["attach"] = req.Attach
	}

	if req.TimeExpire != nil {
		params["time_expire"] = req.TimeExpire.Format("20060102150405")
	}

	return params
}

// generateNonceStr 生成随机字符串
func (c *Client) generateNonceStr() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// sign 签名
func (c *Client) sign(params map[string]string) string {
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
	signStr.WriteString("&key=")
	signStr.WriteString(c.config.APIKey)

	// MD5签名
	hash := md5.Sum([]byte(signStr.String()))
	return fmt.Sprintf("%X", hash)
}

// buildXMLRequest 构建XML请求
func (c *Client) buildXMLRequest(params map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("<xml>")

	for k, v := range params {
		buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}

	buf.WriteString("</xml>")
	return buf.Bytes(), nil
}

// sendRequest 发送请求
func (c *Client) sendRequest(url string, data []byte) ([]byte, error) {
	resp, err := c.httpClient.Post(url, "application/xml", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// parsePaymentResponse 解析支付响应
func (c *Client) parsePaymentResponse(data []byte) (*PaymentResponse, error) {
	var response struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
		ResultCode string   `xml:"result_code"`
		ErrCode    string   `xml:"err_code"`
		ErrCodeDes string   `xml:"err_code_des"`
		AppID      string   `xml:"appid"`
		MchID      string   `xml:"mch_id"`
		NonceStr   string   `xml:"nonce_str"`
		Sign       string   `xml:"sign"`
		PrepayID   string   `xml:"prepay_id"`
		TradeType  string   `xml:"trade_type"`
		CodeURL    string   `xml:"code_url"`
	}

	if err := xml.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if response.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("微信支付返回错误: %s", response.ReturnMsg)
	}

	if response.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("微信支付业务错误: %s - %s", response.ErrCode, response.ErrCodeDes)
	}

	return &PaymentResponse{
		PrepayID:  response.PrepayID,
		CodeURL:   response.CodeURL,
		TradeType: response.TradeType,
		Success:   true,
	}, nil
}

// QueryPayment 查询支付
func (c *Client) QueryPayment(outTradeNo string) (*QueryResponse, error) {
	logger.Info("查询微信支付状态", zap.String("out_trade_no", outTradeNo))

	params := map[string]string{
		"appid":        c.config.AppID,
		"mch_id":       c.config.MchID,
		"out_trade_no": outTradeNo,
		"nonce_str":    c.generateNonceStr(),
	}

	// 签名
	params["sign"] = c.sign(params)

	// 构建XML请求
	xmlData, err := c.buildXMLRequest(params)
	if err != nil {
		return nil, fmt.Errorf("构建XML请求失败: %v", err)
	}

	// 发送请求
	response, err := c.sendRequest(c.config.GatewayURL+"/pay/orderquery", xmlData)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	// 解析响应
	return c.parseQueryResponse(response)
}

// parseQueryResponse 解析查询响应
func (c *Client) parseQueryResponse(data []byte) (*QueryResponse, error) {
	var response struct {
		XMLName       xml.Name `xml:"xml"`
		ReturnCode    string   `xml:"return_code"`
		ReturnMsg     string   `xml:"return_msg"`
		ResultCode    string   `xml:"result_code"`
		ErrCode       string   `xml:"err_code"`
		ErrCodeDes    string   `xml:"err_code_des"`
		OutTradeNo    string   `xml:"out_trade_no"`
		TransactionID string   `xml:"transaction_id"`
		TradeState    string   `xml:"trade_state"`
		TotalFee      string   `xml:"total_fee"`
		TimeEnd       string   `xml:"time_end"`
	}

	if err := xml.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if response.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("微信支付返回错误: %s", response.ReturnMsg)
	}

	if response.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("微信支付业务错误: %s - %s", response.ErrCode, response.ErrCodeDes)
	}

	// 转换支付状态
	var status model.PaymentStatus
	switch response.TradeState {
	case "SUCCESS":
		status = model.PaymentStatusSuccess
	case "REFUND":
		status = model.PaymentStatusRefunded
	case "NOTPAY":
		status = model.PaymentStatusPending
	case "CLOSED":
		status = model.PaymentStatusCancelled
	case "REVOKED":
		status = model.PaymentStatusCancelled
	case "USERPAYING":
		status = model.PaymentStatusPaying
	case "PAYERROR":
		status = model.PaymentStatusFailed
	default:
		status = model.PaymentStatusFailed
	}

	// 转换金额（分转元）
	totalFee, _ := strconv.ParseInt(response.TotalFee, 10, 64)
	totalAmount := decimal.NewFromInt(totalFee).Div(decimal.NewFromInt(100))

	return &QueryResponse{
		OutTradeNo:    response.OutTradeNo,
		TransactionID: response.TransactionID,
		TradeState:    response.TradeState,
		TotalFee:      totalAmount,
		Status:        status,
		TimeEnd:       response.TimeEnd,
		Success:       true,
	}, nil
}

// VerifyCallback 验证回调
func (c *Client) VerifyCallback(params map[string]string) error {
	sign := params["sign"]
	delete(params, "sign")

	// 计算签名
	calculatedSign := c.sign(params)

	// 验证签名
	if sign != calculatedSign {
		return fmt.Errorf("签名验证失败")
	}

	return nil
}

// ParseCallback 解析回调数据
func (c *Client) ParseCallback(data []byte) (*CallbackData, error) {
	var callback CallbackData
	if err := xml.Unmarshal(data, &callback); err != nil {
		return nil, fmt.Errorf("解析回调数据失败: %v", err)
	}

	// 验证基本字段
	if callback.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("回调返回失败: %s", callback.ReturnMsg)
	}

	if callback.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("回调业务失败: %s - %s", callback.ErrCode, callback.ErrCodeDes)
	}

	return &callback, nil
}
