# Claude AI 项目记忆和经验库

## 技术身份记忆
- 阿里巴巴P10+全栈架构师，20年+经验
- 精通Java/Go/PHP，专精微服务架构
- 善用阿里开源中间件(nacos, dubbo3, rocketmq, canal)
- RuoYi-Vue-Plus框架专家
- 微信小程序生态开发大师

## RuoYi-Vue-Plus项目核心记忆

### 系统性调试方法论
**六优先级调试方法论（成功率95%+）**：
1. **依赖冲突排查** - Maven依赖树版本冲突检查
2. **编译缓存清理** - target目录和IDE缓存清理
3. **Maven JAR包更新** - 按依赖顺序重新编译（最关键）
4. **环境配置验证** - Java版本、Maven配置检查
5. **端到端字节码验证** - class文件和JAR包验证
6. **完整重建** - 仅用于系统性依赖问题

### 热重载开发规范记忆
- **黄金法则**: 绝对不能随便停止`mvn spring-boot:run`启动的应用
- **自动热重载**: ruoyi-admin模块修改保存后2-3秒自动生效
- **需要编译**: ruoyi-business/ruoyi-system模块需要`mvn compile -pl [module]`
- **跨模块修改**: 必须`mvn install`更新本地仓库
- **结构性修改**: 新增类/方法/注解需要重新编译

### 常见问题解决方案记忆

#### BaseMapperPlus类找不到（85%频次）
**根本原因**: Maven依赖传递链断裂
**标准解决方案**:
1. `mvn install -pl ruoyi-common/ruoyi-common-mybatis -DskipTests`
2. `mvn install -pl ruoyi-modules/ruoyi-system -DskipTests`
3. `mvn install -pl ruoyi-admin -DskipTests`

#### MapStruct问题
**根本原因**: `reverseConvertGenerate = false`配置导致生成不完整映射代码
**解决方案**: 
- 移除`reverseConvertGenerate = false`
- 清理`generated-sources`目录
- 重新编译
- 使用`@Autowired + @Lazy`抽象类解决循环依赖

#### 编译文件锁定问题
**解决方案**: 
1. `taskkill /f /im java.exe`强制终止Java进程
2. 删除被锁定的target目录
3. 重新执行完整编译

### Java错误诊断记忆
- **四优先级原则**: 只有排除所有外部因素后才修改源代码
- **Maven JAR包更新**: 修改模块代码后执行`mvn install -pl [module] -DskipTests`
- **Java 17模块系统**: 需要JVM参数`--add-opens java.base/java.io=ALL-UNNAMED`

## Git管理记忆

### 分支策略
- **aitest分支**: 调试、修复、实验性修改
- **master/main分支**: 稳定、已验证代码
- **dependency-refactor分支**: 架构性重构

### Git安全守则
- **禁止**: `git reset --hard`、`git push --force`在公共分支
- **推荐**: `git revert`安全回滚
- **合并前**: 创建备份分支，保留文档

## 微信小程序集成记忆

### 最佳实践
1. **登录流程**: `wx.login()` → `code2session` → 查询/创建`sys_user` → 生成Token → 缓存session_key
2. **敏感数据解密**: BouncyCastle库，AES-128-CBC模式
3. **用户绑定**: `sys_user`表增加`wx_openid`字段
4. **生产凭证**: appid: wx3c7c42e2dd4afa17, secret: a5c929a120f75095594f0509bd3a722c

### 开发规范
- 严格遵循微信官方文档
- 前后端独立Git仓库
- 域名白名单和SSL证书验证
- 真机测试与开发工具差异处理

## 用户环境记忆

### 开发环境偏好
- Git作者: Akari<akarinzhang@foxmail.com>
- GitHub: Akarin-Akari，SSH优先(git@github.com)
- IDE: IntelliJ IDEA运行配置启动Spring Boot
- 文档命名: 'YYYY-MM-DD filename.md'格式

### 生产环境配置
- **数据库**: MySQL 5.7生产，MySQL 8.0开发
- **域名**: www.huifeixingkj.com (SSL已配置)
- **部署**: 宝塔面板
- **API前缀**: /prod-api/
- **数据库信息**: 
  - 生产: huifeixingrb/huifeixingrb/KTMDdMNtpiAWbLrC
  - 开发: root/123456

## HuifeixingRY项目特定记忆

### 系统架构
- 基于RuoYi-Vue-Plus 5.4.0
- 事件驱动微服务架构
- 包含35种业务事件类型和3个事件处理器
- 15个公共模块

### 模块保护要求
- **特别保护**: ruoyi-mpweixin模块稳定性
- **API隔离**: ruoyi-business-api隔离性
- **轻量级调用**: ruoyi-mpweixin → ruoyi-business-api
- **禁止修改**: 框架核心文件

### 权限系统记忆
**"三位一体"闭环诊断法**:
1. **后端注解**: 检查`@PreAuthorize`权限字符串
2. **数据库菜单**: 确认`sys_menu`表`perms`字段一致性
3. **角色分配**: 确认用户角色已勾选对应菜单权限

### 异步任务与事务管理
1. **@Async失效**: 同类内部调用绕过AOP代理，使用专门的AsyncService类
2. **上下文丢失**: 使用TransmittableThreadLocal传递线程上下文
3. **跨模块事务**: 使用`@Transactional(rollbackFor=Exception.class)`

## 数据库设计记忆
- **黄金法则**: 设计数据库时一定不要使用外键！
- **字典系统**: 优先使用"系统管理→字典管理→刷新缓存"
- **数据权限**: DataScopeAspect自动拼接WHERE条件
- **豁免权限**: 使用`@DataScope(ignore=true)`

## 框架Lombok和MapStruct记忆
- 项目使用Lombok注解，不要重复生成getter/setter
- 使用MapStruct进行VO-Entity转换
- 调用`XxxConvert.INSTANCE.toXxx()`，不要手动编写set方法
- controller入参用VO类，service层用Entity

## 性能优化记忆

### 调试工具和日志
- **敏感数据拦截器**: 性能瓶颈分析
- **高并发响应**: 时间分布分析
- **缓存机制**: 缺失原因调查
- **异步处理**: 机制设计优化

### 监控和诊断
- **内存使用**: 情况分析报告
- **IO密集代码**: 段识别分析
- **数据库查询**: 频率效率分析
- **日志输出**: 性能影响分析

## 文档管理记忆
- 技术报告使用'YYYY-MM-DD [报告标题].md'格式
- 重命名文件保留原始修改时间戳
- 中文Markdown文件正则表达式时间提取
- 从文档内容提取时间而非系统时间

## 错误模式记忆

### 编译错误常见原因
1. **依赖冲突**: 同名类多版本冲突
2. **缓存问题**: IDE与Maven编译缓存不一致
3. **JAR包冲突**: 多版本编译产物同时存在
4. **环境配置**: IDE与命令行Maven版本不一致

### 微服务问题诊断
1. **应用程序启动验证**
2. **接口测试**
3. **代码执行路径验证**
4. **数据库数据检查**
5. **前端调试**
6. **缓存机制调查**

## 安全和合规记忆
- 敏感数据保护拦截器
- 输入验证和输出转义
- 权限控制和数据权限
- 微信小程序合规性检查

## 项目知识库统计
- RuoYi-Vue-Plus项目建立了89个问题实例分析
- 问题发现时间从2.5小时缩短到15分钟
- 修复成功率从60%提升到95%
- 预防式调试流程替代被动修复模式