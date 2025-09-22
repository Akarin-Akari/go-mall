package errors

// 用户模块错误定义
var (
	// 用户注册错误 (21000-21099)
	ErrUserRegistrationFailed = NewBusinessError("21001", "用户注册失败")
	ErrUserEmailExists        = NewBusinessError("21002", "邮箱已存在")
	ErrUserPhoneExists        = NewBusinessError("21003", "手机号已存在")
	ErrUserUsernameExists     = NewBusinessError("21004", "用户名已存在")
	ErrUserInvalidEmail       = NewValidationError("30010", "邮箱格式无效")
	ErrUserInvalidPhone       = NewValidationError("30011", "手机号格式无效")
	ErrUserPasswordTooWeak    = NewValidationError("30012", "密码强度不足")

	// 用户登录错误 (21100-21199)
	ErrUserLoginFailed       = NewAuthError("40010", "用户登录失败")
	ErrUserNotFound          = NewAuthError("40011", "用户不存在")
	ErrUserPasswordIncorrect = NewAuthError("40012", "密码错误")
	ErrUserAccountLocked     = NewAuthError("40013", "账户已锁定")
	ErrUserAccountDisabled   = NewAuthError("40014", "账户已禁用")
	ErrUserTooManyAttempts   = NewAuthError("40015", "登录尝试次数过多")

	// 用户权限错误 (21200-21299)
	ErrUserPermissionDenied = NewPermissionError("50010", "用户权限不足")
	ErrUserRoleNotFound     = NewPermissionError("50011", "用户角色不存在")
	ErrUserResourceAccess   = NewPermissionError("50012", "用户资源访问被拒绝")
)

// 商品模块错误定义
var (
	// 商品基础错误 (22000-22099)
	ErrProductNotFound        = NewBusinessError("22001", "商品不存在")
	ErrProductInactive        = NewBusinessError("22002", "商品已下架")
	ErrProductOutOfStock      = NewBusinessError("22003", "商品库存不足")
	ErrProductPriceInvalid    = NewValidationError("30020", "商品价格无效")
	ErrProductNameRequired    = NewValidationError("30021", "商品名称不能为空")
	ErrProductCategoryInvalid = NewValidationError("30022", "商品分类无效")

	// 商品库存错误 (22100-22199)
	ErrInventoryInsufficient      = NewBusinessError("22101", "库存不足")
	ErrInventoryReservationFailed = NewBusinessError("22102", "库存预留失败")
	ErrInventoryReleasesFailed    = NewBusinessError("22103", "库存释放失败")
	ErrInventoryLockTimeout       = NewBusinessError("22104", "库存锁定超时")

	// 商品分类错误 (22200-22299)
	ErrCategoryNotFound    = NewBusinessError("22201", "商品分类不存在")
	ErrCategoryInactive    = NewBusinessError("22202", "商品分类已禁用")
	ErrCategoryHasProducts = NewBusinessError("22203", "分类下存在商品，无法删除")
)

// 订单模块错误定义
var (
	// 订单基础错误 (23000-23099)
	ErrOrderNotFound      = NewBusinessError("23001", "订单不存在")
	ErrOrderInvalidStatus = NewBusinessError("23002", "订单状态无效")
	ErrOrderCannotCancel  = NewBusinessError("23003", "订单无法取消")
	ErrOrderCannotModify  = NewBusinessError("23004", "订单无法修改")
	ErrOrderExpired       = NewBusinessError("23005", "订单已过期")
	ErrOrderAlreadyPaid   = NewBusinessError("23006", "订单已支付")

	// 订单创建错误 (23100-23199)
	ErrOrderCreateFailed     = NewBusinessError("23101", "订单创建失败")
	ErrOrderItemsEmpty       = NewValidationError("30030", "订单商品列表不能为空")
	ErrOrderAmountInvalid    = NewValidationError("30031", "订单金额无效")
	ErrOrderAddressRequired  = NewValidationError("30032", "收货地址不能为空")
	ErrOrderCalculationError = NewBusinessError("23102", "订单金额计算错误")

	// 订单配送错误 (23200-23299)
	ErrShippingNotAvailable       = NewBusinessError("23201", "配送服务不可用")
	ErrShippingAddressInvalid     = NewBusinessError("23202", "配送地址无效")
	ErrShippingMethodNotSupported = NewBusinessError("23203", "不支持的配送方式")
)

// 支付模块错误定义
var (
	// 支付基础错误 (已在主错误码中定义，这里提供便捷方法)
	ErrPaymentNotFound          = NewPaymentError("90001", "支付记录不存在")
	ErrPaymentMethodInvalid     = NewPaymentError("90002", "支付方式无效")
	ErrPaymentAmountInvalid     = NewPaymentError("90003", "支付金额无效")
	ErrPaymentInsufficientFunds = NewPaymentError("90004", "余额不足")
	ErrPaymentExpired           = NewPaymentError("90005", "支付已过期")
	ErrPaymentAlreadyCompleted  = NewPaymentError("90006", "支付已完成")

	// 支付渠道错误 (91000-91999)
	ErrAlipayConfigInvalid    = NewPaymentError("91001", "支付宝配置无效")
	ErrAlipaySignatureInvalid = NewPaymentError("91002", "支付宝签名验证失败")
	ErrWechatConfigInvalid    = NewPaymentError("91003", "微信支付配置无效")
	ErrWechatSignatureInvalid = NewPaymentError("91004", "微信支付签名验证失败")
	ErrUnionPayUnavailable    = NewPaymentError("91005", "银联支付不可用")

	// 退款错误 (92000-92999)
	ErrRefundNotAllowed       = NewPaymentError("92001", "不允许退款")
	ErrRefundAmountExceed     = NewPaymentError("92002", "退款金额超过支付金额")
	ErrRefundProcessingFailed = NewPaymentError("92003", "退款处理失败")
	ErrRefundAlreadyProcessed = NewPaymentError("92004", "退款已处理")
)

// 购物车模块错误定义
var (
	// 购物车基础错误 (24000-24099)
	ErrCartNotFound            = NewBusinessError("24001", "购物车不存在")
	ErrCartItemNotFound        = NewBusinessError("24002", "购物车商品不存在")
	ErrCartEmpty               = NewBusinessError("24003", "购物车为空")
	ErrCartItemQuantityInvalid = NewValidationError("30040", "购物车商品数量无效")
	ErrCartItemAlreadyExists   = NewBusinessError("24004", "商品已在购物车中")
	ErrCartSyncFailed          = NewBusinessError("24005", "购物车同步失败")
)

// 文件上传模块错误定义
var (
	// 文件上传基础错误 (25000-25099)
	ErrFileUploadFailed     = NewError("25001", "文件上传失败", ErrorLevelError, CategoryUpload)
	ErrFileTypeNotSupported = NewValidationError("30050", "不支持的文件类型")
	ErrFileSizeExceed       = NewValidationError("30051", "文件大小超过限制")
	ErrFileNameInvalid      = NewValidationError("30052", "文件名无效")
	ErrFileContentInvalid   = NewValidationError("30053", "文件内容无效")
	ErrFileStorageFailure   = NewSystemError("10010", "文件存储失败")

	// 图片处理错误 (25100-25199)
	ErrImageProcessingFailed  = NewError("25101", "图片处理失败", ErrorLevelError, CategoryUpload)
	ErrImageFormatInvalid     = NewValidationError("30060", "图片格式无效")
	ErrImageSizeInvalid       = NewValidationError("30061", "图片尺寸无效")
	ErrImageCompressionFailed = NewError("25102", "图片压缩失败", ErrorLevelError, CategoryUpload)
)

// 系统级错误定义
var (
	// 配置错误 (10100-10199)
	ErrConfigNotFound   = NewSystemError("10101", "配置不存在")
	ErrConfigInvalid    = NewSystemError("10102", "配置无效")
	ErrConfigLoadFailed = NewSystemError("10103", "配置加载失败")

	// 数据库错误 (扩展)
	ErrDatabaseConnectionFailed  = NewDatabaseError("60010", "数据库连接失败")
	ErrDatabaseQueryTimeout      = NewDatabaseError("60011", "数据库查询超时")
	ErrDatabaseTransactionFailed = NewDatabaseError("60012", "数据库事务失败")
	ErrDatabaseMigrationFailed   = NewDatabaseError("60013", "数据库迁移失败")

	// 缓存错误 (10200-10299)
	ErrCacheConnectionFailed = NewSystemError("10201", "缓存连接失败")
	ErrCacheOperationFailed  = NewSystemError("10202", "缓存操作失败")
	ErrCacheKeyNotFound      = NewSystemError("10203", "缓存键不存在")
	ErrCacheSerialization    = NewSystemError("10204", "缓存序列化失败")

	// 消息队列错误 (10300-10399)
	ErrMQConnectionFailed = NewSystemError("10301", "消息队列连接失败")
	ErrMQPublishFailed    = NewSystemError("10302", "消息发布失败")
	ErrMQConsumeError     = NewSystemError("10303", "消息消费错误")
	ErrMQTimeout          = NewSystemError("10304", "消息队列超时")
)

// 第三方服务错误定义
var (
	// 短信服务错误 (81000-81099)
	ErrSMSServiceUnavailable = NewThirdPartyError("81001", "短信服务不可用")
	ErrSMSSendFailed         = NewThirdPartyError("81002", "短信发送失败")
	ErrSMSTemplateInvalid    = NewThirdPartyError("81003", "短信模板无效")
	ErrSMSQuotaExceed        = NewThirdPartyError("81004", "短信配额超限")

	// 邮件服务错误 (81100-81199)
	ErrEmailServiceUnavailable = NewThirdPartyError("81101", "邮件服务不可用")
	ErrEmailSendFailed         = NewThirdPartyError("81102", "邮件发送失败")
	ErrEmailTemplateInvalid    = NewThirdPartyError("81103", "邮件模板无效")

	// 对象存储错误 (81200-81299)
	ErrOSSServiceUnavailable = NewThirdPartyError("81201", "对象存储服务不可用")
	ErrOSSUploadFailed       = NewThirdPartyError("81202", "文件上传到OSS失败")
	ErrOSSDownloadFailed     = NewThirdPartyError("81203", "从OSS下载文件失败")
	ErrOSSDeleteFailed       = NewThirdPartyError("81204", "从OSS删除文件失败")

	// CDN服务错误 (81300-81399)
	ErrCDNServiceUnavailable = NewThirdPartyError("81301", "CDN服务不可用")
	ErrCDNPushFailed         = NewThirdPartyError("81302", "CDN推送失败")
	ErrCDNPurgeFailed        = NewThirdPartyError("81303", "CDN清除失败")
)

// 业务规则错误定义
var (
	// 促销活动错误 (26000-26099)
	ErrPromotionNotFound        = NewBusinessError("26001", "促销活动不存在")
	ErrPromotionExpired         = NewBusinessError("26002", "促销活动已过期")
	ErrPromotionQuotaExceed     = NewBusinessError("26003", "促销活动参与次数超限")
	ErrPromotionConditionNotMet = NewBusinessError("26004", "不满足促销活动条件")

	// 优惠券错误 (26100-26199)
	ErrCouponNotFound        = NewBusinessError("26101", "优惠券不存在")
	ErrCouponExpired         = NewBusinessError("26102", "优惠券已过期")
	ErrCouponUsed            = NewBusinessError("26103", "优惠券已使用")
	ErrCouponMinAmountNotMet = NewBusinessError("26104", "未达到优惠券最低使用金额")

	// 积分错误 (26200-26299)
	ErrPointsInsufficient     = NewBusinessError("26201", "积分不足")
	ErrPointsExpired          = NewBusinessError("26202", "积分已过期")
	ErrPointsCalculationError = NewBusinessError("26203", "积分计算错误")
)

// GetBusinessErrorByCode 根据错误码获取预定义的业务错误
func GetBusinessErrorByCode(code ErrorCode) *BusinessError {
	errorMap := map[ErrorCode]*BusinessError{
		// 用户错误
		"21001": ErrUserRegistrationFailed,
		"21002": ErrUserEmailExists,
		"21003": ErrUserPhoneExists,
		"21004": ErrUserUsernameExists,
		"40010": ErrUserLoginFailed,
		"40011": ErrUserNotFound,
		"40012": ErrUserPasswordIncorrect,

		// 商品错误
		"22001": ErrProductNotFound,
		"22002": ErrProductInactive,
		"22003": ErrProductOutOfStock,

		// 订单错误
		"23001": ErrOrderNotFound,
		"23002": ErrOrderInvalidStatus,
		"23003": ErrOrderCannotCancel,

		// 支付错误
		"90001": ErrPaymentNotFound,
		"90002": ErrPaymentMethodInvalid,
		"90003": ErrPaymentAmountInvalid,

		// 系统错误
		"10001": NewSystemError("10001", "系统内部错误"),
		"10002": NewSystemError("10002", "系统超时"),
		"10003": NewSystemError("10003", "系统过载"),
	}

	if err, exists := errorMap[code]; exists {
		// 返回错误的副本，避免修改原始错误
		newErr := *err
		newErr.Timestamp = time.Now()
		newErr.StackTrace = captureStackTrace(1)
		return &newErr
	}

	// 如果没有找到预定义错误，返回通用错误
	return NewSystemError("10001", "未知错误").WithDetails(fmt.Sprintf("错误码: %s", code))
}
