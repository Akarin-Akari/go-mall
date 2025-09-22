# 📘 第2章：泛型与高级类型应用

> 掌握TypeScript的泛型编程，构建灵活且类型安全的代码

## 🎯 学习目标

通过本章学习，你将掌握：

- 泛型的基本概念和语法
- 泛型约束和条件类型
- 映射类型和模板字面量类型
- 实用工具类型的使用
- 高级类型编程技巧
- Mall-Frontend项目中的泛型应用

## 📖 目录

- [泛型基础](#泛型基础)
- [泛型约束](#泛型约束)
- [条件类型](#条件类型)
- [映射类型](#映射类型)
- [模板字面量类型](#模板字面量类型)
- [实用工具类型](#实用工具类型)
- [高级类型编程](#高级类型编程)
- [Mall-Frontend实战案例](#mall-frontend实战案例)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🧬 泛型基础

### 什么是泛型？

泛型允许我们在定义函数、类或接口时使用类型参数，使代码更加灵活和可重用。

```typescript
// 不使用泛型 - 代码重复
function identityString(arg: string): string {
  return arg;
}

function identityNumber(arg: number): number {
  return arg;
}

// 使用泛型 - 一个函数处理多种类型
function identity<T>(arg: T): T {
  return arg;
}

// 使用时指定类型
const stringResult = identity<string>('hello');
const numberResult = identity<number>(42);

// 类型推断
const autoString = identity('hello'); // TypeScript自动推断为string
const autoNumber = identity(42); // TypeScript自动推断为number
```

### 🔄 语言对比：泛型实现方式

```java
// Java - 泛型（类型擦除）
public class GenericExample {
    // 泛型方法
    public static <T> T identity(T arg) {
        return arg;
    }

    // 泛型类
    public static class Box<T> {
        private T value;

        public Box(T value) {
            this.value = value;
        }

        public T getValue() {
            return value;
        }
    }

    public static void main(String[] args) {
        // 使用时指定类型
        String stringResult = identity("hello");
        Integer numberResult = identity(42);

        // 泛型类使用
        Box<String> stringBox = new Box<>("hello");
        Box<Integer> intBox = new Box<>(42);
    }
}
```

```python
# Python - 泛型类型提示（3.5+）
from typing import TypeVar, Generic, List

T = TypeVar('T')

# 泛型函数
def identity(arg: T) -> T:
    return arg

# 泛型类
class Box(Generic[T]):
    def __init__(self, value: T) -> None:
        self._value = value

    def get_value(self) -> T:
        return self._value

# 使用示例
string_result = identity("hello")  # 类型推断为str
number_result = identity(42)       # 类型推断为int

string_box: Box[str] = Box("hello")
int_box: Box[int] = Box(42)

# Python运行时不强制类型检查
mixed_box: Box[str] = Box(42)  # mypy会警告，但运行时正常
```

```csharp
// C# - 泛型（运行时保留类型信息）
public class GenericExample
{
    // 泛型方法
    public static T Identity<T>(T arg)
    {
        return arg;
    }

    // 泛型类
    public class Box<T>
    {
        private T _value;

        public Box(T value)
        {
            _value = value;
        }

        public T Value => _value;
    }

    static void Main()
    {
        // 类型推断
        var stringResult = Identity("hello");
        var numberResult = Identity(42);

        // 显式类型指定
        var stringBox = new Box<string>("hello");
        var intBox = new Box<int>(42);
    }
}
```

```go
// Go - 泛型（Go 1.18+）
package main

import "fmt"

// 泛型函数
func Identity[T any](arg T) T {
    return arg
}

// 泛型结构体
type Box[T any] struct {
    value T
}

func NewBox[T any](value T) *Box[T] {
    return &Box[T]{value: value}
}

func (b *Box[T]) GetValue() T {
    return b.value
}

func main() {
    // 类型推断
    stringResult := Identity("hello")
    numberResult := Identity(42)

    // 显式类型指定
    stringBox := NewBox[string]("hello")
    intBox := NewBox[int](42)

    fmt.Println(stringResult, numberResult)
    fmt.Println(stringBox.GetValue(), intBox.GetValue())
}
```

**💡 泛型特性对比：**

| 特性          | TypeScript   | Java            | Python     | C#         | Go         |
| ------------- | ------------ | --------------- | ---------- | ---------- | ---------- |
| **类型擦除**  | 编译时擦除   | 运行时擦除      | 运行时保留 | 运行时保留 | 运行时保留 |
| **类型推断**  | 强大         | 有限            | 基础       | 强大       | 基础       |
| **约束语法**  | `extends`    | `extends/super` | `bound`    | `where`    | 类型约束   |
| **协变/逆变** | 支持         | 支持            | 支持       | 支持       | 不支持     |
| **性能影响**  | 无（编译时） | 最小            | 无         | 最小       | 无         |

### 泛型函数

```typescript
// 基础泛型函数
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

const person = { name: '张三', age: 30, email: 'zhangsan@example.com' };
const name = getProperty(person, 'name'); // 类型为string
const age = getProperty(person, 'age'); // 类型为number

// 多个类型参数
function merge<T, U>(obj1: T, obj2: U): T & U {
  return { ...obj1, ...obj2 };
}

const merged = merge({ name: '李四' }, { age: 25 });
// merged的类型为 { name: string } & { age: number }
```

### 泛型接口

```typescript
// 泛型接口定义
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  timestamp: Date;
}

// 使用泛型接口
interface User {
  id: number;
  username: string;
  email: string;
}

interface Product {
  id: number;
  name: string;
  price: number;
}

// 具体类型的API响应
type UserResponse = ApiResponse<User>;
type ProductListResponse = ApiResponse<Product[]>;
type LoginResponse = ApiResponse<{ token: string; user: User }>;

// 泛型接口的实现
class ApiClient {
  async get<T>(url: string): Promise<ApiResponse<T>> {
    const response = await fetch(url);
    return response.json();
  }

  async post<T, U>(url: string, data: T): Promise<ApiResponse<U>> {
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return response.json();
  }
}
```

### 泛型类

```typescript
// 泛型类定义
class GenericRepository<T> {
  private items: T[] = [];

  add(item: T): void {
    this.items.push(item);
  }

  findById<K extends keyof T>(key: K, value: T[K]): T | undefined {
    return this.items.find(item => item[key] === value);
  }

  getAll(): T[] {
    return [...this.items];
  }

  update(predicate: (item: T) => boolean, updates: Partial<T>): boolean {
    const index = this.items.findIndex(predicate);
    if (index !== -1) {
      this.items[index] = { ...this.items[index], ...updates };
      return true;
    }
    return false;
  }

  delete(predicate: (item: T) => boolean): boolean {
    const index = this.items.findIndex(predicate);
    if (index !== -1) {
      this.items.splice(index, 1);
      return true;
    }
    return false;
  }
}

// 使用泛型类
const userRepository = new GenericRepository<User>();
const productRepository = new GenericRepository<Product>();

userRepository.add({ id: 1, username: 'john', email: 'john@example.com' });
const user = userRepository.findById('id', 1);
```

---

## 🔒 泛型约束

### 基础约束

泛型约束限制类型参数必须满足某些条件：

```typescript
// 约束类型参数必须有length属性
interface Lengthwise {
  length: number;
}

function loggingIdentity<T extends Lengthwise>(arg: T): T {
  console.log(arg.length); // 现在可以访问length属性
  return arg;
}

// 正确使用
loggingIdentity('hello'); // string有length属性
loggingIdentity([1, 2, 3]); // array有length属性
loggingIdentity({ length: 10, value: 3 }); // 对象有length属性

// 错误使用
// loggingIdentity(3); // number没有length属性，编译错误
```

### 🔄 语言对比：泛型约束机制

```java
// Java - 有界类型参数
interface Lengthwise {
    int getLength();
}

public class GenericConstraints {
    // 上界约束 - extends
    public static <T extends Lengthwise> T loggingIdentity(T arg) {
        System.out.println(arg.getLength());
        return arg;
    }

    // 多重约束
    public static <T extends Comparable<T> & Serializable> void process(T item) {
        // T必须同时实现Comparable和Serializable
    }

    // 通配符约束
    public static void printList(List<? extends Number> list) {
        for (Number n : list) {
            System.out.println(n);
        }
    }

    // 下界约束 - super
    public static void addNumbers(List<? super Integer> list) {
        list.add(42);
    }
}
```

```python
# Python - 类型约束（Protocol和TypeVar）
from typing import TypeVar, Protocol, List
from abc import abstractmethod

# 协议约束
class Lengthwise(Protocol):
    @property
    def length(self) -> int: ...

T = TypeVar('T', bound=Lengthwise)

def logging_identity(arg: T) -> T:
    print(arg.length)  # 可以访问length属性
    return arg

# 使用示例
class MyString:
    def __init__(self, value: str):
        self.value = value

    @property
    def length(self) -> int:
        return len(self.value)

# 正确使用
result = logging_identity(MyString("hello"))

# 多重约束
from typing import Union
NumberType = TypeVar('NumberType', bound=Union[int, float])

def add_numbers(a: NumberType, b: NumberType) -> NumberType:
    return a + b
```

```csharp
// C# - 泛型约束
public interface ILengthwise
{
    int Length { get; }
}

public class GenericConstraints
{
    // where子句约束
    public static T LoggingIdentity<T>(T arg) where T : ILengthwise
    {
        Console.WriteLine(arg.Length);
        return arg;
    }

    // 多重约束
    public static T Process<T>(T item)
        where T : class, IComparable<T>, new()
    {
        // T必须是引用类型，实现IComparable，有无参构造函数
        return new T();
    }

    // 协变和逆变
    public static void PrintList<T>(IEnumerable<T> list)
        where T : IComparable<T>
    {
        foreach (var item in list.OrderBy(x => x))
        {
            Console.WriteLine(item);
        }
    }
}
```

```go
// Go - 类型约束（接口约束）
package main

import "fmt"

// 接口约束
type Lengthwise interface {
    Length() int
}

// 泛型函数约束
func LoggingIdentity[T Lengthwise](arg T) T {
    fmt.Println(arg.Length())
    return arg
}

// 类型集合约束
type Numeric interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64
}

func Add[T Numeric](a, b T) T {
    return a + b
}

// 实现示例
type MyString struct {
    value string
}

func (s MyString) Length() int {
    return len(s.value)
}

func main() {
    result := LoggingIdentity(MyString{"hello"})
    fmt.Println(result)

    sum := Add(1, 2)        // int
    sumFloat := Add(1.5, 2.5) // float64
    fmt.Println(sum, sumFloat)
}
```

**💡 泛型约束对比：**

| 特性           | TypeScript   | Java              | Python       | C#       | Go           |
| -------------- | ------------ | ----------------- | ------------ | -------- | ------------ |
| **约束语法**   | `extends`    | `extends/super`   | `bound=`     | `where`  | 接口约束     |
| **多重约束**   | `&` 交叉类型 | `&` 连接          | `Union` 类型 | `,` 分隔 | 接口嵌入     |
| **协变/逆变**  | 支持         | `? extends/super` | 支持         | `in/out` | 不支持       |
| **类型集合**   | 联合类型     | 不支持            | `Union`      | 不支持   | `~` 类型集合 |
| **运行时检查** | 无           | 有限              | 完整         | 完整     | 完整         |

### keyof约束

```typescript
// 使用keyof约束
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

// 约束类型参数为对象的键
function updateProperty<T, K extends keyof T>(obj: T, key: K, value: T[K]): T {
  return { ...obj, [key]: value };
}

const user = { id: 1, name: '张三', age: 30 };
const updatedUser = updateProperty(user, 'age', 31); // 类型安全
// updateProperty(user, "age", "31"); // 编译错误：类型不匹配
// updateProperty(user, "invalid", 31); // 编译错误：键不存在
```

### 条件约束

```typescript
// 条件约束
type NonNullable<T> = T extends null | undefined ? never : T;

type Example1 = NonNullable<string | null>; // string
type Example2 = NonNullable<number | undefined>; // number
type Example3 = NonNullable<boolean | null | undefined>; // boolean

// 复杂条件约束
type ApiEndpoint<T> = T extends string
  ? T extends `/${string}`
    ? T
    : never
  : never;

type ValidEndpoint = ApiEndpoint<'/users'>; // "/users"
type InvalidEndpoint = ApiEndpoint<'users'>; // never
```

---

## ❓ 条件类型

### 基础条件类型

条件类型基于类型关系进行类型选择：

```typescript
// 基础语法：T extends U ? X : Y
type IsString<T> = T extends string ? true : false;

type Test1 = IsString<string>; // true
type Test2 = IsString<number>; // false
type Test3 = IsString<'hello'>; // true

// 实用的条件类型
type NonNullable<T> = T extends null | undefined ? never : T;
type ReturnType<T> = T extends (...args: any[]) => infer R ? R : any;
type Parameters<T> = T extends (...args: infer P) => any ? P : never;
```

### 分布式条件类型

```typescript
// 分布式条件类型
type ToArray<T> = T extends any ? T[] : never;

type StringArray = ToArray<string>; // string[]
type NumberArray = ToArray<number>; // number[]
type UnionArray = ToArray<string | number>; // string[] | number[]

// 过滤联合类型
type Exclude<T, U> = T extends U ? never : T;
type Extract<T, U> = T extends U ? T : never;

type Numbers = Exclude<string | number | boolean, string>; // number | boolean
type Strings = Extract<string | number | boolean, string>; // string
```

### infer关键字

```typescript
// 使用infer推断类型
type ReturnType<T> = T extends (...args: any[]) => infer R ? R : any;

function getString(): string {
  return '';
}
function getNumber(): number {
  return 0;
}

type StringReturn = ReturnType<typeof getString>; // string
type NumberReturn = ReturnType<typeof getNumber>; // number

// 推断数组元素类型
type ArrayElement<T> = T extends (infer U)[] ? U : never;

type StringElement = ArrayElement<string[]>; // string
type NumberElement = ArrayElement<number[]>; // number

// 推断Promise类型
type PromiseType<T> = T extends Promise<infer U> ? U : never;

type StringPromise = PromiseType<Promise<string>>; // string
type NumberPromise = PromiseType<Promise<number>>; // number

// 复杂推断
type FunctionArgs<T> = T extends (...args: infer A) => any ? A : never;

function example(a: string, b: number, c: boolean): void {}
type ExampleArgs = FunctionArgs<typeof example>; // [string, number, boolean]
```

---

## 🗺️ 映射类型

### 基础映射类型

映射类型基于现有类型创建新类型：

```typescript
// 基础映射类型语法
type Readonly<T> = {
  readonly [P in keyof T]: T[P];
};

type Partial<T> = {
  [P in keyof T]?: T[P];
};

type Required<T> = {
  [P in keyof T]-?: T[P];
};

// 使用映射类型
interface User {
  id: number;
  name: string;
  email: string;
  age?: number;
}

type ReadonlyUser = Readonly<User>;
// {
//     readonly id: number;
//     readonly name: string;
//     readonly email: string;
//     readonly age?: number;
// }

type PartialUser = Partial<User>;
// {
//     id?: number;
//     name?: string;
//     email?: string;
//     age?: number;
// }

type RequiredUser = Required<User>;
// {
//     id: number;
//     name: string;
//     email: string;
//     age: number; // 注意：age变为必需
// }
```

### 高级映射类型

```typescript
// 键重映射
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

type UserGetters = Getters<User>;
// {
//     getId: () => number;
//     getName: () => string;
//     getEmail: () => string;
//     getAge: () => number | undefined;
// }

// 条件映射
type NonFunctionPropertyNames<T> = {
  [K in keyof T]: T[K] extends Function ? never : K;
}[keyof T];

type NonFunctionProperties<T> = Pick<T, NonFunctionPropertyNames<T>>;

class Example {
  name: string = '';
  age: number = 0;
  getName(): string {
    return this.name;
  }
  getAge(): number {
    return this.age;
  }
}

type ExampleData = NonFunctionProperties<Example>;
// { name: string; age: number; }

// 深度映射
type DeepReadonly<T> = {
  readonly [P in keyof T]: T[P] extends object ? DeepReadonly<T[P]> : T[P];
};

interface NestedUser {
  id: number;
  profile: {
    name: string;
    settings: {
      theme: string;
      notifications: boolean;
    };
  };
}

type DeepReadonlyUser = DeepReadonly<NestedUser>;
// 所有嵌套属性都变为readonly
```

---

## 📝 模板字面量类型

### 基础模板字面量

```typescript
// 基础模板字面量类型
type World = 'world';
type Greeting = `hello ${World}`; // "hello world"

// 联合类型的模板字面量
type EmailLocaleIDs = 'welcome_email' | 'email_heading';
type FooterLocaleIDs = 'footer_title' | 'footer_sendoff';

type AllLocaleIDs = `${EmailLocaleIDs | FooterLocaleIDs}_id`;
// "welcome_email_id" | "email_heading_id" | "footer_title_id" | "footer_sendoff_id"

// 实用的模板字面量类型
type EventName<T extends string> = `on${Capitalize<T>}`;

type ClickEvent = EventName<'click'>; // "onClick"
type ChangeEvent = EventName<'change'>; // "onChange"
```

### 高级模板字面量应用

```typescript
// API路径类型
type ApiPath<T extends string> = `/api/${T}`;
type UserPaths = ApiPath<'users' | 'products' | 'orders'>;
// "/api/users" | "/api/products" | "/api/orders"

// 状态机类型
type State = 'idle' | 'loading' | 'success' | 'error';
type Action = 'fetch' | 'reset' | 'retry';
type StateAction = `${State}_${Action}`;
// "idle_fetch" | "idle_reset" | "idle_retry" | "loading_fetch" | ...

// CSS属性类型
type CSSProperty = 'margin' | 'padding' | 'border';
type CSSDirection = 'top' | 'right' | 'bottom' | 'left';
type CSSPropertyWithDirection = `${CSSProperty}-${CSSDirection}`;
// "margin-top" | "margin-right" | "margin-bottom" | "margin-left" | ...

// 数据库字段类型
type TableName = 'users' | 'products' | 'orders';
type FieldName = 'id' | 'name' | 'created_at';
type DatabaseField = `${TableName}.${FieldName}`;
// "users.id" | "users.name" | "users.created_at" | "products.id" | ...
```

---

## 🛠️ 实用工具类型

### 内置工具类型

TypeScript提供了许多内置的实用工具类型：

```typescript
// Partial<T> - 所有属性变为可选
interface User {
  id: number;
  name: string;
  email: string;
}

type PartialUser = Partial<User>;
// { id?: number; name?: string; email?: string; }

// Required<T> - 所有属性变为必需
type RequiredUser = Required<PartialUser>;
// { id: number; name: string; email: string; }

// Pick<T, K> - 选择指定属性
type UserBasic = Pick<User, 'id' | 'name'>;
// { id: number; name: string; }

// Omit<T, K> - 排除指定属性
type UserWithoutId = Omit<User, 'id'>;
// { name: string; email: string; }

// Record<K, T> - 创建键值对类型
type UserRoles = Record<string, 'admin' | 'user' | 'guest'>;
// { [key: string]: "admin" | "user" | "guest"; }

// Exclude<T, U> - 从联合类型中排除
type StringOrNumber = string | number | boolean;
type OnlyStringOrNumber = Exclude<StringOrNumber, boolean>;
// string | number

// Extract<T, U> - 从联合类型中提取
type OnlyString = Extract<StringOrNumber, string>;
// string

// NonNullable<T> - 排除null和undefined
type NonNullableString = NonNullable<string | null | undefined>;
// string

// ReturnType<T> - 获取函数返回类型
function getUser(): User {
  return {} as User;
}
type GetUserReturn = ReturnType<typeof getUser>;
// User

// Parameters<T> - 获取函数参数类型
function updateUser(id: number, data: Partial<User>): void {}
type UpdateUserParams = Parameters<typeof updateUser>;
// [number, Partial<User>]
```

### 自定义工具类型

```typescript
// 深度可选
type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

// 深度必需
type DeepRequired<T> = {
  [P in keyof T]-?: T[P] extends object ? DeepRequired<T[P]> : T[P];
};

// 可空类型
type Nullable<T> = T | null;

// 可选键
type OptionalKeys<T> = {
  [K in keyof T]-?: {} extends Pick<T, K> ? K : never;
}[keyof T];

// 必需键
type RequiredKeys<T> = {
  [K in keyof T]-?: {} extends Pick<T, K> ? never : K;
}[keyof T];

// 函数属性
type FunctionPropertyNames<T> = {
  [K in keyof T]: T[K] extends Function ? K : never;
}[keyof T];

type FunctionProperties<T> = Pick<T, FunctionPropertyNames<T>>;

// 非函数属性
type NonFunctionPropertyNames<T> = {
  [K in keyof T]: T[K] extends Function ? never : K;
}[keyof T];

type NonFunctionProperties<T> = Pick<T, NonFunctionPropertyNames<T>>;

// 使用示例
interface ExampleInterface {
  id: number;
  name?: string;
  email: string;
  getName(): string;
  setName(name: string): void;
}

type OptionalKeysExample = OptionalKeys<ExampleInterface>; // "name"
type RequiredKeysExample = RequiredKeys<ExampleInterface>; // "id" | "email"
type FunctionPropsExample = FunctionProperties<ExampleInterface>;
// { getName(): string; setName(name: string): void; }
type NonFunctionPropsExample = NonFunctionProperties<ExampleInterface>;
// { id: number; name?: string; email: string; }
```

---

## 🧠 高级类型编程

### 递归类型

```typescript
// JSON类型定义
type JSONValue = string | number | boolean | null | JSONObject | JSONArray;

interface JSONObject {
  [key: string]: JSONValue;
}

interface JSONArray extends Array<JSONValue> {}

// 路径类型
type Path<T> = T extends object
  ? {
      [K in keyof T]: K extends string
        ? T[K] extends object
          ? K | `${K}.${Path<T[K]>}`
          : K
        : never;
    }[keyof T]
  : never;

interface NestedObject {
  user: {
    profile: {
      name: string;
      age: number;
    };
    settings: {
      theme: string;
    };
  };
  products: {
    list: string[];
  };
}

type NestedPaths = Path<NestedObject>;
// "user" | "products" | "user.profile" | "user.settings" | "user.profile.name" |
// "user.profile.age" | "user.settings.theme" | "products.list"
```

### 类型体操

```typescript
// 字符串操作类型
type Reverse<S extends string> = S extends `${infer First}${infer Rest}`
  ? `${Reverse<Rest>}${First}`
  : '';

type ReversedHello = Reverse<'hello'>; // "olleh"

// 数组长度计算
type Length<T extends readonly any[]> = T['length'];

type ArrayLength = Length<[1, 2, 3, 4, 5]>; // 5

// 元组转联合
type TupleToUnion<T extends readonly any[]> = T[number];

type Union = TupleToUnion<[string, number, boolean]>; // string | number | boolean

// 联合转交叉
type UnionToIntersection<U> = (U extends any ? (k: U) => void : never) extends (
  k: infer I
) => void
  ? I
  : never;

type Intersection = UnionToIntersection<{ a: string } | { b: number }>;
// { a: string } & { b: number }

// 函数重载类型
type Overload = {
  (x: string): string;
  (x: number): number;
  (x: boolean): boolean;
};

type OverloadReturn = ReturnType<Overload>; // string | number | boolean
```

---

## 🛍️ Mall-Frontend实战案例

### API客户端泛型设计

基于Mall-Frontend项目的实际API设计，我们来看如何使用泛型构建类型安全的API客户端：

<augment_code_snippet path="mall-frontend/src/services/api.ts" mode="EXCERPT">

```typescript
// 认证相关API
export const authAPI = {
  // 用户登录
  login: (
    data: LoginRequest
  ): Promise<
    ApiResponse<{
      user: User;
      token: string;
      refresh_token: string;
    }>
  > => {
    return http.post(API_ENDPOINTS.AUTH.LOGIN, data, {
      skipAuth: true,
      showSuccessMessage: true,
      successMessage: '登录成功',
    });
  },

  // 获取用户信息
  getProfile: (): Promise<ApiResponse<User>> => {
    return http.get(API_ENDPOINTS.AUTH.PROFILE);
  },
};

// 商品相关API
export const productAPI = {
  // 获取商品列表
  getProducts: (
    params: PaginationParams & {
      category_id?: number;
      status?: string;
      min_price?: number;
      max_price?: number;
    }
  ): Promise<ApiResponse<PageResult<Product>>> => {
    return http.get(API_ENDPOINTS.PRODUCTS.LIST, { params });
  },
};
```

</augment_code_snippet>

### 改进的泛型API客户端设计

```typescript
// 定义API端点映射
interface ApiEndpoints {
  // 认证相关
  'POST /auth/login': {
    request: LoginRequest;
    response: { user: User; token: string; refresh_token: string };
  };
  'GET /auth/profile': {
    request: void;
    response: User;
  };
  'POST /auth/logout': {
    request: void;
    response: null;
  };

  // 用户管理
  'GET /users': {
    request: PaginationParams;
    response: PageResult<User>;
  };
  'GET /users/:id': {
    request: { id: number };
    response: User;
  };
  'PUT /users/:id': {
    request: { id: number; data: Partial<User> };
    response: User;
  };

  // 商品管理
  'GET /products': {
    request: PaginationParams & {
      category_id?: number;
      status?: string;
      min_price?: number;
      max_price?: number;
    };
    response: PageResult<Product>;
  };
  'POST /products': {
    request: Omit<Product, 'id' | 'created_at' | 'updated_at'>;
    response: Product;
  };
}

// 提取HTTP方法和路径
type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';
type ExtractMethod<T> = T extends `${infer M} ${string}` ? M : never;
type ExtractPath<T> = T extends `${string} ${infer P}` ? P : never;

// 类型安全的API客户端
class TypedApiClient {
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  // 通用请求方法
  async request<K extends keyof ApiEndpoints>(
    endpoint: K,
    ...args: ApiEndpoints[K]['request'] extends void
      ? []
      : [ApiEndpoints[K]['request']]
  ): Promise<ApiResponse<ApiEndpoints[K]['response']>> {
    const method = this.extractMethod(endpoint);
    const path = this.extractPath(endpoint);
    const [data] = args;

    // 实现具体的HTTP请求逻辑
    const response = await fetch(`${this.baseURL}${path}`, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: data ? JSON.stringify(data) : undefined,
    });

    return response.json();
  }

  // GET请求的便捷方法
  async get<K extends keyof ApiEndpoints>(
    endpoint: K extends `GET ${string}` ? K : never,
    ...args: ApiEndpoints[K]['request'] extends void
      ? []
      : [ApiEndpoints[K]['request']]
  ): Promise<ApiResponse<ApiEndpoints[K]['response']>> {
    return this.request(endpoint, ...args);
  }

  // POST请求的便捷方法
  async post<K extends keyof ApiEndpoints>(
    endpoint: K extends `POST ${string}` ? K : never,
    data: ApiEndpoints[K]['request']
  ): Promise<ApiResponse<ApiEndpoints[K]['response']>> {
    return this.request(endpoint, data);
  }

  private extractMethod(endpoint: string): HttpMethod {
    return endpoint.split(' ')[0] as HttpMethod;
  }

  private extractPath(endpoint: string): string {
    return endpoint.split(' ')[1];
  }
}

// 使用示例
const apiClient = new TypedApiClient('/api/v1');

// 类型安全的API调用
const loginResult = await apiClient.post('POST /auth/login', {
  username: 'john',
  password: 'password123',
});
// loginResult的类型自动推断为 ApiResponse<{ user: User; token: string; refresh_token: string }>

const userProfile = await apiClient.get('GET /auth/profile');
// userProfile的类型自动推断为 ApiResponse<User>

const products = await apiClient.get('GET /products', {
  page: 1,
  page_size: 20,
  category_id: 1,
});
// products的类型自动推断为 ApiResponse<PageResult<Product>>
```

### 状态管理中的泛型应用

```typescript
// 通用的异步状态类型
interface AsyncState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

// 通用的分页状态类型
interface PaginatedState<T> extends AsyncState<T[]> {
  pagination: {
    page: number;
    pageSize: number;
    total: number;
  };
}

// 通用的CRUD状态管理
interface CrudState<T> {
  list: PaginatedState<T>;
  detail: AsyncState<T>;
  create: AsyncState<T>;
  update: AsyncState<T>;
  delete: AsyncState<boolean>;
}

// 具体的状态类型
type UserState = CrudState<User>;
type ProductState = CrudState<Product>;
type OrderState = CrudState<Order>;

// 通用的异步Action创建器
function createAsyncActions<T, P = void>(name: string) {
  return {
    request: (payload: P) => ({ type: `${name}_REQUEST`, payload }),
    success: (payload: T) => ({ type: `${name}_SUCCESS`, payload }),
    failure: (error: string) => ({ type: `${name}_FAILURE`, payload: error }),
  };
}

// 使用泛型创建具体的actions
const userActions = {
  fetchUsers: createAsyncActions<PageResult<User>, PaginationParams>(
    'FETCH_USERS'
  ),
  fetchUserDetail: createAsyncActions<User, number>('FETCH_USER_DETAIL'),
  createUser: createAsyncActions<User, Omit<User, 'id'>>('CREATE_USER'),
  updateUser: createAsyncActions<User, { id: number; data: Partial<User> }>(
    'UPDATE_USER'
  ),
  deleteUser: createAsyncActions<boolean, number>('DELETE_USER'),
};

// 通用的异步reducer
function createAsyncReducer<T>(initialState: AsyncState<T>) {
  return (state = initialState, action: any): AsyncState<T> => {
    const { type, payload } = action;

    if (type.endsWith('_REQUEST')) {
      return { ...state, loading: true, error: null };
    }

    if (type.endsWith('_SUCCESS')) {
      return { ...state, loading: false, data: payload, error: null };
    }

    if (type.endsWith('_FAILURE')) {
      return { ...state, loading: false, error: payload };
    }

    return state;
  };
}
```

### 表单验证的泛型设计

```typescript
// 验证规则类型
type ValidationRule<T> = {
  required?: boolean;
  min?: number;
  max?: number;
  pattern?: RegExp;
  custom?: (value: T) => string | null;
};

// 表单验证配置
type FormValidation<T> = {
  [K in keyof T]?: ValidationRule<T[K]>;
};

// 验证结果类型
type ValidationResult<T> = {
  [K in keyof T]?: string;
};

// 通用表单验证器
class FormValidator<T extends Record<string, any>> {
  private rules: FormValidation<T>;

  constructor(rules: FormValidation<T>) {
    this.rules = rules;
  }

  validate(data: T): ValidationResult<T> {
    const errors: ValidationResult<T> = {};

    for (const key in this.rules) {
      const rule = this.rules[key];
      const value = data[key];

      if (rule?.required && (!value || value === '')) {
        errors[key] = `${String(key)}是必填项`;
        continue;
      }

      if (rule?.min && value && value.length < rule.min) {
        errors[key] = `${String(key)}最少${rule.min}个字符`;
        continue;
      }

      if (rule?.max && value && value.length > rule.max) {
        errors[key] = `${String(key)}最多${rule.max}个字符`;
        continue;
      }

      if (rule?.pattern && value && !rule.pattern.test(value)) {
        errors[key] = `${String(key)}格式不正确`;
        continue;
      }

      if (rule?.custom && value) {
        const customError = rule.custom(value);
        if (customError) {
          errors[key] = customError;
        }
      }
    }

    return errors;
  }

  isValid(data: T): boolean {
    const errors = this.validate(data);
    return Object.keys(errors).length === 0;
  }
}

// 使用示例
interface LoginForm {
  username: string;
  password: string;
  remember: boolean;
}

const loginValidator = new FormValidator<LoginForm>({
  username: {
    required: true,
    min: 3,
    max: 20,
    pattern: /^[a-zA-Z0-9_]+$/,
  },
  password: {
    required: true,
    min: 6,
    custom: value => {
      if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(value)) {
        return '密码必须包含大小写字母和数字';
      }
      return null;
    },
  },
});

const formData: LoginForm = {
  username: 'john',
  password: 'password123',
  remember: false,
};

const errors = loginValidator.validate(formData);
const isValid = loginValidator.isValid(formData);
```

---

## 🎯 面试常考知识点

### 1. 泛型的核心概念

**Q: 什么是泛型？为什么需要泛型？**

**A: 泛型的核心价值：**

- **代码复用**：一个函数/类可以处理多种类型
- **类型安全**：在编译时保证类型正确性
- **性能优化**：避免运行时类型检查和装箱拆箱
- **更好的IDE支持**：提供准确的类型提示和重构

```typescript
// 没有泛型的问题
function identityAny(arg: any): any {
  return arg; // 丢失了类型信息
}

// 使用泛型的解决方案
function identity<T>(arg: T): T {
  return arg; // 保持类型信息
}

const result = identity('hello'); // 类型为string，不是any
```

### 2. 泛型约束的深度理解

**Q: 什么是泛型约束？如何使用？**

**A: 泛型约束限制类型参数的范围：**

```typescript
// 基础约束
interface Lengthwise {
  length: number;
}

function loggingIdentity<T extends Lengthwise>(arg: T): T {
  console.log(arg.length); // 可以安全访问length属性
  return arg;
}

// keyof约束
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key]; // 确保key是obj的有效属性
}

// 条件约束
type ApiResponse<T> = T extends string ? { message: T } : { data: T };
```

### 3. 映射类型的应用

**Q: 如何使用映射类型创建实用工具类型？**

**A: 映射类型的常见应用模式：**

```typescript
// 创建可选版本
type Partial<T> = {
  [P in keyof T]?: T[P];
};

// 创建只读版本
type Readonly<T> = {
  readonly [P in keyof T]: T[P];
};

// 选择特定属性
type Pick<T, K extends keyof T> = {
  [P in K]: T[P];
};

// 排除特定属性
type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;

// 键重映射
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};
```

### 4. 条件类型的高级应用

**Q: 如何使用条件类型实现类型推断？**

**A: 条件类型与infer的组合使用：**

```typescript
// 推断函数返回类型
type ReturnType<T> = T extends (...args: any[]) => infer R ? R : any;

// 推断数组元素类型
type ArrayElement<T> = T extends (infer U)[] ? U : never;

// 推断Promise类型
type PromiseType<T> = T extends Promise<infer U> ? U : never;

// 推断函数参数类型
type Parameters<T> = T extends (...args: infer P) => any ? P : never;

// 复杂的类型推断
type DeepReadonly<T> = {
  readonly [P in keyof T]: T[P] extends object ? DeepReadonly<T[P]> : T[P];
};
```

### 5. 实际项目中的泛型设计

**Q: 在实际项目中如何设计泛型架构？**

**A: 企业级泛型设计原则：**

```typescript
// 1. API客户端的泛型设计
interface ApiClient {
  get<T>(url: string): Promise<ApiResponse<T>>;
  post<T, U>(url: string, data: T): Promise<ApiResponse<U>>;
  put<T, U>(url: string, data: T): Promise<ApiResponse<U>>;
  delete<T>(url: string): Promise<ApiResponse<T>>;
}

// 2. 状态管理的泛型设计
interface AsyncState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

// 3. 表单处理的泛型设计
interface FormConfig<T> {
  fields: {
    [K in keyof T]: FieldConfig<T[K]>;
  };
  validation: ValidationRules<T>;
  onSubmit: (data: T) => Promise<void>;
}

// 4. 数据仓库的泛型设计
interface Repository<T, K = number> {
  findById(id: K): Promise<T | null>;
  findAll(filter?: Partial<T>): Promise<T[]>;
  create(data: Omit<T, 'id'>): Promise<T>;
  update(id: K, data: Partial<T>): Promise<T>;
  delete(id: K): Promise<boolean>;
}
```

---

## 🏋️ 实战练习

### 练习1: 设计类型安全的事件系统

**题目**: 为电商系统设计一个类型安全的事件发布订阅系统

**要求**:

1. 支持多种业务事件类型
2. 事件数据的类型安全
3. 支持事件过滤和转换
4. 提供异步事件处理

**解决方案**:

```typescript
// 事件类型定义
interface EventMap {
  // 用户事件
  'user:registered': { user: User; timestamp: Date };
  'user:login': { user: User; ip: string; userAgent: string };
  'user:logout': { userId: number; sessionDuration: number };
  'user:profile_updated': { userId: number; changes: Partial<User> };

  // 商品事件
  'product:created': { product: Product; creator: User };
  'product:updated': {
    productId: number;
    changes: Partial<Product>;
    updater: User;
  };
  'product:deleted': { productId: number; product: Product; deleter: User };
  'product:stock_changed': {
    productId: number;
    oldStock: number;
    newStock: number;
  };

  // 订单事件
  'order:created': { order: Order; items: OrderItem[] };
  'order:paid': { orderId: number; payment: Payment };
  'order:shipped': { orderId: number; trackingNumber: string };
  'order:delivered': { orderId: number; deliveryTime: Date };
  'order:cancelled': { orderId: number; reason: string; refund?: Payment };

  // 购物车事件
  'cart:item_added': { userId: number; item: CartItem };
  'cart:item_removed': { userId: number; itemId: number };
  'cart:cleared': { userId: number; itemCount: number };
}

// 事件监听器类型
type EventListener<T> = (data: T) => void | Promise<void>;

// 事件过滤器类型
type EventFilter<T> = (data: T) => boolean;

// 事件转换器类型
type EventTransformer<T, U> = (data: T) => U;

// 订阅配置
interface SubscriptionConfig<T> {
  listener: EventListener<T>;
  filter?: EventFilter<T>;
  once?: boolean;
  priority?: number;
}

// 类型安全的事件发射器
class TypedEventEmitter {
  private listeners: {
    [K in keyof EventMap]?: Array<{
      config: SubscriptionConfig<EventMap[K]>;
      id: string;
    }>;
  } = {};

  private eventHistory: Array<{
    event: keyof EventMap;
    data: any;
    timestamp: Date;
  }> = [];

  // 订阅事件
  on<K extends keyof EventMap>(
    event: K,
    config: SubscriptionConfig<EventMap[K]>
  ): () => void {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }

    const id = this.generateId();
    this.listeners[event]!.push({ config, id });

    // 按优先级排序
    this.listeners[event]!.sort(
      (a, b) => (b.config.priority || 0) - (a.config.priority || 0)
    );

    // 返回取消订阅函数
    return () => this.off(event, id);
  }

  // 简化的订阅方法
  subscribe<K extends keyof EventMap>(
    event: K,
    listener: EventListener<EventMap[K]>,
    filter?: EventFilter<EventMap[K]>
  ): () => void {
    return this.on(event, { listener, filter });
  }

  // 一次性订阅
  once<K extends keyof EventMap>(
    event: K,
    listener: EventListener<EventMap[K]>
  ): void {
    this.on(event, { listener, once: true });
  }

  // 取消订阅
  private off<K extends keyof EventMap>(event: K, id: string): void {
    const eventListeners = this.listeners[event];
    if (eventListeners) {
      const index = eventListeners.findIndex(item => item.id === id);
      if (index > -1) {
        eventListeners.splice(index, 1);
      }
    }
  }

  // 发布事件
  async emit<K extends keyof EventMap>(
    event: K,
    data: EventMap[K]
  ): Promise<void> {
    // 记录事件历史
    this.eventHistory.push({
      event,
      data,
      timestamp: new Date(),
    });

    const eventListeners = this.listeners[event];
    if (!eventListeners) return;

    // 并行处理所有监听器
    const promises = eventListeners
      .filter(({ config }) => !config.filter || config.filter(data))
      .map(async ({ config, id }) => {
        try {
          await config.listener(data);

          // 如果是一次性监听器，移除它
          if (config.once) {
            this.off(event, id);
          }
        } catch (error) {
          console.error(`Error in event listener for ${String(event)}:`, error);
        }
      });

    await Promise.all(promises);
  }

  // 批量发布事件
  async emitBatch<K extends keyof EventMap>(
    events: Array<{ event: K; data: EventMap[K] }>
  ): Promise<void> {
    await Promise.all(events.map(({ event, data }) => this.emit(event, data)));
  }

  // 事件转换和转发
  transform<K extends keyof EventMap, T>(
    sourceEvent: K,
    targetEvent: keyof EventMap,
    transformer: EventTransformer<EventMap[K], T>
  ): () => void {
    return this.on(sourceEvent, {
      listener: async data => {
        const transformedData = transformer(data);
        await this.emit(targetEvent as any, transformedData as any);
      },
    });
  }

  // 获取事件历史
  getEventHistory<K extends keyof EventMap>(
    event?: K,
    limit?: number
  ): Array<{ event: keyof EventMap; data: any; timestamp: Date }> {
    let history = this.eventHistory;

    if (event) {
      history = history.filter(item => item.event === event);
    }

    if (limit) {
      history = history.slice(-limit);
    }

    return history;
  }

  // 清理事件历史
  clearHistory(): void {
    this.eventHistory = [];
  }

  // 移除所有监听器
  removeAllListeners<K extends keyof EventMap>(event?: K): void {
    if (event) {
      delete this.listeners[event];
    } else {
      this.listeners = {};
    }
  }

  private generateId(): string {
    return Math.random().toString(36).substr(2, 9);
  }
}

// 使用示例
const eventBus = new TypedEventEmitter();

// 订阅用户注册事件
const unsubscribeUserRegistered = eventBus.subscribe(
  'user:registered',
  async ({ user, timestamp }) => {
    // 发送欢迎邮件
    await sendWelcomeEmail(user.email);
    console.log(`用户 ${user.username} 在 ${timestamp} 注册`);
  }
);

// 订阅订单创建事件（带过滤器）
eventBus.subscribe(
  'order:created',
  async ({ order, items }) => {
    // 只处理大额订单
    await processLargeOrder(order, items);
  },
  ({ order }) => parseFloat(order.total_amount) > 1000 // 过滤器：只处理1000元以上的订单
);

// 一次性订阅
eventBus.once('user:login', ({ user, ip }) => {
  console.log(`用户 ${user.username} 首次登录，IP: ${ip}`);
});

// 事件转换：将商品库存变化转换为库存警告
eventBus.transform(
  'product:stock_changed',
  'product:low_stock_warning' as any,
  ({ productId, newStock }) => ({
    productId,
    currentStock: newStock,
    threshold: 10,
    severity: newStock < 5 ? 'critical' : 'warning',
  })
);

// 发布事件
await eventBus.emit('user:registered', {
  user: { id: 1, username: 'john', email: 'john@example.com' } as User,
  timestamp: new Date(),
});

await eventBus.emit('order:created', {
  order: { id: 1, total_amount: '1500.00' } as Order,
  items: [],
});

// 批量发布事件
await eventBus.emitBatch([
  {
    event: 'product:stock_changed',
    data: { productId: 1, oldStock: 15, newStock: 3 },
  },
  {
    event: 'cart:item_added',
    data: { userId: 1, item: {} as CartItem },
  },
]);

// 查看事件历史
const recentEvents = eventBus.getEventHistory(undefined, 10);
console.log('最近10个事件:', recentEvents);
```

这个事件系统的特点：

1. **类型安全**: 所有事件类型和数据都有严格的类型检查
2. **灵活过滤**: 支持基于数据内容的事件过滤
3. **优先级处理**: 支持监听器优先级排序
4. **异步处理**: 支持异步事件监听器
5. **事件转换**: 支持事件数据转换和转发
6. **历史记录**: 保存事件历史用于调试和分析
7. **错误处理**: 单个监听器错误不影响其他监听器

### 练习2: 通用数据仓库设计

**题目**: 设计一个通用的数据仓库模式，支持不同的数据源

**要求**:

1. 支持多种数据源（API、LocalStorage、IndexedDB）
2. 统一的CRUD接口
3. 缓存策略支持
4. 类型安全的查询构建器

**解决方案**:

```typescript
// 数据源接口
interface DataSource<T, K = string | number> {
  findById(id: K): Promise<T | null>;
  findAll(filter?: Partial<T>): Promise<T[]>;
  create(data: Omit<T, 'id'>): Promise<T>;
  update(id: K, data: Partial<T>): Promise<T>;
  delete(id: K): Promise<boolean>;
  count(filter?: Partial<T>): Promise<number>;
}

// 查询构建器
class QueryBuilder<T> {
  private filters: Array<(item: T) => boolean> = [];
  private sortFn?: (a: T, b: T) => number;
  private limitValue?: number;
  private offsetValue?: number;

  where<K extends keyof T>(
    field: K,
    operator: '=' | '!=' | '>' | '<' | '>=' | '<=' | 'in' | 'like',
    value: T[K] | T[K][]
  ): this {
    this.filters.push(item => {
      const fieldValue = item[field];

      switch (operator) {
        case '=':
          return fieldValue === value;
        case '!=':
          return fieldValue !== value;
        case '>':
          return fieldValue > value;
        case '<':
          return fieldValue < value;
        case '>=':
          return fieldValue >= value;
        case '<=':
          return fieldValue <= value;
        case 'in':
          return Array.isArray(value) && value.includes(fieldValue);
        case 'like':
          return String(fieldValue).includes(String(value));
        default:
          return true;
      }
    });
    return this;
  }

  orderBy<K extends keyof T>(
    field: K,
    direction: 'asc' | 'desc' = 'asc'
  ): this {
    this.sortFn = (a, b) => {
      const aVal = a[field];
      const bVal = b[field];
      const result = aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
      return direction === 'desc' ? -result : result;
    };
    return this;
  }

  limit(count: number): this {
    this.limitValue = count;
    return this;
  }

  offset(count: number): this {
    this.offsetValue = count;
    return this;
  }

  execute(data: T[]): T[] {
    let result = data;

    // 应用过滤器
    for (const filter of this.filters) {
      result = result.filter(filter);
    }

    // 应用排序
    if (this.sortFn) {
      result = result.sort(this.sortFn);
    }

    // 应用偏移和限制
    if (this.offsetValue) {
      result = result.slice(this.offsetValue);
    }

    if (this.limitValue) {
      result = result.slice(0, this.limitValue);
    }

    return result;
  }
}

// API数据源实现
class ApiDataSource<T extends { id: any }> implements DataSource<T> {
  constructor(
    private baseUrl: string,
    private httpClient: any
  ) {}

  async findById(id: T['id']): Promise<T | null> {
    try {
      const response = await this.httpClient.get(`${this.baseUrl}/${id}`);
      return response.data;
    } catch (error) {
      if (error.status === 404) return null;
      throw error;
    }
  }

  async findAll(filter?: Partial<T>): Promise<T[]> {
    const response = await this.httpClient.get(this.baseUrl, {
      params: filter,
    });
    return response.data;
  }

  async create(data: Omit<T, 'id'>): Promise<T> {
    const response = await this.httpClient.post(this.baseUrl, data);
    return response.data;
  }

  async update(id: T['id'], data: Partial<T>): Promise<T> {
    const response = await this.httpClient.put(`${this.baseUrl}/${id}`, data);
    return response.data;
  }

  async delete(id: T['id']): Promise<boolean> {
    try {
      await this.httpClient.delete(`${this.baseUrl}/${id}`);
      return true;
    } catch (error) {
      return false;
    }
  }

  async count(filter?: Partial<T>): Promise<number> {
    const response = await this.httpClient.get(`${this.baseUrl}/count`, {
      params: filter,
    });
    return response.data.count;
  }
}

// 内存数据源实现
class MemoryDataSource<T extends { id: any }> implements DataSource<T> {
  private data: T[] = [];
  private nextId = 1;

  async findById(id: T['id']): Promise<T | null> {
    return this.data.find(item => item.id === id) || null;
  }

  async findAll(filter?: Partial<T>): Promise<T[]> {
    if (!filter) return [...this.data];

    return this.data.filter(item => {
      return Object.entries(filter).every(
        ([key, value]) => item[key as keyof T] === value
      );
    });
  }

  async create(data: Omit<T, 'id'>): Promise<T> {
    const newItem = { ...data, id: this.nextId++ } as T;
    this.data.push(newItem);
    return newItem;
  }

  async update(id: T['id'], data: Partial<T>): Promise<T> {
    const index = this.data.findIndex(item => item.id === id);
    if (index === -1) throw new Error('Item not found');

    this.data[index] = { ...this.data[index], ...data };
    return this.data[index];
  }

  async delete(id: T['id']): Promise<boolean> {
    const index = this.data.findIndex(item => item.id === id);
    if (index === -1) return false;

    this.data.splice(index, 1);
    return true;
  }

  async count(filter?: Partial<T>): Promise<number> {
    const items = await this.findAll(filter);
    return items.length;
  }
}

// 缓存策略接口
interface CacheStrategy<T> {
  get(key: string): Promise<T | null>;
  set(key: string, value: T, ttl?: number): Promise<void>;
  delete(key: string): Promise<void>;
  clear(): Promise<void>;
}

// 内存缓存实现
class MemoryCache<T> implements CacheStrategy<T> {
  private cache = new Map<string, { value: T; expiry?: number }>();

  async get(key: string): Promise<T | null> {
    const item = this.cache.get(key);
    if (!item) return null;

    if (item.expiry && Date.now() > item.expiry) {
      this.cache.delete(key);
      return null;
    }

    return item.value;
  }

  async set(key: string, value: T, ttl?: number): Promise<void> {
    const expiry = ttl ? Date.now() + ttl * 1000 : undefined;
    this.cache.set(key, { value, expiry });
  }

  async delete(key: string): Promise<void> {
    this.cache.delete(key);
  }

  async clear(): Promise<void> {
    this.cache.clear();
  }
}

// 通用仓库实现
class Repository<T extends { id: any }> {
  private cache?: CacheStrategy<T | T[]>;

  constructor(
    private dataSource: DataSource<T>,
    private cacheStrategy?: CacheStrategy<T | T[]>,
    private cacheTtl: number = 300 // 5分钟默认缓存
  ) {
    this.cache = cacheStrategy;
  }

  async findById(id: T['id']): Promise<T | null> {
    const cacheKey = `${this.constructor.name}:${id}`;

    // 尝试从缓存获取
    if (this.cache) {
      const cached = (await this.cache.get(cacheKey)) as T | null;
      if (cached) return cached;
    }

    // 从数据源获取
    const item = await this.dataSource.findById(id);

    // 缓存结果
    if (item && this.cache) {
      await this.cache.set(cacheKey, item, this.cacheTtl);
    }

    return item;
  }

  async findAll(filter?: Partial<T>): Promise<T[]> {
    const cacheKey = `${this.constructor.name}:all:${JSON.stringify(filter || {})}`;

    // 尝试从缓存获取
    if (this.cache) {
      const cached = (await this.cache.get(cacheKey)) as T[] | null;
      if (cached) return cached;
    }

    // 从数据源获取
    const items = await this.dataSource.findAll(filter);

    // 缓存结果
    if (this.cache) {
      await this.cache.set(cacheKey, items, this.cacheTtl);
    }

    return items;
  }

  async create(data: Omit<T, 'id'>): Promise<T> {
    const item = await this.dataSource.create(data);

    // 清除相关缓存
    if (this.cache) {
      await this.clearRelatedCache();
    }

    return item;
  }

  async update(id: T['id'], data: Partial<T>): Promise<T> {
    const item = await this.dataSource.update(id, data);

    // 更新缓存
    if (this.cache) {
      const cacheKey = `${this.constructor.name}:${id}`;
      await this.cache.set(cacheKey, item, this.cacheTtl);
      await this.clearRelatedCache();
    }

    return item;
  }

  async delete(id: T['id']): Promise<boolean> {
    const result = await this.dataSource.delete(id);

    // 清除缓存
    if (this.cache && result) {
      const cacheKey = `${this.constructor.name}:${id}`;
      await this.cache.delete(cacheKey);
      await this.clearRelatedCache();
    }

    return result;
  }

  async count(filter?: Partial<T>): Promise<number> {
    return this.dataSource.count(filter);
  }

  // 查询构建器
  query(): QueryBuilder<T> {
    return new QueryBuilder<T>();
  }

  // 执行查询
  async executeQuery(query: QueryBuilder<T>): Promise<T[]> {
    const allItems = await this.findAll();
    return query.execute(allItems);
  }

  private async clearRelatedCache(): Promise<void> {
    if (this.cache) {
      // 这里可以实现更智能的缓存清除策略
      // 目前简单地清除所有缓存
      await this.cache.clear();
    }
  }
}

// 使用示例
interface User {
  id: number;
  username: string;
  email: string;
  age: number;
  status: 'active' | 'inactive';
  created_at: string;
}

// 创建用户仓库
const userApiDataSource = new ApiDataSource<User>('/api/users', httpClient);
const userCache = new MemoryCache<User | User[]>();
const userRepository = new Repository(userApiDataSource, userCache);

// 基本CRUD操作
const user = await userRepository.findById(1);
const users = await userRepository.findAll({ status: 'active' });
const newUser = await userRepository.create({
  username: 'john',
  email: 'john@example.com',
  age: 30,
  status: 'active',
  created_at: new Date().toISOString(),
});

// 使用查询构建器
const activeAdults = await userRepository.executeQuery(
  userRepository
    .query()
    .where('status', '=', 'active')
    .where('age', '>=', 18)
    .orderBy('created_at', 'desc')
    .limit(10)
);

// 复杂查询
const recentUsers = await userRepository.executeQuery(
  userRepository
    .query()
    .where('email', 'like', '@gmail.com')
    .where('age', 'in', [25, 30, 35])
    .orderBy('username', 'asc')
    .offset(20)
    .limit(10)
);
```

这个数据仓库系统的特点：

1. **多数据源支持**: 统一接口支持API、内存、LocalStorage等
2. **类型安全**: 完整的TypeScript类型支持
3. **缓存策略**: 可插拔的缓存实现
4. **查询构建器**: 类型安全的查询构建
5. **自动缓存管理**: 智能的缓存更新和清除
6. **扩展性**: 易于添加新的数据源和缓存策略

---

## 📚 本章总结

通过本章学习，我们深入掌握了TypeScript的泛型编程和高级类型应用：

### 🎯 核心收获

1. **泛型编程** 🧬
   - 掌握了泛型函数、接口、类的设计和使用
   - 理解了泛型约束和条件类型的应用场景
   - 学会了使用泛型构建可复用的代码架构

2. **高级类型** 🚀
   - 掌握了映射类型和模板字面量类型
   - 学会了使用条件类型和infer进行类型推断
   - 理解了TypeScript类型系统的强大表达能力

3. **实用工具类型** 🛠️
   - 熟练使用内置工具类型（Partial、Pick、Omit等）
   - 学会了设计自定义工具类型
   - 掌握了类型编程的常见模式

4. **企业级应用** 💼
   - 分析了Mall-Frontend项目中的泛型设计
   - 学会了在API客户端、状态管理中应用泛型
   - 掌握了大型项目的类型架构设计

### 🚀 技术进阶

- **下一步学习**: React组件设计与Hooks应用
- **实践建议**: 在项目中逐步引入高级类型特性
- **深入方向**: 类型编程和元编程技巧

### 💡 最佳实践

1. **渐进式复杂度**: 从简单泛型开始，逐步引入高级特性
2. **类型约束**: 合理使用约束提高类型安全性
3. **工具类型**: 充分利用内置和自定义工具类型
4. **性能考虑**: 避免过度复杂的类型计算

泛型是TypeScript最强大的特性之一，掌握它将让你的代码更加灵活、安全和可维护！ 🎉

---

_下一章我们将学习《React组件设计与Hooks应用》，探索如何在React中应用TypeScript的类型系统！_ 🚀
