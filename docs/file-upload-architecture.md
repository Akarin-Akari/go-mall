# Mall-Go文件上传系统架构设计

## 📋 **系统概述**

Mall-Go文件上传系统是一个企业级的文件管理解决方案，支持商品图片、用户头像等多种文件类型的安全上传、存储和访问。

## 🏗️ **整体架构**

### **分层架构设计**
```
┌─────────────────────────────────────────────────────────────┐
│                    API接口层 (Handler)                        │
├─────────────────────────────────────────────────────────────┤
│                    业务逻辑层 (Service)                       │
├─────────────────────────────────────────────────────────────┤
│                    文件处理层 (FileProcessor)                 │
├─────────────────────────────────────────────────────────────┤
│                    存储适配层 (StorageAdapter)                │
├─────────────────────────────────────────────────────────────┤
│                    物理存储层 (LocalStorage/CloudStorage)     │
└─────────────────────────────────────────────────────────────┘
```

### **核心组件**

1. **FileUploadHandler** - HTTP请求处理
2. **FileUploadService** - 业务逻辑处理
3. **FileValidator** - 文件验证器
4. **StorageManager** - 存储管理器
5. **FileMetadata** - 文件元数据管理

## 🔧 **技术选型**

### **存储策略**
- **本地存储**: 开发和测试环境
- **云存储**: 生产环境（可选：阿里云OSS、腾讯云COS、AWS S3）
- **CDN加速**: 静态资源分发

### **支持的文件类型**
```go
// 图片文件
ImageTypes = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

// 文档文件（可选）
DocumentTypes = []string{".pdf", ".doc", ".docx"}

// 视频文件（可选）
VideoTypes = []string{".mp4", ".avi", ".mov"}
```

### **安全限制**
- **文件大小**: 图片5MB，文档10MB，视频50MB
- **文件数量**: 单次上传最多10个文件
- **MIME类型验证**: 严格验证文件类型
- **文件内容检查**: 防止恶意文件上传

## 📊 **数据模型设计**

### **文件信息表 (files)**
```sql
CREATE TABLE files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(36) UNIQUE NOT NULL COMMENT '文件唯一标识',
    original_name VARCHAR(255) NOT NULL COMMENT '原始文件名',
    stored_name VARCHAR(255) NOT NULL COMMENT '存储文件名',
    file_path VARCHAR(500) NOT NULL COMMENT '文件路径',
    file_size BIGINT NOT NULL COMMENT '文件大小(字节)',
    mime_type VARCHAR(100) NOT NULL COMMENT 'MIME类型',
    file_type ENUM('image', 'document', 'video', 'other') NOT NULL COMMENT '文件类型',
    storage_type ENUM('local', 'oss', 'cos', 's3') DEFAULT 'local' COMMENT '存储类型',
    upload_user_id BIGINT NOT NULL COMMENT '上传用户ID',
    business_type VARCHAR(50) COMMENT '业务类型(avatar, product, etc)',
    business_id BIGINT COMMENT '关联业务ID',
    access_url VARCHAR(500) COMMENT '访问URL',
    is_public BOOLEAN DEFAULT FALSE COMMENT '是否公开访问',
    status ENUM('uploading', 'success', 'failed', 'deleted') DEFAULT 'uploading',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_user_id (upload_user_id),
    INDEX idx_business (business_type, business_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

## 🔐 **权限控制设计**

### **权限资源定义**
```go
const (
    ResourceFile = "file"
    
    ActionUpload = "upload"   // 上传文件
    ActionView   = "view"     // 查看文件
    ActionDelete = "delete"   // 删除文件
    ActionManage = "manage"   // 管理文件
)
```

### **权限规则**
- **普通用户**: 可以上传和查看自己的文件
- **商家**: 可以上传商品图片，管理店铺相关文件
- **管理员**: 可以管理所有文件

## 🛡️ **安全设计**

### **文件验证流程**
1. **大小验证**: 检查文件大小是否超限
2. **类型验证**: 验证文件扩展名和MIME类型
3. **内容验证**: 检查文件头部魔数
4. **恶意检测**: 扫描潜在的恶意内容
5. **重命名存储**: 使用UUID重命名，防止路径遍历

### **访问控制**
- **权限验证**: 基于Casbin的访问控制
- **防盗链**: Referer检查和Token验证
- **访问日志**: 记录所有文件访问行为

## 📁 **存储结构设计**

### **本地存储目录结构**
```
uploads/
├── images/
│   ├── avatars/
│   │   └── 2024/01/
│   │       └── uuid.jpg
│   ├── products/
│   │   └── 2024/01/
│   │       └── uuid.jpg
│   └── temp/
│       └── uploading_files
├── documents/
│   └── 2024/01/
│       └── uuid.pdf
└── videos/
    └── 2024/01/
        └── uuid.mp4
```

### **URL访问模式**
```
# 公开访问
GET /api/v1/files/public/{uuid}

# 私有访问（需要认证）
GET /api/v1/files/private/{uuid}

# 缩略图访问
GET /api/v1/files/thumbnail/{uuid}?size=small|medium|large
```

## 🚀 **性能优化**

### **上传优化**
- **分片上传**: 大文件分片上传，支持断点续传
- **并发上传**: 多文件并发处理
- **压缩处理**: 图片自动压缩和格式转换
- **缓存策略**: 文件元数据缓存

### **访问优化**
- **CDN分发**: 静态资源CDN加速
- **缓存控制**: HTTP缓存头设置
- **懒加载**: 图片懒加载支持
- **多尺寸**: 自动生成多种尺寸缩略图

## 📈 **监控和日志**

### **关键指标**
- 上传成功率
- 平均上传时间
- 存储空间使用率
- 访问频率统计

### **日志记录**
- 文件上传日志
- 访问请求日志
- 错误异常日志
- 安全事件日志

## 🔄 **扩展性设计**

### **存储扩展**
- 支持多种存储后端
- 存储策略可配置
- 自动故障转移

### **功能扩展**
- 图片处理服务
- 视频转码服务
- 文件预览服务
- 批量操作API

## 📋 **实施计划**

### **第一阶段**: 基础功能
- 本地存储实现
- 基本文件上传
- 权限控制集成

### **第二阶段**: 增强功能
- 云存储支持
- 图片处理
- 缩略图生成

### **第三阶段**: 高级功能
- 分片上传
- CDN集成
- 监控告警

## 🎯 **成功标准**

- ✅ 支持常见文件类型上传
- ✅ 完整的权限控制
- ✅ 安全的文件验证
- ✅ 高性能的文件访问
- ✅ 完善的错误处理
- ✅ 详细的操作日志
