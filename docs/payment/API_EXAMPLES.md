# 支付系统 API 使用示例

## 目录

- [环境准备](#环境准备)
- [基础使用流程](#基础使用流程)
- [代码示例](#代码示例)
- [前端集成示例](#前端集成示例)
- [错误处理](#错误处理)
- [最佳实践](#最佳实践)

## 环境准备

### 1. 获取支付凭证

**支付宝沙箱环境:**
1. 访问 [支付宝开放平台](https://open.alipay.com/)
2. 创建应用获取 `APP_ID`
3. 生成应用私钥和获取支付宝公钥
4. 配置回调地址

**微信支付测试环境:**
1. 访问 [微信支付商户平台](https://pay.weixin.qq.com/)
2. 获取 `APP_ID` 和 `MCH_ID`
3. 设置 API 密钥
4. 配置回调地址

### 2. 配置环境变量

```bash
# .env 文件
PAYMENT_ENVIRONMENT=dev
PAYMENT_DEBUG=true

# 支付宝配置
ALIPAY_APP_ID=2021000000000000
ALIPAY_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC..."
ALIPAY_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuWJKrQ6SWvS9..."

# 微信支付配置
WECHAT_APP_ID=wx1234567890abcdef
WECHAT_MCH_ID=1234567890
WECHAT_API_KEY=your_32_character_api_key_here
```

## 基础使用流程

### 完整支付流程示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "mall-go/internal/model"
    "mall-go/pkg/payment"
    
    "github.com/shopspring/decimal"
)

func main() {
    // 1. 初始化支付服务
    config := payment.LoadConfigFromEnv()
    paymentService, err := payment.NewService(db, config)
    if err != nil {
        log.Fatal("初始化支付服务失败:", err)
    }
    
    // 2. 创建订单（示例）
    order := createTestOrder()
    
    // 3. 创建支付
    paymentResp := createPayment(paymentService, order)
    fmt.Printf("支付创建成功，支付单号: %s\n", paymentResp.PaymentNo)
    
    // 4. 轮询查询支付状态（实际应用中通过回调处理）
    checkPaymentStatus(paymentService, paymentResp.PaymentID)
    
    // 5. 处理退款（如需要）
    // processRefund(paymentService, paymentResp.PaymentID)
}

func createTestOrder() *model.Order {
    return &model.Order{
        ID:          123,
        OrderNo:     "ORDER20231201001",
        UserID:      1,
        TotalAmount: decimal.NewFromFloat(99.99),
        Status:      model.OrderStatusPending,
    }
}

func createPayment(service *payment.Service, order *model.Order) *model.PaymentCreateResponse {
    req := &model.PaymentCreateRequest{
        OrderID:        order.ID,
        PaymentMethod:  model.PaymentMethodAlipay,
        Amount:         order.TotalAmount,
        Subject:        "测试商品购买",
        Description:    "这是一个测试订单",
        ExpiredMinutes: 30,
        NotifyURL:      "https://your-domain.com/api/v1/payments/callback/alipay",
        ReturnURL:      "https://your-domain.com/payment/success",
    }
    
    resp, err := service.CreatePayment(req)
    if err != nil {
        log.Fatal("创建支付失败:", err)
    }
    
    return resp
}

func checkPaymentStatus(service *payment.Service, paymentID uint) {
    for i := 0; i < 10; i++ {
        req := &model.PaymentQueryRequest{
            PaymentID: paymentID,
        }
        
        resp, err := service.QueryPayment(req)
        if err != nil {
            log.Printf("查询支付状态失败: %v", err)
            continue
        }
        
        fmt.Printf("支付状态: %s\n", resp.PaymentStatus)
        
        if resp.PaymentStatus == model.PaymentStatusSuccess {
            fmt.Println("支付成功！")
            break
        } else if resp.PaymentStatus == model.PaymentStatusFailed {
            fmt.Println("支付失败！")
            break
        }
        
        time.Sleep(5 * time.Second)
    }
}
```

## 代码示例

### 1. HTTP 客户端示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type PaymentClient struct {
    BaseURL string
    Token   string
    Client  *http.Client
}

func NewPaymentClient(baseURL, token string) *PaymentClient {
    return &PaymentClient{
        BaseURL: baseURL,
        Token:   token,
        Client:  &http.Client{Timeout: 30 * time.Second},
    }
}

// 创建支付
func (c *PaymentClient) CreatePayment(req CreatePaymentRequest) (*CreatePaymentResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/v1/payments", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+c.Token)
    
    resp, err := c.Client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    var result APIResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }
    
    if result.Code != 200 {
        return nil, fmt.Errorf("API错误: %s", result.Message)
    }
    
    var paymentResp CreatePaymentResponse
    if err := json.Unmarshal(result.Data, &paymentResp); err != nil {
        return nil, err
    }
    
    return &paymentResp, nil
}

// 查询支付状态
func (c *PaymentClient) QueryPayment(paymentID uint) (*QueryPaymentResponse, error) {
    url := fmt.Sprintf("%s/api/v1/payments/%d", c.BaseURL, paymentID)
    
    httpReq, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Authorization", "Bearer "+c.Token)
    
    resp, err := c.Client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    var result APIResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }
    
    if result.Code != 200 {
        return nil, fmt.Errorf("API错误: %s", result.Message)
    }
    
    var queryResp QueryPaymentResponse
    if err := json.Unmarshal(result.Data, &queryResp); err != nil {
        return nil, err
    }
    
    return &queryResp, nil
}

// 数据结构定义
type CreatePaymentRequest struct {
    OrderID        uint   `json:"order_id"`
    PaymentMethod  string `json:"payment_method"`
    Amount         string `json:"amount"`
    Subject        string `json:"subject"`
    Description    string `json:"description"`
    ExpiredMinutes int    `json:"expired_minutes"`
    NotifyURL      string `json:"notify_url"`
    ReturnURL      string `json:"return_url"`
}

type CreatePaymentResponse struct {
    PaymentID     uint                   `json:"payment_id"`
    PaymentNo     string                 `json:"payment_no"`
    PaymentMethod string                 `json:"payment_method"`
    Amount        string                 `json:"amount"`
    PaymentData   map[string]interface{} `json:"payment_data"`
    ExpiredAt     string                 `json:"expired_at"`
    CreatedAt     string                 `json:"created_at"`
}

type QueryPaymentResponse struct {
    PaymentID     uint   `json:"payment_id"`
    PaymentNo     string `json:"payment_no"`
    PaymentStatus string `json:"payment_status"`
    Amount        string `json:"amount"`
    PaidAt        string `json:"paid_at"`
}

type APIResponse struct {
    Code    int             `json:"code"`
    Message string          `json:"message"`
    Data    json.RawMessage `json:"data"`
}
```

### 2. 使用示例

```go
func main() {
    client := NewPaymentClient("http://localhost:8080", "your_jwt_token")
    
    // 创建支付
    createReq := CreatePaymentRequest{
        OrderID:        123,
        PaymentMethod:  "alipay",
        Amount:         "99.99",
        Subject:        "测试商品",
        Description:    "测试商品描述",
        ExpiredMinutes: 30,
        NotifyURL:      "https://your-domain.com/callback/alipay",
        ReturnURL:      "https://your-domain.com/success",
    }
    
    paymentResp, err := client.CreatePayment(createReq)
    if err != nil {
        log.Fatal("创建支付失败:", err)
    }
    
    fmt.Printf("支付创建成功: %+v\n", paymentResp)
    
    // 查询支付状态
    queryResp, err := client.QueryPayment(paymentResp.PaymentID)
    if err != nil {
        log.Fatal("查询支付失败:", err)
    }
    
    fmt.Printf("支付状态: %s\n", queryResp.PaymentStatus)
}
```

## 前端集成示例

### 1. JavaScript/Vue.js 示例

```javascript
// payment.js
class PaymentAPI {
    constructor(baseURL, token) {
        this.baseURL = baseURL;
        this.token = token;
    }
    
    async createPayment(orderData) {
        const response = await fetch(`${this.baseURL}/api/v1/payments`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.token}`
            },
            body: JSON.stringify(orderData)
        });
        
        const result = await response.json();
        
        if (result.code !== 200) {
            throw new Error(result.message);
        }
        
        return result.data;
    }
    
    async queryPayment(paymentId) {
        const response = await fetch(`${this.baseURL}/api/v1/payments/${paymentId}`, {
            headers: {
                'Authorization': `Bearer ${this.token}`
            }
        });
        
        const result = await response.json();
        
        if (result.code !== 200) {
            throw new Error(result.message);
        }
        
        return result.data;
    }
    
    async getPaymentMethods() {
        const response = await fetch(`${this.baseURL}/api/v1/payments/methods`);
        const result = await response.json();
        
        if (result.code !== 200) {
            throw new Error(result.message);
        }
        
        return result.data;
    }
}

// Vue.js 组件示例
export default {
    data() {
        return {
            paymentMethods: [],
            selectedMethod: '',
            orderAmount: 99.99,
            paymentStatus: 'pending',
            qrCodeUrl: '',
            paymentId: null
        }
    },
    
    async mounted() {
        await this.loadPaymentMethods();
    },
    
    methods: {
        async loadPaymentMethods() {
            try {
                const api = new PaymentAPI('/api', this.$store.state.token);
                this.paymentMethods = await api.getPaymentMethods();
            } catch (error) {
                this.$message.error('加载支付方式失败: ' + error.message);
            }
        },
        
        async createPayment() {
            if (!this.selectedMethod) {
                this.$message.warning('请选择支付方式');
                return;
            }
            
            try {
                const api = new PaymentAPI('/api', this.$store.state.token);
                const paymentData = await api.createPayment({
                    order_id: this.$route.params.orderId,
                    payment_method: this.selectedMethod,
                    amount: this.orderAmount.toString(),
                    subject: '商品购买',
                    description: '购买商品',
                    expired_minutes: 30
                });
                
                this.paymentId = paymentData.payment_id;
                
                // 处理不同支付方式的响应
                if (this.selectedMethod === 'alipay') {
                    this.qrCodeUrl = paymentData.payment_data.qr_code;
                    this.showQRCode();
                } else if (this.selectedMethod === 'wechat') {
                    this.qrCodeUrl = paymentData.payment_data.code_url;
                    this.showQRCode();
                }
                
                // 开始轮询支付状态
                this.startPolling();
                
            } catch (error) {
                this.$message.error('创建支付失败: ' + error.message);
            }
        },
        
        showQRCode() {
            // 使用 qrcode.js 生成二维码
            const qr = qrcode(0, 'M');
            qr.addData(this.qrCodeUrl);
            qr.make();
            
            document.getElementById('qrcode').innerHTML = qr.createImgTag(4);
        },
        
        startPolling() {
            const pollInterval = setInterval(async () => {
                try {
                    const api = new PaymentAPI('/api', this.$store.state.token);
                    const payment = await api.queryPayment(this.paymentId);
                    
                    this.paymentStatus = payment.payment_status;
                    
                    if (payment.payment_status === 'success') {
                        clearInterval(pollInterval);
                        this.$message.success('支付成功！');
                        this.$router.push('/payment/success');
                    } else if (payment.payment_status === 'failed') {
                        clearInterval(pollInterval);
                        this.$message.error('支付失败！');
                    }
                } catch (error) {
                    console.error('查询支付状态失败:', error);
                }
            }, 3000); // 每3秒查询一次
            
            // 5分钟后停止轮询
            setTimeout(() => {
                clearInterval(pollInterval);
            }, 300000);
        }
    }
}
```

### 2. React 示例

```jsx
import React, { useState, useEffect } from 'react';
import { message } from 'antd';

const PaymentComponent = ({ orderId, orderAmount }) => {
    const [paymentMethods, setPaymentMethods] = useState([]);
    const [selectedMethod, setSelectedMethod] = useState('');
    const [paymentStatus, setPaymentStatus] = useState('pending');
    const [qrCodeUrl, setQrCodeUrl] = useState('');
    const [paymentId, setPaymentId] = useState(null);
    
    useEffect(() => {
        loadPaymentMethods();
    }, []);
    
    const loadPaymentMethods = async () => {
        try {
            const response = await fetch('/api/v1/payments/methods');
            const result = await response.json();
            
            if (result.code === 200) {
                setPaymentMethods(result.data);
            }
        } catch (error) {
            message.error('加载支付方式失败');
        }
    };
    
    const createPayment = async () => {
        if (!selectedMethod) {
            message.warning('请选择支付方式');
            return;
        }
        
        try {
            const response = await fetch('/api/v1/payments', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify({
                    order_id: orderId,
                    payment_method: selectedMethod,
                    amount: orderAmount.toString(),
                    subject: '商品购买',
                    expired_minutes: 30
                })
            });
            
            const result = await response.json();
            
            if (result.code === 200) {
                setPaymentId(result.data.payment_id);
                
                if (selectedMethod === 'alipay' || selectedMethod === 'wechat') {
                    setQrCodeUrl(result.data.payment_data.qr_code || result.data.payment_data.code_url);
                }
                
                startPolling(result.data.payment_id);
            } else {
                message.error(result.message);
            }
        } catch (error) {
            message.error('创建支付失败');
        }
    };
    
    const startPolling = (paymentId) => {
        const pollInterval = setInterval(async () => {
            try {
                const response = await fetch(`/api/v1/payments/${paymentId}`, {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    }
                });
                
                const result = await response.json();
                
                if (result.code === 200) {
                    setPaymentStatus(result.data.payment_status);
                    
                    if (result.data.payment_status === 'success') {
                        clearInterval(pollInterval);
                        message.success('支付成功！');
                        // 跳转到成功页面
                    } else if (result.data.payment_status === 'failed') {
                        clearInterval(pollInterval);
                        message.error('支付失败！');
                    }
                }
            } catch (error) {
                console.error('查询支付状态失败:', error);
            }
        }, 3000);
        
        // 5分钟后停止轮询
        setTimeout(() => clearInterval(pollInterval), 300000);
    };
    
    return (
        <div className="payment-component">
            <h3>选择支付方式</h3>
            <div className="payment-methods">
                {paymentMethods.map(method => (
                    <div key={method.payment_method} className="payment-method">
                        <input
                            type="radio"
                            id={method.payment_method}
                            name="payment_method"
                            value={method.payment_method}
                            onChange={(e) => setSelectedMethod(e.target.value)}
                        />
                        <label htmlFor={method.payment_method}>
                            <img src={method.icon} alt={method.display_name} />
                            {method.display_name}
                        </label>
                    </div>
                ))}
            </div>
            
            <button onClick={createPayment} disabled={!selectedMethod}>
                立即支付 ¥{orderAmount}
            </button>
            
            {qrCodeUrl && (
                <div className="qr-code">
                    <h4>请使用{selectedMethod === 'alipay' ? '支付宝' : '微信'}扫码支付</h4>
                    <img src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrCodeUrl)}`} alt="支付二维码" />
                    <p>支付状态: {paymentStatus}</p>
                </div>
            )}
        </div>
    );
};

export default PaymentComponent;
```

## 错误处理

### 常见错误及处理方式

```go
func handlePaymentError(err error) {
    switch {
    case strings.Contains(err.Error(), "INVALID_PAYMENT_METHOD"):
        // 无效的支付方式
        log.Println("请检查支付方式配置")
        
    case strings.Contains(err.Error(), "PAYMENT_EXPIRED"):
        // 支付已过期
        log.Println("支付已过期，请重新创建支付")
        
    case strings.Contains(err.Error(), "INSUFFICIENT_AMOUNT"):
        // 金额不足
        log.Println("支付金额不足")
        
    case strings.Contains(err.Error(), "PAYMENT_NOT_FOUND"):
        // 支付记录不存在
        log.Println("支付记录不存在")
        
    default:
        log.Printf("未知错误: %v", err)
    }
}
```

## 最佳实践

### 1. 安全建议

- 始终验证回调签名
- 使用HTTPS传输敏感数据
- 实现幂等性检查
- 设置合理的超时时间
- 记录详细的操作日志

### 2. 性能优化

- 使用连接池
- 实现缓存机制
- 异步处理回调
- 合理设置重试策略

### 3. 监控告警

```go
// 监控支付成功率
func monitorPaymentSuccessRate() {
    // 统计最近1小时的支付成功率
    successRate := calculateSuccessRate(time.Hour)
    
    if successRate < 0.95 { // 成功率低于95%
        sendAlert("支付成功率过低", successRate)
    }
}

// 监控支付响应时间
func monitorPaymentResponseTime() {
    // 监控平均响应时间
    avgResponseTime := calculateAvgResponseTime()
    
    if avgResponseTime > 5*time.Second {
        sendAlert("支付响应时间过长", avgResponseTime)
    }
}
```

### 4. 测试建议

- 编写完整的单元测试
- 进行集成测试
- 使用沙箱环境测试
- 模拟各种异常情况
- 压力测试验证性能

---

更多详细信息请参考 [支付系统文档](./README.md)。
