# 第1章：前端架构设计原则 🏗️

> *"优秀的架构不是设计出来的，而是演进出来的！"* 🚀

## 📚 本章导览

前端架构设计是构建大型、可维护、可扩展Web应用的基石。随着业务复杂度的增加，如何设计一个既能满足当前需求，又能适应未来变化的前端架构，成为了每个前端工程师必须掌握的核心技能。本章将从架构设计的基本原则出发，结合Mall-Frontend项目的实际案例，深入探讨现代前端架构的设计理念和最佳实践。

### 🎯 学习目标

通过本章学习，你将掌握：

- **架构设计原则** - 理解SOLID、DRY、KISS等设计原则在前端的应用
- **分层架构模式** - 掌握表现层、业务层、数据层的职责划分
- **模块化设计** - 学会组件化、模块化的设计方法和最佳实践
- **依赖管理** - 理解依赖注入、控制反转在前端的实现
- **可扩展性设计** - 掌握插件化、微前端等扩展性架构模式
- **性能架构** - 学会从架构层面优化应用性能
- **架构演进策略** - 理解架构重构和渐进式升级方法
- **企业级实践** - 大型团队的前端架构治理经验

### 🛠️ 技术栈概览

```typescript
{
  "architecturePatterns": {
    "layered": ["Presentation", "Business", "Data", "Infrastructure"],
    "modular": ["Feature Modules", "Shared Modules", "Core Modules"],
    "component": ["Atomic Design", "Container/Presenter", "Compound Components"]
  },
  "designPrinciples": {
    "solid": ["SRP", "OCP", "LSP", "ISP", "DIP"],
    "others": ["DRY", "KISS", "YAGNI", "Composition over Inheritance"]
  },
  "scalabilityPatterns": {
    "microfrontends": ["Module Federation", "Single-SPA", "Micro Apps"],
    "plugins": ["Plugin Architecture", "Event-Driven", "Hook System"]
  },
  "tools": ["TypeScript", "ESLint", "Prettier", "Husky", "Nx", "Lerna"]
}
```

### 📖 本章目录

- [架构设计基本原则](#架构设计基本原则)
- [分层架构模式](#分层架构模式)
- [模块化与组件化设计](#模块化与组件化设计)
- [依赖管理与控制反转](#依赖管理与控制反转)
- [可扩展性架构设计](#可扩展性架构设计)
- [性能导向的架构设计](#性能导向的架构设计)
- [架构治理与演进](#架构治理与演进)
- [企业级架构实践](#企业级架构实践)
- [Mall-Frontend架构分析](#mall-frontend架构分析)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🎯 架构设计基本原则

### SOLID原则在前端的应用

SOLID原则是面向对象设计的五大基本原则，在前端开发中同样具有重要的指导意义：

```typescript
// 1. 单一职责原则 (Single Responsibility Principle)
// ❌ 违反SRP的组件 - 职责过多
class UserProfileComponent {
  // 用户数据管理
  private userData: User;
  
  // UI渲染
  render() { /* ... */ }
  
  // 数据验证
  validateUserData() { /* ... */ }
  
  // API调用
  async saveUser() { /* ... */ }
  
  // 权限检查
  checkPermissions() { /* ... */ }
  
  // 日志记录
  logUserAction() { /* ... */ }
}

// ✅ 遵循SRP的设计 - 职责分离
// 用户数据管理
class UserDataManager {
  private userData: User;
  
  getUserData(): User { return this.userData; }
  setUserData(data: User): void { this.userData = data; }
}

// 数据验证服务
class UserValidator {
  validate(userData: User): ValidationResult {
    // 专注于数据验证逻辑
    return { isValid: true, errors: [] };
  }
}

// API服务
class UserApiService {
  async saveUser(userData: User): Promise<User> {
    // 专注于API调用
    return fetch('/api/users', {
      method: 'POST',
      body: JSON.stringify(userData)
    }).then(res => res.json());
  }
}

// 权限服务
class PermissionService {
  checkUserPermission(user: User, action: string): boolean {
    // 专注于权限检查逻辑
    return user.permissions.includes(action);
  }
}

// UI组件 - 只负责渲染
function UserProfileComponent({ 
  userManager, 
  validator, 
  apiService, 
  permissionService 
}: {
  userManager: UserDataManager;
  validator: UserValidator;
  apiService: UserApiService;
  permissionService: PermissionService;
}) {
  const [user, setUser] = useState(userManager.getUserData());
  
  const handleSave = async () => {
    const validation = validator.validate(user);
    if (validation.isValid) {
      await apiService.saveUser(user);
    }
  };
  
  return (
    <div>
      {/* 专注于UI渲染 */}
      <UserForm user={user} onSave={handleSave} />
    </div>
  );
}
```

```typescript
// 2. 开放封闭原则 (Open/Closed Principle)
// 对扩展开放，对修改封闭

// ❌ 违反OCP的设计 - 每次添加新类型都要修改原有代码
class NotificationManager {
  sendNotification(type: string, message: string) {
    switch (type) {
      case 'email':
        this.sendEmail(message);
        break;
      case 'sms':
        this.sendSMS(message);
        break;
      case 'push':
        this.sendPush(message);
        break;
      // 每次添加新类型都要修改这里
      default:
        throw new Error('Unsupported notification type');
    }
  }
  
  private sendEmail(message: string) { /* ... */ }
  private sendSMS(message: string) { /* ... */ }
  private sendPush(message: string) { /* ... */ }
}

// ✅ 遵循OCP的设计 - 通过接口扩展
interface NotificationProvider {
  send(message: string): Promise<void>;
}

class EmailProvider implements NotificationProvider {
  async send(message: string): Promise<void> {
    // 邮件发送逻辑
    console.log('Sending email:', message);
  }
}

class SMSProvider implements NotificationProvider {
  async send(message: string): Promise<void> {
    // 短信发送逻辑
    console.log('Sending SMS:', message);
  }
}

class PushProvider implements NotificationProvider {
  async send(message: string): Promise<void> {
    // 推送发送逻辑
    console.log('Sending push:', message);
  }
}

// 新增微信通知 - 无需修改原有代码
class WeChatProvider implements NotificationProvider {
  async send(message: string): Promise<void> {
    console.log('Sending WeChat:', message);
  }
}

class NotificationManager {
  private providers = new Map<string, NotificationProvider>();
  
  // 注册通知提供者
  registerProvider(type: string, provider: NotificationProvider) {
    this.providers.set(type, provider);
  }
  
  // 发送通知 - 无需修改
  async sendNotification(type: string, message: string) {
    const provider = this.providers.get(type);
    if (!provider) {
      throw new Error(`No provider registered for type: ${type}`);
    }
    await provider.send(message);
  }
}

// 使用示例
const notificationManager = new NotificationManager();
notificationManager.registerProvider('email', new EmailProvider());
notificationManager.registerProvider('sms', new SMSProvider());
notificationManager.registerProvider('push', new PushProvider());
notificationManager.registerProvider('wechat', new WeChatProvider()); // 新增类型
```

```typescript
// 3. 里氏替换原则 (Liskov Substitution Principle)
// 子类必须能够替换其基类

// ❌ 违反LSP的设计
abstract class Shape {
  abstract calculateArea(): number;
  abstract setWidth(width: number): void;
  abstract setHeight(height: number): void;
}

class Rectangle extends Shape {
  constructor(private width: number, private height: number) {
    super();
  }
  
  calculateArea(): number {
    return this.width * this.height;
  }
  
  setWidth(width: number): void {
    this.width = width;
  }
  
  setHeight(height: number): void {
    this.height = height;
  }
}

class Square extends Rectangle {
  constructor(side: number) {
    super(side, side);
  }
  
  // 违反LSP - 改变了基类的行为
  setWidth(width: number): void {
    this.width = width;
    this.height = width; // 强制保持正方形
  }
  
  setHeight(height: number): void {
    this.width = height; // 强制保持正方形
    this.height = height;
  }
}

// 这会导致意外的行为
function resizeShape(shape: Rectangle) {
  shape.setWidth(5);
  shape.setHeight(4);
  // 期望面积是20，但如果是Square，面积会是16
  console.log('Area:', shape.calculateArea());
}

// ✅ 遵循LSP的设计
interface Drawable {
  draw(): void;
  calculateArea(): number;
}

class Rectangle implements Drawable {
  constructor(private width: number, private height: number) {}
  
  draw(): void {
    console.log(`Drawing rectangle: ${this.width}x${this.height}`);
  }
  
  calculateArea(): number {
    return this.width * this.height;
  }
  
  setDimensions(width: number, height: number): void {
    this.width = width;
    this.height = height;
  }
}

class Square implements Drawable {
  constructor(private side: number) {}
  
  draw(): void {
    console.log(`Drawing square: ${this.side}x${this.side}`);
  }
  
  calculateArea(): number {
    return this.side * this.side;
  }
  
  setSide(side: number): void {
    this.side = side;
  }
}

// 现在可以安全地替换
function drawShapes(shapes: Drawable[]) {
  shapes.forEach(shape => {
    shape.draw();
    console.log('Area:', shape.calculateArea());
  });
}
```

```typescript
// 4. 接口隔离原则 (Interface Segregation Principle)
// 客户端不应该依赖它不需要的接口

// ❌ 违反ISP的设计 - 臃肿的接口
interface Worker {
  work(): void;
  eat(): void;
  sleep(): void;
  code(): void;
  design(): void;
  test(): void;
  deploy(): void;
}

class Developer implements Worker {
  work(): void { console.log('Working...'); }
  eat(): void { console.log('Eating...'); }
  sleep(): void { console.log('Sleeping...'); }
  code(): void { console.log('Coding...'); }
  design(): void { throw new Error('Developers do not design'); } // 不需要
  test(): void { console.log('Testing...'); }
  deploy(): void { console.log('Deploying...'); }
}

// ✅ 遵循ISP的设计 - 接口分离
interface Workable {
  work(): void;
}

interface Eatable {
  eat(): void;
}

interface Sleepable {
  sleep(): void;
}

interface Codeable {
  code(): void;
}

interface Designable {
  design(): void;
}

interface Testable {
  test(): void;
}

interface Deployable {
  deploy(): void;
}

// 开发者只实现需要的接口
class Developer implements Workable, Eatable, Sleepable, Codeable, Testable, Deployable {
  work(): void { console.log('Working...'); }
  eat(): void { console.log('Eating...'); }
  sleep(): void { console.log('Sleeping...'); }
  code(): void { console.log('Coding...'); }
  test(): void { console.log('Testing...'); }
  deploy(): void { console.log('Deploying...'); }
}

// 设计师只实现需要的接口
class Designer implements Workable, Eatable, Sleepable, Designable {
  work(): void { console.log('Working...'); }
  eat(): void { console.log('Eating...'); }
  sleep(): void { console.log('Sleeping...'); }
  design(): void { console.log('Designing...'); }
}
```

```typescript
// 5. 依赖倒置原则 (Dependency Inversion Principle)
// 高层模块不应该依赖低层模块，两者都应该依赖抽象

// ❌ 违反DIP的设计 - 高层模块依赖具体实现
class MySQLDatabase {
  save(data: any): void {
    console.log('Saving to MySQL:', data);
  }
  
  find(id: string): any {
    console.log('Finding in MySQL:', id);
    return { id, name: 'User' };
  }
}

class UserService {
  private database = new MySQLDatabase(); // 直接依赖具体实现
  
  createUser(userData: any): void {
    this.database.save(userData);
  }
  
  getUser(id: string): any {
    return this.database.find(id);
  }
}

// ✅ 遵循DIP的设计 - 依赖抽象
interface Database {
  save(data: any): void;
  find(id: string): any;
}

class MySQLDatabase implements Database {
  save(data: any): void {
    console.log('Saving to MySQL:', data);
  }
  
  find(id: string): any {
    console.log('Finding in MySQL:', id);
    return { id, name: 'User' };
  }
}

class MongoDatabase implements Database {
  save(data: any): void {
    console.log('Saving to MongoDB:', data);
  }
  
  find(id: string): any {
    console.log('Finding in MongoDB:', id);
    return { id, name: 'User' };
  }
}

class UserService {
  constructor(private database: Database) {} // 依赖抽象
  
  createUser(userData: any): void {
    this.database.save(userData);
  }
  
  getUser(id: string): any {
    return this.database.find(id);
  }
}

// 使用依赖注入
const mysqlDb = new MySQLDatabase();
const mongoDb = new MongoDatabase();

const userServiceWithMySQL = new UserService(mysqlDb);
const userServiceWithMongo = new UserService(mongoDb);
```

### 其他重要设计原则

```typescript
// DRY (Don't Repeat Yourself) - 不要重复自己
// ❌ 重复代码
function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

function validateUserEmail(user: User): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(user.email);
}

function validateContactEmail(contact: Contact): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(contact.email);
}

// ✅ 消除重复
class EmailValidator {
  private static readonly EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  
  static validate(email: string): boolean {
    return this.EMAIL_REGEX.test(email);
  }
}

function validateUserEmail(user: User): boolean {
  return EmailValidator.validate(user.email);
}

function validateContactEmail(contact: Contact): boolean {
  return EmailValidator.validate(contact.email);
}

// KISS (Keep It Simple, Stupid) - 保持简单
// ❌ 过度复杂的设计
class ComplexUserManager {
  private users: Map<string, User> = new Map();
  private userFactories: Map<string, UserFactory> = new Map();
  private userValidators: Map<string, UserValidator> = new Map();
  private userTransformers: Map<string, UserTransformer> = new Map();
  
  createUser(type: string, data: any): User {
    const factory = this.userFactories.get(type);
    const validator = this.userValidators.get(type);
    const transformer = this.userTransformers.get(type);
    
    if (!factory || !validator || !transformer) {
      throw new Error('Invalid user type');
    }
    
    const transformedData = transformer.transform(data);
    const validationResult = validator.validate(transformedData);
    
    if (!validationResult.isValid) {
      throw new Error('Invalid user data');
    }
    
    return factory.create(transformedData);
  }
}

// ✅ 简单直接的设计
class SimpleUserManager {
  private users: User[] = [];
  
  createUser(userData: UserData): User {
    if (!this.isValidUserData(userData)) {
      throw new Error('Invalid user data');
    }
    
    const user = new User(userData);
    this.users.push(user);
    return user;
  }
  
  private isValidUserData(data: UserData): boolean {
    return data.email && data.name && data.email.includes('@');
  }
}

// YAGNI (You Aren't Gonna Need It) - 你不会需要它
// ❌ 过度设计 - 添加可能永远不会用到的功能
class OverEngineeredComponent {
  // 为了"将来可能的需求"添加的复杂功能
  private plugins: Plugin[] = [];
  private eventBus: EventBus = new EventBus();
  private configManager: ConfigManager = new ConfigManager();
  private themeManager: ThemeManager = new ThemeManager();
  private i18nManager: I18nManager = new I18nManager();
  
  // 实际上只需要简单的渲染功能
  render() {
    return <div>Hello World</div>;
  }
}

// ✅ 只实现当前需要的功能
function SimpleComponent() {
  return <div>Hello World</div>;
}

// 当真正需要时再添加功能
function EnhancedComponent({ theme, message }: { theme?: string; message: string }) {
  return <div className={theme}>{message}</div>;
}
```

---

## 🏗️ 分层架构模式

### 经典三层架构

前端应用的分层架构通常包括表现层、业务层和数据层，每层都有明确的职责：

```typescript
// 数据层 (Data Layer) - 负责数据获取和持久化
interface Repository<T> {
  findById(id: string): Promise<T | null>;
  findAll(): Promise<T[]>;
  create(entity: T): Promise<T>;
  update(id: string, entity: Partial<T>): Promise<T>;
  delete(id: string): Promise<void>;
}

class ProductRepository implements Repository<Product> {
  private apiClient: ApiClient;

  constructor(apiClient: ApiClient) {
    this.apiClient = apiClient;
  }

  async findById(id: string): Promise<Product | null> {
    try {
      return await this.apiClient.get<Product>(`/api/products/${id}`);
    } catch (error) {
      if (error.status === 404) return null;
      throw error;
    }
  }

  async findAll(): Promise<Product[]> {
    const response = await this.apiClient.get<{ data: Product[] }>('/api/products');
    return response.data;
  }

  async create(product: Product): Promise<Product> {
    return await this.apiClient.post<Product>('/api/products', product);
  }

  async update(id: string, product: Partial<Product>): Promise<Product> {
    return await this.apiClient.put<Product>(`/api/products/${id}`, product);
  }

  async delete(id: string): Promise<void> {
    await this.apiClient.delete(`/api/products/${id}`);
  }
}

class UserRepository implements Repository<User> {
  private apiClient: ApiClient;
  private cacheManager: CacheManager;

  constructor(apiClient: ApiClient, cacheManager: CacheManager) {
    this.apiClient = apiClient;
    this.cacheManager = cacheManager;
  }

  async findById(id: string): Promise<User | null> {
    // 先检查缓存
    const cacheKey = `user:${id}`;
    const cached = await this.cacheManager.get<User>(cacheKey);
    if (cached) return cached;

    // 从API获取
    try {
      const user = await this.apiClient.get<User>(`/api/users/${id}`);
      // 缓存结果
      await this.cacheManager.set(cacheKey, user, { ttl: 5 * 60 * 1000 });
      return user;
    } catch (error) {
      if (error.status === 404) return null;
      throw error;
    }
  }

  async findAll(): Promise<User[]> {
    const response = await this.apiClient.get<{ data: User[] }>('/api/users');
    return response.data;
  }

  async create(user: User): Promise<User> {
    const newUser = await this.apiClient.post<User>('/api/users', user);
    // 清除相关缓存
    await this.cacheManager.invalidate('user:*');
    return newUser;
  }

  async update(id: string, user: Partial<User>): Promise<User> {
    const updatedUser = await this.apiClient.put<User>(`/api/users/${id}`, user);
    // 更新缓存
    await this.cacheManager.set(`user:${id}`, updatedUser);
    return updatedUser;
  }

  async delete(id: string): Promise<void> {
    await this.apiClient.delete(`/api/users/${id}`);
    // 清除缓存
    await this.cacheManager.delete(`user:${id}`);
  }
}
```

```typescript
// 业务层 (Business Layer) - 负责业务逻辑处理
interface ProductService {
  getProducts(filters?: ProductFilters): Promise<Product[]>;
  getProduct(id: string): Promise<Product>;
  createProduct(productData: CreateProductData): Promise<Product>;
  updateProduct(id: string, updates: Partial<Product>): Promise<Product>;
  deleteProduct(id: string): Promise<void>;
  searchProducts(query: string): Promise<Product[]>;
  getRecommendedProducts(userId: string): Promise<Product[]>;
}

class ProductServiceImpl implements ProductService {
  constructor(
    private productRepository: ProductRepository,
    private userRepository: UserRepository,
    private recommendationService: RecommendationService,
    private eventBus: EventBus
  ) {}

  async getProducts(filters?: ProductFilters): Promise<Product[]> {
    let products = await this.productRepository.findAll();

    // 应用过滤器
    if (filters) {
      products = this.applyFilters(products, filters);
    }

    // 应用业务规则
    products = this.applyBusinessRules(products);

    return products;
  }

  async getProduct(id: string): Promise<Product> {
    const product = await this.productRepository.findById(id);
    if (!product) {
      throw new Error(`Product with id ${id} not found`);
    }

    // 记录查看事件
    this.eventBus.emit('product:viewed', { productId: id, timestamp: new Date() });

    return product;
  }

  async createProduct(productData: CreateProductData): Promise<Product> {
    // 业务验证
    this.validateProductData(productData);

    // 生成SKU
    const sku = this.generateSKU(productData);

    const product: Product = {
      ...productData,
      id: generateId(),
      sku,
      status: 'active',
      createdAt: new Date(),
      updatedAt: new Date()
    };

    const createdProduct = await this.productRepository.create(product);

    // 发布事件
    this.eventBus.emit('product:created', createdProduct);

    return createdProduct;
  }

  async updateProduct(id: string, updates: Partial<Product>): Promise<Product> {
    const existingProduct = await this.getProduct(id);

    // 业务验证
    this.validateProductUpdates(existingProduct, updates);

    const updatedProduct = await this.productRepository.update(id, {
      ...updates,
      updatedAt: new Date()
    });

    // 发布事件
    this.eventBus.emit('product:updated', { before: existingProduct, after: updatedProduct });

    return updatedProduct;
  }

  async deleteProduct(id: string): Promise<void> {
    const product = await this.getProduct(id);

    // 业务规则检查
    if (product.status === 'active' && product.stock > 0) {
      throw new Error('Cannot delete active product with stock');
    }

    await this.productRepository.delete(id);

    // 发布事件
    this.eventBus.emit('product:deleted', product);
  }

  async searchProducts(query: string): Promise<Product[]> {
    // 搜索逻辑
    const products = await this.productRepository.findAll();
    return products.filter(product =>
      product.name.toLowerCase().includes(query.toLowerCase()) ||
      product.description.toLowerCase().includes(query.toLowerCase())
    );
  }

  async getRecommendedProducts(userId: string): Promise<Product[]> {
    const user = await this.userRepository.findById(userId);
    if (!user) {
      throw new Error('User not found');
    }

    // 使用推荐服务
    const recommendedIds = await this.recommendationService.getRecommendations(userId);

    const products = await Promise.all(
      recommendedIds.map(id => this.productRepository.findById(id))
    );

    return products.filter(Boolean) as Product[];
  }

  private applyFilters(products: Product[], filters: ProductFilters): Product[] {
    let filtered = products;

    if (filters.category) {
      filtered = filtered.filter(p => p.categoryId === filters.category);
    }

    if (filters.minPrice !== undefined) {
      filtered = filtered.filter(p => parseFloat(p.price) >= filters.minPrice!);
    }

    if (filters.maxPrice !== undefined) {
      filtered = filtered.filter(p => parseFloat(p.price) <= filters.maxPrice!);
    }

    if (filters.inStock) {
      filtered = filtered.filter(p => p.stock > 0);
    }

    return filtered;
  }

  private applyBusinessRules(products: Product[]): Product[] {
    return products
      .filter(p => p.status === 'active') // 只显示活跃商品
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime()); // 按创建时间排序
  }

  private validateProductData(data: CreateProductData): void {
    if (!data.name || data.name.trim().length === 0) {
      throw new Error('Product name is required');
    }

    if (!data.price || parseFloat(data.price) <= 0) {
      throw new Error('Product price must be greater than 0');
    }

    if (!data.categoryId) {
      throw new Error('Product category is required');
    }
  }

  private validateProductUpdates(existing: Product, updates: Partial<Product>): void {
    if (updates.price && parseFloat(updates.price) <= 0) {
      throw new Error('Product price must be greater than 0');
    }

    if (updates.stock && updates.stock < 0) {
      throw new Error('Product stock cannot be negative');
    }
  }

  private generateSKU(productData: CreateProductData): string {
    const categoryPrefix = productData.categoryId.toString().padStart(3, '0');
    const timestamp = Date.now().toString().slice(-6);
    const random = Math.random().toString(36).substr(2, 4).toUpperCase();
    return `${categoryPrefix}-${timestamp}-${random}`;
  }
}
```

```typescript
// 表现层 (Presentation Layer) - 负责UI渲染和用户交互
interface ProductListProps {
  filters?: ProductFilters;
  onProductSelect?: (product: Product) => void;
  onFiltersChange?: (filters: ProductFilters) => void;
}

// 容器组件 - 连接业务层
function ProductListContainer({ filters, onProductSelect, onFiltersChange }: ProductListProps) {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const productService = useService<ProductService>('ProductService');

  useEffect(() => {
    loadProducts();
  }, [filters]);

  const loadProducts = async () => {
    setLoading(true);
    setError(null);

    try {
      const products = await productService.getProducts(filters);
      setProducts(products);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load products');
    } finally {
      setLoading(false);
    }
  };

  const handleProductClick = (product: Product) => {
    onProductSelect?.(product);
  };

  const handleFilterChange = (newFilters: ProductFilters) => {
    onFiltersChange?.(newFilters);
  };

  if (loading) {
    return <ProductListSkeleton />;
  }

  if (error) {
    return <ErrorMessage message={error} onRetry={loadProducts} />;
  }

  return (
    <ProductListPresentation
      products={products}
      filters={filters}
      onProductClick={handleProductClick}
      onFiltersChange={handleFilterChange}
    />
  );
}

// 展示组件 - 纯UI渲染
interface ProductListPresentationProps {
  products: Product[];
  filters?: ProductFilters;
  onProductClick: (product: Product) => void;
  onFiltersChange: (filters: ProductFilters) => void;
}

function ProductListPresentation({
  products,
  filters,
  onProductClick,
  onFiltersChange
}: ProductListPresentationProps) {
  return (
    <div className="product-list">
      <ProductFilters
        filters={filters}
        onChange={onFiltersChange}
      />

      <div className="product-grid">
        {products.map(product => (
          <ProductCard
            key={product.id}
            product={product}
            onClick={() => onProductClick(product)}
          />
        ))}
      </div>

      {products.length === 0 && (
        <EmptyState message="No products found" />
      )}
    </div>
  );
}

// 原子组件 - 最小UI单元
interface ProductCardProps {
  product: Product;
  onClick: () => void;
}

function ProductCard({ product, onClick }: ProductCardProps) {
  return (
    <div className="product-card" onClick={onClick}>
      <img
        src={product.images[0]}
        alt={product.name}
        className="product-image"
        loading="lazy"
      />

      <div className="product-info">
        <h3 className="product-name">{product.name}</h3>
        <p className="product-price">¥{product.price}</p>

        {product.discountPrice && (
          <p className="product-discount">
            <span className="original-price">¥{product.price}</span>
            <span className="discount-price">¥{product.discountPrice}</span>
          </p>
        )}

        <div className="product-meta">
          <span className="product-stock">
            库存: {product.stock}
          </span>

          {product.stock === 0 && (
            <span className="out-of-stock">缺货</span>
          )}
        </div>
      </div>
    </div>
  );
}
```

### 依赖注入容器

```typescript
// 依赖注入容器实现
type Constructor<T = {}> = new (...args: any[]) => T;
type ServiceFactory<T> = () => T;
type ServiceIdentifier<T> = string | symbol | Constructor<T>;

interface ServiceDescriptor<T> {
  identifier: ServiceIdentifier<T>;
  factory: ServiceFactory<T>;
  singleton: boolean;
  dependencies: ServiceIdentifier<any>[];
}

class DIContainer {
  private services = new Map<ServiceIdentifier<any>, ServiceDescriptor<any>>();
  private instances = new Map<ServiceIdentifier<any>, any>();

  // 注册服务
  register<T>(
    identifier: ServiceIdentifier<T>,
    factory: ServiceFactory<T>,
    options: { singleton?: boolean; dependencies?: ServiceIdentifier<any>[] } = {}
  ): void {
    const { singleton = true, dependencies = [] } = options;

    this.services.set(identifier, {
      identifier,
      factory,
      singleton,
      dependencies
    });
  }

  // 注册类
  registerClass<T>(
    identifier: ServiceIdentifier<T>,
    constructor: Constructor<T>,
    options: { singleton?: boolean; dependencies?: ServiceIdentifier<any>[] } = {}
  ): void {
    const { dependencies = [] } = options;

    this.register(
      identifier,
      () => {
        const deps = dependencies.map(dep => this.resolve(dep));
        return new constructor(...deps);
      },
      options
    );
  }

  // 解析服务
  resolve<T>(identifier: ServiceIdentifier<T>): T {
    const descriptor = this.services.get(identifier);
    if (!descriptor) {
      throw new Error(`Service ${String(identifier)} not registered`);
    }

    // 单例模式检查
    if (descriptor.singleton && this.instances.has(identifier)) {
      return this.instances.get(identifier);
    }

    // 解析依赖
    const dependencies = descriptor.dependencies.map(dep => this.resolve(dep));

    // 创建实例
    const instance = descriptor.factory();

    // 缓存单例
    if (descriptor.singleton) {
      this.instances.set(identifier, instance);
    }

    return instance;
  }

  // 检查服务是否已注册
  has<T>(identifier: ServiceIdentifier<T>): boolean {
    return this.services.has(identifier);
  }

  // 清除所有服务
  clear(): void {
    this.services.clear();
    this.instances.clear();
  }
}

// 全局容器实例
export const container = new DIContainer();

// 服务注册
container.register('ApiClient', () => new ApiClient(process.env.NEXT_PUBLIC_API_BASE_URL!));
container.register('CacheManager', () => new CacheManager());
container.register('EventBus', () => new EventBus());

container.registerClass(
  'ProductRepository',
  ProductRepository,
  { dependencies: ['ApiClient'] }
);

container.registerClass(
  'UserRepository',
  UserRepository,
  { dependencies: ['ApiClient', 'CacheManager'] }
);

container.registerClass(
  'ProductService',
  ProductServiceImpl,
  { dependencies: ['ProductRepository', 'UserRepository', 'RecommendationService', 'EventBus'] }
);

// React Hook for dependency injection
export function useService<T>(identifier: ServiceIdentifier<T>): T {
  return useMemo(() => container.resolve(identifier), [identifier]);
}

// 装饰器支持 (如果使用装饰器)
export function Injectable<T extends Constructor>(constructor: T) {
  return class extends constructor {
    static [Symbol.hasInstance](instance: any) {
      return instance instanceof constructor;
    }
  };
}

export function Inject(identifier: ServiceIdentifier<any>) {
  return function (target: any, propertyKey: string | symbol | undefined, parameterIndex: number) {
    // 装饰器元数据处理
    const existingTokens = Reflect.getMetadata('design:paramtypes', target) || [];
    existingTokens[parameterIndex] = identifier;
    Reflect.defineMetadata('design:paramtypes', existingTokens, target);
  };
}
```

---

## 🧩 模块化与组件化设计

### 特性模块设计

```typescript
// 特性模块结构
// features/product/
//   ├── components/
//   │   ├── ProductCard.tsx
//   │   ├── ProductList.tsx
//   │   └── ProductFilters.tsx
//   ├── services/
//   │   ├── ProductService.ts
//   │   └── ProductRepository.ts
//   ├── types/
//   │   └── Product.ts
//   ├── hooks/
//   │   ├── useProducts.ts
//   │   └── useProductFilters.ts
//   ├── utils/
//   │   └── productUtils.ts
//   └── index.ts

// features/product/types/Product.ts
export interface Product {
  id: string;
  name: string;
  description: string;
  price: string;
  discountPrice?: string;
  images: string[];
  categoryId: string;
  stock: number;
  status: 'active' | 'inactive' | 'draft';
  createdAt: Date;
  updatedAt: Date;
}

export interface ProductFilters {
  category?: string;
  minPrice?: number;
  maxPrice?: number;
  inStock?: boolean;
  search?: string;
}

export interface CreateProductData {
  name: string;
  description: string;
  price: string;
  categoryId: string;
  images: string[];
  stock: number;
}

// features/product/hooks/useProducts.ts
export function useProducts(filters?: ProductFilters) {
  const productService = useService<ProductService>('ProductService');

  return useQuery({
    queryKey: ['products', filters],
    queryFn: () => productService.getProducts(filters),
    staleTime: 5 * 60 * 1000,
  });
}

export function useProduct(id: string) {
  const productService = useService<ProductService>('ProductService');

  return useQuery({
    queryKey: ['product', id],
    queryFn: () => productService.getProduct(id),
    enabled: !!id,
    staleTime: 10 * 60 * 1000,
  });
}

export function useProductMutations() {
  const productService = useService<ProductService>('ProductService');
  const queryClient = useQueryClient();

  const createProduct = useMutation({
    mutationFn: (data: CreateProductData) => productService.createProduct(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });

  const updateProduct = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Partial<Product> }) =>
      productService.updateProduct(id, updates),
    onSuccess: (updatedProduct) => {
      queryClient.setQueryData(['product', updatedProduct.id], updatedProduct);
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });

  const deleteProduct = useMutation({
    mutationFn: (id: string) => productService.deleteProduct(id),
    onSuccess: (_, deletedId) => {
      queryClient.removeQueries({ queryKey: ['product', deletedId] });
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });

  return {
    createProduct,
    updateProduct,
    deleteProduct,
  };
}

// features/product/utils/productUtils.ts
export class ProductUtils {
  static formatPrice(price: string): string {
    const num = parseFloat(price);
    return new Intl.NumberFormat('zh-CN', {
      style: 'currency',
      currency: 'CNY'
    }).format(num);
  }

  static calculateDiscount(originalPrice: string, discountPrice: string): number {
    const original = parseFloat(originalPrice);
    const discount = parseFloat(discountPrice);
    return Math.round(((original - discount) / original) * 100);
  }

  static isInStock(product: Product): boolean {
    return product.stock > 0 && product.status === 'active';
  }

  static getMainImage(product: Product): string {
    return product.images[0] || '/images/placeholder.jpg';
  }

  static generateSEOTitle(product: Product): string {
    return `${product.name} - 优质商品 - Mall商城`;
  }

  static generateSEODescription(product: Product): string {
    return `${product.description.substring(0, 150)}...`;
  }
}

// features/product/index.ts - 模块导出
export * from './types/Product';
export * from './components/ProductCard';
export * from './components/ProductList';
export * from './components/ProductFilters';
export * from './hooks/useProducts';
export * from './utils/productUtils';

// 默认导出模块配置
export default {
  name: 'Product',
  version: '1.0.0',
  dependencies: ['User', 'Category'],
  routes: [
    { path: '/products', component: 'ProductList' },
    { path: '/products/:id', component: 'ProductDetail' },
  ],
  services: ['ProductService', 'ProductRepository'],
};
```

### 原子设计模式

```typescript
// 原子设计层次结构
// atoms/ - 原子组件
//   ├── Button.tsx
//   ├── Input.tsx
//   ├── Label.tsx
//   └── Image.tsx
// molecules/ - 分子组件
//   ├── SearchBox.tsx
//   ├── PriceDisplay.tsx
//   └── ImageGallery.tsx
// organisms/ - 有机体组件
//   ├── ProductCard.tsx
//   ├── ProductGrid.tsx
//   └── Header.tsx
// templates/ - 模板
//   ├── ProductListTemplate.tsx
//   └── ProductDetailTemplate.tsx
// pages/ - 页面
//   ├── ProductListPage.tsx
//   └── ProductDetailPage.tsx

// atoms/Button.tsx - 原子组件
interface ButtonProps {
  variant: 'primary' | 'secondary' | 'danger' | 'ghost';
  size: 'small' | 'medium' | 'large';
  disabled?: boolean;
  loading?: boolean;
  children: React.ReactNode;
  onClick?: () => void;
}

export function Button({
  variant = 'primary',
  size = 'medium',
  disabled = false,
  loading = false,
  children,
  onClick
}: ButtonProps) {
  const baseClasses = 'inline-flex items-center justify-center font-medium rounded-md transition-colors';

  const variantClasses = {
    primary: 'bg-blue-600 text-white hover:bg-blue-700',
    secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300',
    danger: 'bg-red-600 text-white hover:bg-red-700',
    ghost: 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
  };

  const sizeClasses = {
    small: 'px-3 py-1.5 text-sm',
    medium: 'px-4 py-2 text-base',
    large: 'px-6 py-3 text-lg'
  };

  const className = `
    ${baseClasses}
    ${variantClasses[variant]}
    ${sizeClasses[size]}
    ${disabled || loading ? 'opacity-50 cursor-not-allowed' : ''}
  `.trim();

  return (
    <button
      className={className}
      disabled={disabled || loading}
      onClick={onClick}
    >
      {loading && <Spinner className="mr-2" />}
      {children}
    </button>
  );
}

// molecules/PriceDisplay.tsx - 分子组件
interface PriceDisplayProps {
  price: string;
  discountPrice?: string;
  currency?: string;
  size?: 'small' | 'medium' | 'large';
}

export function PriceDisplay({
  price,
  discountPrice,
  currency = 'CNY',
  size = 'medium'
}: PriceDisplayProps) {
  const hasDiscount = discountPrice && discountPrice !== price;
  const discount = hasDiscount ? ProductUtils.calculateDiscount(price, discountPrice) : 0;

  const sizeClasses = {
    small: 'text-sm',
    medium: 'text-base',
    large: 'text-lg'
  };

  return (
    <div className={`price-display ${sizeClasses[size]}`}>
      {hasDiscount ? (
        <div className="flex items-center space-x-2">
          <span className="text-red-600 font-bold">
            {ProductUtils.formatPrice(discountPrice)}
          </span>
          <span className="text-gray-500 line-through">
            {ProductUtils.formatPrice(price)}
          </span>
          <span className="bg-red-100 text-red-800 px-2 py-1 rounded text-xs">
            -{discount}%
          </span>
        </div>
      ) : (
        <span className="text-gray-900 font-bold">
          {ProductUtils.formatPrice(price)}
        </span>
      )}
    </div>
  );
}

// organisms/ProductCard.tsx - 有机体组件
interface ProductCardProps {
  product: Product;
  onAddToCart?: (product: Product) => void;
  onQuickView?: (product: Product) => void;
  onClick?: (product: Product) => void;
}

export function ProductCard({
  product,
  onAddToCart,
  onQuickView,
  onClick
}: ProductCardProps) {
  const [imageLoaded, setImageLoaded] = useState(false);
  const [imageError, setImageError] = useState(false);

  const handleAddToCart = (e: React.MouseEvent) => {
    e.stopPropagation();
    onAddToCart?.(product);
  };

  const handleQuickView = (e: React.MouseEvent) => {
    e.stopPropagation();
    onQuickView?.(product);
  };

  const handleClick = () => {
    onClick?.(product);
  };

  return (
    <div
      className="product-card bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow cursor-pointer"
      onClick={handleClick}
    >
      {/* 图片区域 */}
      <div className="relative aspect-square overflow-hidden rounded-t-lg">
        {!imageLoaded && !imageError && (
          <div className="absolute inset-0 bg-gray-200 animate-pulse" />
        )}

        <img
          src={imageError ? '/images/placeholder.jpg' : ProductUtils.getMainImage(product)}
          alt={product.name}
          className={`w-full h-full object-cover transition-opacity ${
            imageLoaded ? 'opacity-100' : 'opacity-0'
          }`}
          onLoad={() => setImageLoaded(true)}
          onError={() => setImageError(true)}
          loading="lazy"
        />

        {/* 悬浮操作按钮 */}
        <div className="absolute inset-0 bg-black bg-opacity-0 hover:bg-opacity-20 transition-all">
          <div className="absolute top-2 right-2 opacity-0 hover:opacity-100 transition-opacity">
            <Button
              variant="ghost"
              size="small"
              onClick={handleQuickView}
            >
              快速查看
            </Button>
          </div>
        </div>

        {/* 库存状态 */}
        {!ProductUtils.isInStock(product) && (
          <div className="absolute inset-0 bg-gray-900 bg-opacity-50 flex items-center justify-center">
            <span className="text-white font-bold">缺货</span>
          </div>
        )}
      </div>

      {/* 内容区域 */}
      <div className="p-4">
        <h3 className="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
          {product.name}
        </h3>

        <p className="text-gray-600 text-sm mb-3 line-clamp-2">
          {product.description}
        </p>

        <PriceDisplay
          price={product.price}
          discountPrice={product.discountPrice}
          size="medium"
        />

        <div className="mt-4 flex items-center justify-between">
          <span className="text-sm text-gray-500">
            库存: {product.stock}
          </span>

          <Button
            variant="primary"
            size="small"
            disabled={!ProductUtils.isInStock(product)}
            onClick={handleAddToCart}
          >
            加入购物车
          </Button>
        </div>
      </div>
    </div>
  );
}
```

---

## 🎯 面试常考知识点

### 1. 前端架构设计原则

**Q: 如何设计一个可扩展的前端架构？**

**A: 可扩展前端架构的核心要素：**

```typescript
// 可扩展架构设计要点
const scalableArchitecturePrinciples = {
  // 1. 模块化设计
  modularity: {
    principle: '高内聚，低耦合',
    implementation: [
      '按功能划分模块',
      '明确模块边界',
      '定义清晰的接口',
      '避免循环依赖'
    ],
    example: `
      // 模块结构
      features/
        user/
          components/
          services/
          types/
          index.ts
        product/
          components/
          services/
          types/
          index.ts
    `
  },

  // 2. 分层架构
  layeredArchitecture: {
    layers: ['Presentation', 'Business', 'Data', 'Infrastructure'],
    benefits: [
      '职责分离',
      '易于测试',
      '便于维护',
      '支持替换'
    ],
    implementation: `
      // 分层示例
      presentation/    // UI组件
      business/        // 业务逻辑
      data/           // 数据访问
      infrastructure/ // 基础设施
    `
  },

  // 3. 依赖注入
  dependencyInjection: {
    benefits: [
      '降低耦合度',
      '提高可测试性',
      '支持配置化',
      '便于扩展'
    ],
    patterns: ['Constructor Injection', 'Property Injection', 'Method Injection']
  },

  // 4. 事件驱动架构
  eventDriven: {
    components: ['Event Bus', 'Event Handlers', 'Event Publishers'],
    benefits: [
      '松耦合',
      '异步处理',
      '易于扩展',
      '支持插件化'
    ]
  }
};

// 架构质量评估指标
const architectureQualityMetrics = {
  maintainability: {
    metrics: ['代码复杂度', '重复代码率', '模块耦合度'],
    tools: ['ESLint', 'SonarQube', 'CodeClimate']
  },

  scalability: {
    metrics: ['模块数量', '依赖关系', '构建时间'],
    strategies: ['微前端', '代码分割', '懒加载']
  },

  testability: {
    metrics: ['测试覆盖率', '单元测试数量', '集成测试数量'],
    practices: ['TDD', 'BDD', 'Mock/Stub']
  },

  performance: {
    metrics: ['首屏时间', '交互时间', '包体积'],
    optimizations: ['Tree Shaking', 'Code Splitting', 'Caching']
  }
};
```

### 2. 组件设计模式

**Q: 常见的React组件设计模式有哪些？**

**A: React组件设计模式详解：**

```typescript
// 1. 容器/展示组件模式
// Container Component - 负责数据和逻辑
function ProductListContainer() {
  const { data: products, loading, error } = useProducts();
  const [filters, setFilters] = useState<ProductFilters>({});

  const handleFilterChange = (newFilters: ProductFilters) => {
    setFilters(newFilters);
  };

  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;

  return (
    <ProductListPresentation
      products={products}
      filters={filters}
      onFilterChange={handleFilterChange}
    />
  );
}

// Presentation Component - 负责UI渲染
interface ProductListPresentationProps {
  products: Product[];
  filters: ProductFilters;
  onFilterChange: (filters: ProductFilters) => void;
}

function ProductListPresentation({
  products,
  filters,
  onFilterChange
}: ProductListPresentationProps) {
  return (
    <div>
      <ProductFilters filters={filters} onChange={onFilterChange} />
      <ProductGrid products={products} />
    </div>
  );
}

// 2. 高阶组件模式 (HOC)
function withLoading<P extends object>(
  Component: React.ComponentType<P>
) {
  return function WithLoadingComponent(props: P & { loading: boolean }) {
    const { loading, ...restProps } = props;

    if (loading) {
      return <LoadingSpinner />;
    }

    return <Component {...(restProps as P)} />;
  };
}

// 使用HOC
const ProductListWithLoading = withLoading(ProductList);

// 3. Render Props模式
interface DataFetcherProps<T> {
  url: string;
  children: (data: {
    data: T | null;
    loading: boolean;
    error: Error | null;
  }) => React.ReactNode;
}

function DataFetcher<T>({ url, children }: DataFetcherProps<T>) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    fetch(url)
      .then(res => res.json())
      .then(setData)
      .catch(setError)
      .finally(() => setLoading(false));
  }, [url]);

  return <>{children({ data, loading, error })}</>;
}

// 使用Render Props
function ProductPage() {
  return (
    <DataFetcher<Product[]> url="/api/products">
      {({ data: products, loading, error }) => {
        if (loading) return <LoadingSpinner />;
        if (error) return <ErrorMessage error={error} />;
        return <ProductList products={products} />;
      }}
    </DataFetcher>
  );
}

// 4. 复合组件模式
interface TabsContextType {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

const TabsContext = createContext<TabsContextType | null>(null);

function Tabs({ children, defaultTab }: { children: React.ReactNode; defaultTab: string }) {
  const [activeTab, setActiveTab] = useState(defaultTab);

  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <div className="tabs">{children}</div>
    </TabsContext.Provider>
  );
}

function TabList({ children }: { children: React.ReactNode }) {
  return <div className="tab-list">{children}</div>;
}

function Tab({ value, children }: { value: string; children: React.ReactNode }) {
  const context = useContext(TabsContext);
  if (!context) throw new Error('Tab must be used within Tabs');

  const { activeTab, setActiveTab } = context;

  return (
    <button
      className={`tab ${activeTab === value ? 'active' : ''}`}
      onClick={() => setActiveTab(value)}
    >
      {children}
    </button>
  );
}

function TabPanels({ children }: { children: React.ReactNode }) {
  return <div className="tab-panels">{children}</div>;
}

function TabPanel({ value, children }: { value: string; children: React.ReactNode }) {
  const context = useContext(TabsContext);
  if (!context) throw new Error('TabPanel must be used within Tabs');

  const { activeTab } = context;

  if (activeTab !== value) return null;

  return <div className="tab-panel">{children}</div>;
}

// 组合使用
Tabs.List = TabList;
Tabs.Tab = Tab;
Tabs.Panels = TabPanels;
Tabs.Panel = TabPanel;

// 使用复合组件
function ProductDetailPage() {
  return (
    <Tabs defaultTab="description">
      <Tabs.List>
        <Tabs.Tab value="description">商品描述</Tabs.Tab>
        <Tabs.Tab value="reviews">用户评价</Tabs.Tab>
        <Tabs.Tab value="shipping">配送信息</Tabs.Tab>
      </Tabs.List>

      <Tabs.Panels>
        <Tabs.Panel value="description">
          <ProductDescription />
        </Tabs.Panel>
        <Tabs.Panel value="reviews">
          <ProductReviews />
        </Tabs.Panel>
        <Tabs.Panel value="shipping">
          <ShippingInfo />
        </Tabs.Panel>
      </Tabs.Panels>
    </Tabs>
  );
}

// 5. 自定义Hook模式
function useToggle(initialValue: boolean = false) {
  const [value, setValue] = useState(initialValue);

  const toggle = useCallback(() => setValue(v => !v), []);
  const setTrue = useCallback(() => setValue(true), []);
  const setFalse = useCallback(() => setValue(false), []);

  return { value, toggle, setTrue, setFalse };
}

function useLocalStorage<T>(key: string, initialValue: T) {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      return initialValue;
    }
  });

  const setValue = (value: T | ((val: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);
      window.localStorage.setItem(key, JSON.stringify(valueToStore));
    } catch (error) {
      console.error('Error saving to localStorage:', error);
    }
  };

  return [storedValue, setValue] as const;
}

// 使用自定义Hook
function ProductCard({ product }: { product: Product }) {
  const { value: isExpanded, toggle } = useToggle();
  const [favorites, setFavorites] = useLocalStorage<string[]>('favorites', []);

  const isFavorite = favorites.includes(product.id);

  const toggleFavorite = () => {
    setFavorites(prev =>
      isFavorite
        ? prev.filter(id => id !== product.id)
        : [...prev, product.id]
    );
  };

  return (
    <div className="product-card">
      <h3 onClick={toggle}>{product.name}</h3>
      {isExpanded && <p>{product.description}</p>}
      <button onClick={toggleFavorite}>
        {isFavorite ? '❤️' : '🤍'}
      </button>
    </div>
  );
}
```

### 3. 状态管理架构

**Q: 如何选择合适的状态管理方案？**

**A: 状态管理方案选择指南：**

```typescript
// 状态管理方案对比
const stateManagementComparison = {
  // 1. 本地状态 (useState, useReducer)
  localState: {
    useCase: ['组件内部状态', '简单表单', '临时UI状态'],
    pros: ['简单直接', '性能好', '无额外依赖'],
    cons: ['难以共享', '状态提升复杂', '缺乏持久化'],
    example: `
      function Counter() {
        const [count, setCount] = useState(0);
        return (
          <div>
            <span>{count}</span>
            <button onClick={() => setCount(c => c + 1)}>+</button>
          </div>
        );
      }
    `
  },

  // 2. Context API
  contextAPI: {
    useCase: ['主题切换', '用户认证', '语言设置', '中等复杂度应用'],
    pros: ['React内置', '避免prop drilling', '类型安全'],
    cons: ['性能问题', '重新渲染', '复杂状态管理困难'],
    example: `
      const ThemeContext = createContext();

      function ThemeProvider({ children }) {
        const [theme, setTheme] = useState('light');
        return (
          <ThemeContext.Provider value={{ theme, setTheme }}>
            {children}
          </ThemeContext.Provider>
        );
      }
    `
  },

  // 3. Redux Toolkit
  reduxToolkit: {
    useCase: ['大型应用', '复杂状态逻辑', '时间旅行调试', '状态持久化'],
    pros: ['可预测', '强大的开发工具', '中间件支持', '社区成熟'],
    cons: ['学习曲线', '样板代码', '过度工程'],
    example: `
      const counterSlice = createSlice({
        name: 'counter',
        initialState: { value: 0 },
        reducers: {
          increment: (state) => {
            state.value += 1;
          }
        }
      });
    `
  },

  // 4. Zustand
  zustand: {
    useCase: ['中型应用', '简单全局状态', '快速原型'],
    pros: ['轻量级', '简单API', 'TypeScript友好', '无样板代码'],
    cons: ['生态较小', '调试工具有限'],
    example: `
      const useStore = create((set) => ({
        count: 0,
        increment: () => set((state) => ({ count: state.count + 1 }))
      }));
    `
  },

  // 5. React Query + Zustand
  hybrid: {
    useCase: ['现代应用', '服务端状态 + 客户端状态分离'],
    pros: ['职责分离', '最佳实践', '性能优化'],
    cons: ['学习成本', '多个依赖'],
    example: `
      // 服务端状态 - React Query
      const { data: products } = useQuery(['products'], fetchProducts);

      // 客户端状态 - Zustand
      const { cart, addToCart } = useCartStore();
    `
  }
};

// 状态管理决策树
const stateManagementDecisionTree = {
  questions: [
    {
      question: '状态是否需要在多个组件间共享？',
      no: '使用 useState/useReducer',
      yes: '继续下一个问题'
    },
    {
      question: '应用规模是否较大（>50个组件）？',
      no: '考虑 Context API 或 Zustand',
      yes: '继续下一个问题'
    },
    {
      question: '是否需要复杂的状态逻辑和调试？',
      yes: '使用 Redux Toolkit',
      no: '继续下一个问题'
    },
    {
      question: '是否有大量服务端状态？',
      yes: '使用 React Query + Zustand',
      no: '使用 Zustand'
    }
  ]
};
```

---

## 📚 实战练习

### 练习1：设计模块化架构

**任务**: 为Mall-Frontend设计一个模块化的架构，包括用户模块、商品模块、订单模块等。

**要求**:
- 使用特性模块设计模式
- 实现模块间的依赖管理
- 设计清晰的模块接口
- 支持模块的独立开发和测试

### 练习2：实现依赖注入容器

**任务**: 实现一个轻量级的依赖注入容器，支持服务注册、解析和生命周期管理。

**要求**:
- 支持单例和瞬态生命周期
- 实现循环依赖检测
- 提供React Hook集成
- 支持装饰器语法

### 练习3：构建组件库

**任务**: 基于原子设计模式构建一个可复用的组件库。

**要求**:
- 实现原子、分子、有机体组件
- 提供完整的TypeScript类型定义
- 支持主题定制
- 包含Storybook文档

---

## 📚 本章总结

通过本章学习，我们全面掌握了前端架构设计的核心原理：

### 🎯 核心收获

1. **设计原则精通** 🎯
   - 掌握了SOLID原则在前端的应用
   - 学会了DRY、KISS、YAGNI等设计原则
   - 理解了架构设计的权衡和取舍

2. **分层架构设计** 🏗️
   - 掌握了表现层、业务层、数据层的职责划分
   - 学会了依赖注入和控制反转的实现
   - 理解了分层架构的优势和挑战

3. **模块化设计** 🧩
   - 掌握了特性模块和原子设计模式
   - 学会了模块间的依赖管理
   - 理解了组件化的设计思想

4. **架构模式应用** 🔄
   - 掌握了多种React组件设计模式
   - 学会了状态管理方案的选择
   - 理解了架构演进的策略

5. **企业级实践** 🚀
   - 学会了大型项目的架构治理
   - 掌握了架构质量的评估方法
   - 理解了团队协作的架构规范

### 🚀 技术进阶

- **下一步学习**: 状态管理架构设计
- **实践建议**: 在项目中应用分层架构模式
- **深入方向**: 微前端架构和插件化设计

优秀的架构是应用成功的基石，也是团队高效协作的保障！ 🎉

---

*下一章我们将学习《状态管理架构设计》，探索复杂应用的状态管理策略！* 🚀
```
```
```
