# 第3章：组件库设计与开发 🧩

> *"好的组件库是团队效率的倍增器！"* 🚀

## 📚 本章导览

组件库是现代前端开发的基石，它不仅提供了可复用的UI组件，更是设计系统的技术实现。一个优秀的组件库能够显著提升开发效率、保证设计一致性、降低维护成本。本章将从组件库的设计理念出发，深入探讨组件库的架构设计、开发实践、工程化配置，以及与其他技术栈的对比分析。

### 🎯 学习目标

通过本章学习，你将掌握：

- **设计系统理论** - 理解设计系统的核心概念和价值
- **组件库架构** - 掌握可扩展的组件库架构设计
- **组件设计模式** - 学会各种组件设计模式和最佳实践
- **API设计原则** - 掌握组件API的设计原则和方法
- **工程化配置** - 学会组件库的构建、测试、发布流程
- **文档系统** - 构建完善的组件文档和示例系统
- **主题系统** - 实现灵活的主题定制和样式管理
- **性能优化** - 掌握组件库的性能优化技巧

### 🛠️ 技术栈概览

```typescript
{
  "componentLibrary": {
    "core": ["React", "TypeScript", "Styled-components", "CSS-in-JS"],
    "build": ["Rollup", "Webpack", "Vite", "Babel", "PostCSS"],
    "testing": ["Jest", "React Testing Library", "Storybook", "Chromatic"],
    "docs": ["Storybook", "Docusaurus", "VitePress", "Nextra"]
  },
  "designSystem": {
    "tokens": ["Design Tokens", "Style Dictionary", "Theo"],
    "tools": ["Figma", "Sketch", "Adobe XD", "Zeplin"],
    "frameworks": ["Ant Design", "Material-UI", "Chakra UI", "Mantine"]
  },
  "distribution": {
    "registry": ["NPM", "Yarn", "PNPM", "Private Registry"],
    "cdn": ["JSDelivr", "UNPKG", "CDN.js"],
    "bundling": ["ESM", "CJS", "UMD", "IIFE"]
  }
}
```

### 📖 本章目录

- [设计系统基础理论](#设计系统基础理论)
- [组件库架构设计](#组件库架构设计)
- [组件设计模式](#组件设计模式)
- [API设计原则](#api设计原则)
- [主题系统设计](#主题系统设计)
- [工程化配置](#工程化配置)
- [文档系统构建](#文档系统构建)
- [测试策略](#测试策略)
- [性能优化](#性能优化)
- [发布与维护](#发布与维护)
- [跨框架对比](#跨框架对比)
- [Mall-Frontend组件库](#mall-frontend组件库)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🎨 设计系统基础理论

### 设计系统的核心概念

设计系统是一套完整的设计标准、组件库和工具集，用于创建一致的用户体验：

```typescript
// 设计系统架构
interface DesignSystem {
  // 设计原则
  principles: {
    consistency: '一致性 - 保持视觉和交互的统一';
    accessibility: '可访问性 - 确保所有用户都能使用';
    scalability: '可扩展性 - 支持业务快速发展';
    efficiency: '效率性 - 提升设计和开发效率';
  };
  
  // 设计令牌 (Design Tokens)
  tokens: {
    colors: ColorTokens;
    typography: TypographyTokens;
    spacing: SpacingTokens;
    shadows: ShadowTokens;
    borders: BorderTokens;
    animations: AnimationTokens;
  };
  
  // 组件库
  components: {
    atoms: AtomicComponents;      // 原子组件：Button, Input, Icon
    molecules: MolecularComponents; // 分子组件：SearchBox, Card
    organisms: OrganismComponents;  // 有机体组件：Header, ProductList
    templates: TemplateComponents;  // 模板组件：PageLayout
    pages: PageComponents;         // 页面组件：HomePage, ProductPage
  };
  
  // 模式库
  patterns: {
    navigation: NavigationPatterns;
    forms: FormPatterns;
    feedback: FeedbackPatterns;
    data: DataPatterns;
  };
  
  // 工具和资源
  tools: {
    designTools: DesignTools;
    codeTools: CodeTools;
    documentation: Documentation;
    guidelines: Guidelines;
  };
}

// 设计令牌定义
interface ColorTokens {
  // 基础色彩
  primary: {
    50: '#f0f9ff';
    100: '#e0f2fe';
    200: '#bae6fd';
    300: '#7dd3fc';
    400: '#38bdf8';
    500: '#0ea5e9';  // 主色
    600: '#0284c7';
    700: '#0369a1';
    800: '#075985';
    900: '#0c4a6e';
  };
  
  // 语义色彩
  semantic: {
    success: '#10b981';
    warning: '#f59e0b';
    error: '#ef4444';
    info: '#3b82f6';
  };
  
  // 中性色彩
  neutral: {
    white: '#ffffff';
    gray: {
      50: '#f9fafb';
      100: '#f3f4f6';
      200: '#e5e7eb';
      300: '#d1d5db';
      400: '#9ca3af';
      500: '#6b7280';
      600: '#4b5563';
      700: '#374151';
      800: '#1f2937';
      900: '#111827';
    };
    black: '#000000';
  };
}

interface TypographyTokens {
  // 字体家族
  fontFamily: {
    sans: ['Inter', 'system-ui', 'sans-serif'];
    serif: ['Georgia', 'serif'];
    mono: ['Fira Code', 'monospace'];
  };
  
  // 字体大小
  fontSize: {
    xs: '0.75rem';    // 12px
    sm: '0.875rem';   // 14px
    base: '1rem';     // 16px
    lg: '1.125rem';   // 18px
    xl: '1.25rem';    // 20px
    '2xl': '1.5rem';  // 24px
    '3xl': '1.875rem'; // 30px
    '4xl': '2.25rem'; // 36px
    '5xl': '3rem';    // 48px
  };
  
  // 字重
  fontWeight: {
    thin: 100;
    light: 300;
    normal: 400;
    medium: 500;
    semibold: 600;
    bold: 700;
    extrabold: 800;
    black: 900;
  };
  
  // 行高
  lineHeight: {
    none: 1;
    tight: 1.25;
    snug: 1.375;
    normal: 1.5;
    relaxed: 1.625;
    loose: 2;
  };
  
  // 字间距
  letterSpacing: {
    tighter: '-0.05em';
    tight: '-0.025em';
    normal: '0em';
    wide: '0.025em';
    wider: '0.05em';
    widest: '0.1em';
  };
}

interface SpacingTokens {
  // 基础间距单位 (4px)
  unit: 4;
  
  // 间距比例
  scale: {
    0: '0px';
    1: '0.25rem';  // 4px
    2: '0.5rem';   // 8px
    3: '0.75rem';  // 12px
    4: '1rem';     // 16px
    5: '1.25rem';  // 20px
    6: '1.5rem';   // 24px
    8: '2rem';     // 32px
    10: '2.5rem';  // 40px
    12: '3rem';    // 48px
    16: '4rem';    // 64px
    20: '5rem';    // 80px
    24: '6rem';    // 96px
    32: '8rem';    // 128px
  };
  
  // 语义间距
  semantic: {
    xs: 'var(--spacing-1)';
    sm: 'var(--spacing-2)';
    md: 'var(--spacing-4)';
    lg: 'var(--spacing-6)';
    xl: 'var(--spacing-8)';
    '2xl': 'var(--spacing-12)';
  };
}
```

### 原子设计理论

原子设计是一种创建设计系统的方法论，将界面分解为基本构建块：

```typescript
// 原子设计层级结构
const atomicDesignLevels = {
  // 1. 原子 (Atoms) - 最基本的构建块
  atoms: {
    definition: '不能再分解的基本UI元素',
    examples: ['Button', 'Input', 'Label', 'Icon', 'Avatar'],
    characteristics: [
      '单一职责',
      '高度可复用',
      '无业务逻辑',
      '样式一致'
    ],
    implementation: `
      // Button原子组件
      interface ButtonProps {
        variant: 'primary' | 'secondary' | 'outline' | 'ghost';
        size: 'sm' | 'md' | 'lg';
        disabled?: boolean;
        loading?: boolean;
        children: React.ReactNode;
        onClick?: () => void;
      }
      
      const Button: React.FC<ButtonProps> = ({
        variant = 'primary',
        size = 'md',
        disabled = false,
        loading = false,
        children,
        onClick,
        ...props
      }) => {
        return (
          <StyledButton
            variant={variant}
            size={size}
            disabled={disabled || loading}
            onClick={onClick}
            {...props}
          >
            {loading && <Spinner />}
            {children}
          </StyledButton>
        );
      };
    `
  },
  
  // 2. 分子 (Molecules) - 原子的组合
  molecules: {
    definition: '由多个原子组合而成的相对简单的UI组件',
    examples: ['SearchBox', 'FormField', 'Card', 'Breadcrumb'],
    characteristics: [
      '功能相对完整',
      '可独立使用',
      '有简单交互',
      '可配置性强'
    ],
    implementation: `
      // SearchBox分子组件
      interface SearchBoxProps {
        placeholder?: string;
        value?: string;
        onChange?: (value: string) => void;
        onSearch?: (value: string) => void;
        loading?: boolean;
        disabled?: boolean;
      }
      
      const SearchBox: React.FC<SearchBoxProps> = ({
        placeholder = '请输入搜索关键词',
        value,
        onChange,
        onSearch,
        loading = false,
        disabled = false
      }) => {
        const [inputValue, setInputValue] = useState(value || '');
        
        const handleSearch = () => {
          onSearch?.(inputValue);
        };
        
        return (
          <SearchContainer>
            <Input
              placeholder={placeholder}
              value={inputValue}
              onChange={(e) => {
                setInputValue(e.target.value);
                onChange?.(e.target.value);
              }}
              disabled={disabled}
              onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
            />
            <Button
              variant="primary"
              onClick={handleSearch}
              loading={loading}
              disabled={disabled || !inputValue.trim()}
            >
              <SearchIcon />
            </Button>
          </SearchContainer>
        );
      };
    `
  },
  
  // 3. 有机体 (Organisms) - 分子和原子的复杂组合
  organisms: {
    definition: '由分子和原子组成的相对复杂的UI组件',
    examples: ['Header', 'ProductList', 'UserProfile', 'ShoppingCart'],
    characteristics: [
      '功能完整',
      '业务相关',
      '可独立工作',
      '复杂交互'
    ],
    implementation: `
      // ProductList有机体组件
      interface ProductListProps {
        products: Product[];
        loading?: boolean;
        onProductClick?: (product: Product) => void;
        onAddToCart?: (product: Product) => void;
        layout?: 'grid' | 'list';
        filters?: ProductFilters;
        onFiltersChange?: (filters: ProductFilters) => void;
      }
      
      const ProductList: React.FC<ProductListProps> = ({
        products,
        loading = false,
        onProductClick,
        onAddToCart,
        layout = 'grid',
        filters,
        onFiltersChange
      }) => {
        return (
          <ProductListContainer>
            <ProductFilters
              filters={filters}
              onChange={onFiltersChange}
            />
            
            <ProductGrid layout={layout}>
              {loading ? (
                <ProductSkeleton count={8} />
              ) : (
                products.map(product => (
                  <ProductCard
                    key={product.id}
                    product={product}
                    onClick={() => onProductClick?.(product)}
                    onAddToCart={() => onAddToCart?.(product)}
                  />
                ))
              )}
            </ProductGrid>
            
            {products.length === 0 && !loading && (
              <EmptyState
                title="暂无商品"
                description="当前筛选条件下没有找到商品"
                action={
                  <Button onClick={() => onFiltersChange?.({})}>
                    清除筛选
                  </Button>
                }
              />
            )}
          </ProductListContainer>
        );
      };
    `
  },
  
  // 4. 模板 (Templates) - 页面级别的组件布局
  templates: {
    definition: '定义页面结构和布局的组件模板',
    examples: ['PageLayout', 'DashboardLayout', 'AuthLayout'],
    characteristics: [
      '定义页面结构',
      '提供插槽',
      '响应式布局',
      '导航集成'
    ]
  },
  
  // 5. 页面 (Pages) - 具体的页面实例
  pages: {
    definition: '模板的具体实例，包含真实内容和数据',
    examples: ['HomePage', 'ProductDetailPage', 'CheckoutPage'],
    characteristics: [
      '具体业务实现',
      '数据集成',
      '路由处理',
      '状态管理'
    ]
  }
};
```

---

## 🏗️ 组件库架构设计

### 模块化架构

一个优秀的组件库需要清晰的模块化架构来保证可维护性和可扩展性：

```typescript
// 组件库目录结构
const componentLibraryStructure = {
  // 核心目录结构
  structure: `
    packages/
      core/                    # 核心包
        src/
          components/          # 组件源码
            atoms/            # 原子组件
              Button/
                Button.tsx
                Button.test.tsx
                Button.stories.tsx
                index.ts
            molecules/        # 分子组件
            organisms/        # 有机体组件
          hooks/              # 自定义Hooks
          utils/              # 工具函数
          types/              # 类型定义
          themes/             # 主题配置
          tokens/             # 设计令牌
        package.json

      icons/                  # 图标包
        src/
          icons/
          index.ts
        package.json

      themes/                 # 主题包
        src/
          themes/
          tokens/
        package.json

      utils/                  # 工具包
        src/
          helpers/
          validators/
          formatters/
        package.json

    apps/
      storybook/              # 文档和示例
      playground/             # 开发测试环境

    tools/
      build/                  # 构建工具
      eslint-config/          # ESLint配置
      tsconfig/               # TypeScript配置
  `,

  // 包依赖关系
  dependencies: {
    core: {
      dependencies: ['@mall-ui/tokens', '@mall-ui/utils'],
      peerDependencies: ['react', 'react-dom'],
      description: '核心组件库，包含所有UI组件'
    },
    icons: {
      dependencies: ['@mall-ui/core'],
      description: '图标库，提供SVG图标组件'
    },
    themes: {
      dependencies: ['@mall-ui/tokens'],
      description: '主题包，提供预设主题和主题工具'
    },
    utils: {
      dependencies: [],
      description: '工具函数库，提供通用工具函数'
    }
  }
};

// 组件架构设计
interface ComponentArchitecture {
  // 组件基类
  baseComponent: {
    props: BaseComponentProps;
    styling: ComponentStyling;
    accessibility: AccessibilityFeatures;
    testing: TestingSupport;
  };

  // 组件组合模式
  compositionPatterns: {
    compound: CompoundComponentPattern;
    renderProps: RenderPropsPattern;
    hoc: HigherOrderComponentPattern;
    hooks: HooksPattern;
  };

  // 样式系统
  stylingSystem: {
    styledComponents: StyledComponentsConfig;
    cssModules: CSSModulesConfig;
    cssInJs: CSSInJSConfig;
    designTokens: DesignTokensConfig;
  };
}

// 基础组件Props接口
interface BaseComponentProps {
  // 通用属性
  className?: string;
  style?: React.CSSProperties;
  id?: string;

  // 主题属性
  theme?: Theme;
  variant?: string;
  size?: 'sm' | 'md' | 'lg';

  // 状态属性
  disabled?: boolean;
  loading?: boolean;
  error?: boolean;

  // 可访问性属性
  'aria-label'?: string;
  'aria-describedby'?: string;
  'aria-labelledby'?: string;
  role?: string;
  tabIndex?: number;

  // 数据属性
  'data-testid'?: string;
  'data-cy'?: string;
}

// 组件工厂模式
const createComponent = <P extends BaseComponentProps>(
  displayName: string,
  defaultProps: Partial<P>,
  styleConfig: ComponentStyleConfig
) => {
  const Component = React.forwardRef<HTMLElement, P>((props, ref) => {
    const {
      className,
      style,
      theme,
      variant = 'default',
      size = 'md',
      disabled = false,
      loading = false,
      error = false,
      ...restProps
    } = { ...defaultProps, ...props };

    // 样式计算
    const computedStyles = useComputedStyles({
      theme,
      variant,
      size,
      disabled,
      loading,
      error,
      styleConfig
    });

    // 可访问性处理
    const accessibilityProps = useAccessibility({
      disabled,
      loading,
      error,
      ...restProps
    });

    return (
      <StyledComponent
        ref={ref}
        className={clsx(computedStyles.className, className)}
        style={{ ...computedStyles.style, ...style }}
        {...accessibilityProps}
        {...restProps}
      />
    );
  });

  Component.displayName = displayName;
  return Component;
};

// 复合组件模式示例
const Card = {
  // 主组件
  Root: createComponent<CardRootProps>('Card.Root', {}, cardRootStyles),

  // 子组件
  Header: createComponent<CardHeaderProps>('Card.Header', {}, cardHeaderStyles),
  Body: createComponent<CardBodyProps>('Card.Body', {}, cardBodyStyles),
  Footer: createComponent<CardFooterProps>('Card.Footer', {}, cardFooterStyles),

  // 组合组件
  Image: createComponent<CardImageProps>('Card.Image', {}, cardImageStyles),
  Title: createComponent<CardTitleProps>('Card.Title', {}, cardTitleStyles),
  Description: createComponent<CardDescriptionProps>('Card.Description', {}, cardDescriptionStyles),
  Actions: createComponent<CardActionsProps>('Card.Actions', {}, cardActionsStyles),
};

// 使用示例
const ProductCard = () => (
  <Card.Root>
    <Card.Image src="/product.jpg" alt="Product" />
    <Card.Body>
      <Card.Title>产品名称</Card.Title>
      <Card.Description>产品描述信息</Card.Description>
    </Card.Body>
    <Card.Footer>
      <Card.Actions>
        <Button variant="primary">加入购物车</Button>
        <Button variant="outline">查看详情</Button>
      </Card.Actions>
    </Card.Footer>
  </Card.Root>
);
```

### 类型系统设计

```typescript
// 组件库类型系统
namespace ComponentLibraryTypes {
  // 基础类型
  export type Size = 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  export type Variant = 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info';
  export type Color = 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info' | 'neutral';

  // 响应式类型
  export type ResponsiveValue<T> = T | {
    xs?: T;
    sm?: T;
    md?: T;
    lg?: T;
    xl?: T;
  };

  // 间距类型
  export type Spacing = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 8 | 10 | 12 | 16 | 20 | 24 | 32;
  export type ResponsiveSpacing = ResponsiveValue<Spacing>;

  // 主题类型
  export interface Theme {
    colors: ColorPalette;
    typography: Typography;
    spacing: SpacingScale;
    shadows: ShadowScale;
    borders: BorderScale;
    breakpoints: Breakpoints;
    animations: Animations;
  }

  // 样式属性类型
  export interface StyleProps {
    // 布局
    display?: ResponsiveValue<'block' | 'inline' | 'flex' | 'grid' | 'none'>;
    position?: ResponsiveValue<'static' | 'relative' | 'absolute' | 'fixed' | 'sticky'>;

    // 间距
    margin?: ResponsiveSpacing;
    marginTop?: ResponsiveSpacing;
    marginRight?: ResponsiveSpacing;
    marginBottom?: ResponsiveSpacing;
    marginLeft?: ResponsiveSpacing;
    marginX?: ResponsiveSpacing;
    marginY?: ResponsiveSpacing;

    padding?: ResponsiveSpacing;
    paddingTop?: ResponsiveSpacing;
    paddingRight?: ResponsiveSpacing;
    paddingBottom?: ResponsiveSpacing;
    paddingLeft?: ResponsiveSpacing;
    paddingX?: ResponsiveSpacing;
    paddingY?: ResponsiveSpacing;

    // 尺寸
    width?: ResponsiveValue<string | number>;
    height?: ResponsiveValue<string | number>;
    minWidth?: ResponsiveValue<string | number>;
    minHeight?: ResponsiveValue<string | number>;
    maxWidth?: ResponsiveValue<string | number>;
    maxHeight?: ResponsiveValue<string | number>;

    // 颜色
    color?: Color;
    backgroundColor?: Color;
    borderColor?: Color;

    // 文字
    fontSize?: ResponsiveValue<keyof Typography['fontSize']>;
    fontWeight?: ResponsiveValue<keyof Typography['fontWeight']>;
    textAlign?: ResponsiveValue<'left' | 'center' | 'right' | 'justify'>;

    // 边框
    border?: ResponsiveValue<string>;
    borderWidth?: ResponsiveValue<number>;
    borderRadius?: ResponsiveValue<number>;

    // 阴影
    boxShadow?: ResponsiveValue<keyof ShadowScale>;
  }

  // 组件Props类型生成器
  export type ComponentProps<T = {}> = T & BaseComponentProps & StyleProps;

  // 多态组件类型
  export type PolymorphicComponentProps<
    C extends React.ElementType,
    Props = {}
  > = Props &
    Omit<React.ComponentPropsWithoutRef<C>, keyof Props> & {
      as?: C;
    };

  // Ref类型
  export type PolymorphicRef<C extends React.ElementType> =
    React.ComponentPropsWithRef<C>['ref'];

  // 完整的多态组件类型
  export type PolymorphicComponentPropsWithRef<
    C extends React.ElementType,
    Props = {}
  > = PolymorphicComponentProps<C, Props> & {
    ref?: PolymorphicRef<C>;
  };
}

// 多态组件实现示例
interface BoxOwnProps {
  variant?: 'solid' | 'outline' | 'ghost';
  size?: Size;
}

type BoxProps<C extends React.ElementType> =
  PolymorphicComponentPropsWithRef<C, BoxOwnProps>;

type BoxComponent = <C extends React.ElementType = 'div'>(
  props: BoxProps<C>
) => React.ReactElement | null;

const Box: BoxComponent = React.forwardRef(
  <C extends React.ElementType = 'div'>(
    { as, variant = 'solid', size = 'md', ...props }: BoxProps<C>,
    ref?: PolymorphicRef<C>
  ) => {
    const Component = as || 'div';

    return (
      <Component
        ref={ref}
        className={clsx(
          'box',
          `box--variant-${variant}`,
          `box--size-${size}`
        )}
        {...props}
      />
    );
  }
);

// 使用示例
const Examples = () => (
  <>
    {/* 默认为div */}
    <Box>Default div box</Box>

    {/* 渲染为button */}
    <Box as="button" onClick={() => console.log('clicked')}>
      Button box
    </Box>

    {/* 渲染为Link组件 */}
    <Box as={Link} to="/home">
      Link box
    </Box>
  </>
);
```

---

## 🎯 API设计原则

### 组件API设计哲学

优秀的组件API应该遵循以下设计原则：

```typescript
// API设计原则
const apiDesignPrinciples = {
  // 1. 一致性 (Consistency)
  consistency: {
    principle: '相似的功能应该有相似的API',
    examples: {
      // ✅ 一致的size属性
      button: '<Button size="md" />',
      input: '<Input size="md" />',
      select: '<Select size="md" />',

      // ✅ 一致的variant属性
      alert: '<Alert variant="success" />',
      badge: '<Badge variant="success" />',
      button: '<Button variant="success" />'
    },
    benefits: [
      '降低学习成本',
      '提高开发效率',
      '减少认知负担',
      '增强可预测性'
    ]
  },

  // 2. 简洁性 (Simplicity)
  simplicity: {
    principle: '提供简单直观的API，复杂功能通过组合实现',
    examples: {
      // ✅ 简单的基础用法
      basic: '<Button>Click me</Button>',

      // ✅ 通过props扩展功能
      extended: '<Button variant="primary" size="lg" loading>Submit</Button>',

      // ✅ 通过组合实现复杂功能
      composed: `
        <Button.Group>
          <Button>First</Button>
          <Button>Second</Button>
          <Button>Third</Button>
        </Button.Group>
      `
    }
  },

  // 3. 可扩展性 (Extensibility)
  extensibility: {
    principle: '组件应该支持扩展而不需要修改源码',
    patterns: {
      // 渲染属性模式
      renderProps: `
        <DataTable
          data={data}
          renderRow={(item, index) => (
            <CustomRow key={item.id} item={item} index={index} />
          )}
        />
      `,

      // 插槽模式
      slots: `
        <Modal>
          <Modal.Header>
            <Modal.Title>标题</Modal.Title>
            <Modal.CloseButton />
          </Modal.Header>
          <Modal.Body>
            内容
          </Modal.Body>
          <Modal.Footer>
            <Button>确定</Button>
          </Modal.Footer>
        </Modal>
      `,

      // 自定义渲染器
      customRenderer: `
        <Form>
          <Form.Field
            name="email"
            render={({ field, meta }) => (
              <CustomEmailInput {...field} error={meta.error} />
            )}
          />
        </Form>
      `
    }
  },

  // 4. 类型安全 (Type Safety)
  typeSafety: {
    principle: '提供完整的TypeScript类型支持',
    implementation: `
      // 严格的类型定义
      interface ButtonProps {
        variant: 'primary' | 'secondary' | 'outline' | 'ghost';
        size: 'sm' | 'md' | 'lg';
        disabled?: boolean;
        loading?: boolean;
        children: React.ReactNode;
        onClick?: (event: React.MouseEvent<HTMLButtonElement>) => void;
      }

      // 泛型支持
      interface SelectProps<T> {
        options: T[];
        value?: T;
        onChange?: (value: T) => void;
        getOptionLabel?: (option: T) => string;
        getOptionValue?: (option: T) => string | number;
      }

      // 条件类型
      type InputProps<T extends 'text' | 'number' | 'email'> = {
        type: T;
        value: T extends 'number' ? number : string;
        onChange: (value: T extends 'number' ? number : string) => void;
      };
    `
  },

  // 5. 可访问性 (Accessibility)
  accessibility: {
    principle: '默认提供良好的可访问性支持',
    features: [
      'ARIA属性自动添加',
      '键盘导航支持',
      '屏幕阅读器友好',
      '焦点管理',
      '语义化HTML'
    ],
    implementation: `
      const Button = ({ children, disabled, ...props }) => {
        return (
          <button
            type="button"
            disabled={disabled}
            aria-disabled={disabled}
            {...props}
          >
            {children}
          </button>
        );
      };

      const Modal = ({ isOpen, onClose, children, ...props }) => {
        const modalRef = useRef<HTMLDivElement>(null);

        // 焦点管理
        useFocusTrap(modalRef, isOpen);

        // ESC键关闭
        useKeyPress('Escape', onClose, isOpen);

        return (
          <div
            ref={modalRef}
            role="dialog"
            aria-modal="true"
            aria-hidden={!isOpen}
            {...props}
          >
            {children}
          </div>
        );
      };
    `
  }
};

// 组件API设计模式
const componentApiPatterns = {
  // 1. 受控与非受控组件
  controlledVsUncontrolled: {
    // 受控组件
    controlled: `
      const [value, setValue] = useState('');

      <Input
        value={value}
        onChange={(e) => setValue(e.target.value)}
      />
    `,

    // 非受控组件
    uncontrolled: `
      <Input
        defaultValue="initial value"
        onChange={(e) => console.log(e.target.value)}
      />
    `,

    // 混合模式
    hybrid: `
      const Input = ({ value, defaultValue, onChange, ...props }) => {
        const [internalValue, setInternalValue] = useState(defaultValue || '');
        const isControlled = value !== undefined;
        const currentValue = isControlled ? value : internalValue;

        const handleChange = (e) => {
          const newValue = e.target.value;
          if (!isControlled) {
            setInternalValue(newValue);
          }
          onChange?.(e);
        };

        return (
          <input
            value={currentValue}
            onChange={handleChange}
            {...props}
          />
        );
      };
    `
  },

  // 2. 复合组件模式
  compoundComponents: {
    definition: '将相关的组件组合在一起，提供更灵活的API',
    implementation: `
      // 主组件
      const Tabs = ({ children, defaultActiveKey, activeKey, onChange }) => {
        const [internalActiveKey, setInternalActiveKey] = useState(defaultActiveKey);
        const isControlled = activeKey !== undefined;
        const currentActiveKey = isControlled ? activeKey : internalActiveKey;

        const handleTabChange = (key) => {
          if (!isControlled) {
            setInternalActiveKey(key);
          }
          onChange?.(key);
        };

        return (
          <TabsContext.Provider value={{
            activeKey: currentActiveKey,
            onTabChange: handleTabChange
          }}>
            <div className="tabs">
              {children}
            </div>
          </TabsContext.Provider>
        );
      };

      // 子组件
      Tabs.List = ({ children }) => (
        <div className="tabs-list" role="tablist">
          {children}
        </div>
      );

      Tabs.Tab = ({ tabKey, children, disabled }) => {
        const { activeKey, onTabChange } = useContext(TabsContext);
        const isActive = activeKey === tabKey;

        return (
          <button
            className={clsx('tab', { 'tab--active': isActive })}
            role="tab"
            aria-selected={isActive}
            disabled={disabled}
            onClick={() => !disabled && onTabChange(tabKey)}
          >
            {children}
          </button>
        );
      };

      Tabs.Panels = ({ children }) => (
        <div className="tabs-panels">
          {children}
        </div>
      );

      Tabs.Panel = ({ tabKey, children }) => {
        const { activeKey } = useContext(TabsContext);
        const isActive = activeKey === tabKey;

        return (
          <div
            className="tab-panel"
            role="tabpanel"
            hidden={!isActive}
          >
            {children}
          </div>
        );
      };

      // 使用示例
      <Tabs defaultActiveKey="tab1">
        <Tabs.List>
          <Tabs.Tab tabKey="tab1">Tab 1</Tabs.Tab>
          <Tabs.Tab tabKey="tab2">Tab 2</Tabs.Tab>
          <Tabs.Tab tabKey="tab3" disabled>Tab 3</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panels>
          <Tabs.Panel tabKey="tab1">Content 1</Tabs.Panel>
          <Tabs.Panel tabKey="tab2">Content 2</Tabs.Panel>
          <Tabs.Panel tabKey="tab3">Content 3</Tabs.Panel>
        </Tabs.Panels>
      </Tabs>
    `
  },

  // 3. 渲染属性模式
  renderProps: {
    definition: '通过函数prop来自定义组件的渲染逻辑',
    implementation: `
      interface DataFetcherProps<T> {
        url: string;
        children: (state: {
          data: T | null;
          loading: boolean;
          error: Error | null;
          refetch: () => void;
        }) => React.ReactNode;
      }

      const DataFetcher = <T,>({ url, children }: DataFetcherProps<T>) => {
        const [data, setData] = useState<T | null>(null);
        const [loading, setLoading] = useState(false);
        const [error, setError] = useState<Error | null>(null);

        const fetchData = useCallback(async () => {
          setLoading(true);
          setError(null);

          try {
            const response = await fetch(url);
            const result = await response.json();
            setData(result);
          } catch (err) {
            setError(err as Error);
          } finally {
            setLoading(false);
          }
        }, [url]);

        useEffect(() => {
          fetchData();
        }, [fetchData]);

        return children({ data, loading, error, refetch: fetchData });
      };

      // 使用示例
      <DataFetcher<User[]> url="/api/users">
        {({ data, loading, error, refetch }) => {
          if (loading) return <Spinner />;
          if (error) return <ErrorMessage error={error} onRetry={refetch} />;
          if (!data) return <EmptyState />;

          return (
            <UserList users={data} onRefresh={refetch} />
          );
        }}
      </DataFetcher>
    `
  }
};
```

---

## 🎨 主题系统设计

### 设计令牌 (Design Tokens)

设计令牌是设计系统的原子单位，定义了所有的设计决策：

```typescript
// 设计令牌定义
const designTokens = {
  // 颜色令牌
  colors: {
    // 基础色彩
    primitive: {
      blue: {
        50: '#eff6ff',
        100: '#dbeafe',
        200: '#bfdbfe',
        300: '#93c5fd',
        400: '#60a5fa',
        500: '#3b82f6',
        600: '#2563eb',
        700: '#1d4ed8',
        800: '#1e40af',
        900: '#1e3a8a',
      },
      gray: {
        50: '#f9fafb',
        100: '#f3f4f6',
        200: '#e5e7eb',
        300: '#d1d5db',
        400: '#9ca3af',
        500: '#6b7280',
        600: '#4b5563',
        700: '#374151',
        800: '#1f2937',
        900: '#111827',
      },
      red: {
        50: '#fef2f2',
        100: '#fee2e2',
        200: '#fecaca',
        300: '#fca5a5',
        400: '#f87171',
        500: '#ef4444',
        600: '#dc2626',
        700: '#b91c1c',
        800: '#991b1b',
        900: '#7f1d1d',
      },
      green: {
        50: '#f0fdf4',
        100: '#dcfce7',
        200: '#bbf7d0',
        300: '#86efac',
        400: '#4ade80',
        500: '#22c55e',
        600: '#16a34a',
        700: '#15803d',
        800: '#166534',
        900: '#14532d',
      },
    },

    // 语义色彩
    semantic: {
      primary: {
        light: 'var(--color-blue-400)',
        main: 'var(--color-blue-500)',
        dark: 'var(--color-blue-600)',
        contrast: '#ffffff',
      },
      secondary: {
        light: 'var(--color-gray-400)',
        main: 'var(--color-gray-500)',
        dark: 'var(--color-gray-600)',
        contrast: '#ffffff',
      },
      success: {
        light: 'var(--color-green-400)',
        main: 'var(--color-green-500)',
        dark: 'var(--color-green-600)',
        contrast: '#ffffff',
      },
      warning: {
        light: '#fbbf24',
        main: '#f59e0b',
        dark: '#d97706',
        contrast: '#ffffff',
      },
      error: {
        light: 'var(--color-red-400)',
        main: 'var(--color-red-500)',
        dark: 'var(--color-red-600)',
        contrast: '#ffffff',
      },
      info: {
        light: 'var(--color-blue-400)',
        main: 'var(--color-blue-500)',
        dark: 'var(--color-blue-600)',
        contrast: '#ffffff',
      },
    },

    // 表面色彩
    surface: {
      background: '#ffffff',
      paper: '#ffffff',
      overlay: 'rgba(0, 0, 0, 0.5)',
      disabled: 'var(--color-gray-100)',
    },

    // 文本色彩
    text: {
      primary: 'var(--color-gray-900)',
      secondary: 'var(--color-gray-600)',
      disabled: 'var(--color-gray-400)',
      hint: 'var(--color-gray-500)',
    },

    // 边框色彩
    border: {
      default: 'var(--color-gray-200)',
      focus: 'var(--color-blue-500)',
      error: 'var(--color-red-500)',
    },
  },

  // 字体令牌
  typography: {
    fontFamily: {
      sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      serif: ['Georgia', 'serif'],
      mono: ['Fira Code', 'Consolas', 'monospace'],
    },

    fontSize: {
      xs: '0.75rem',     // 12px
      sm: '0.875rem',    // 14px
      base: '1rem',      // 16px
      lg: '1.125rem',    // 18px
      xl: '1.25rem',     // 20px
      '2xl': '1.5rem',   // 24px
      '3xl': '1.875rem', // 30px
      '4xl': '2.25rem',  // 36px
      '5xl': '3rem',     // 48px
      '6xl': '3.75rem',  // 60px
    },

    fontWeight: {
      thin: 100,
      extralight: 200,
      light: 300,
      normal: 400,
      medium: 500,
      semibold: 600,
      bold: 700,
      extrabold: 800,
      black: 900,
    },

    lineHeight: {
      none: 1,
      tight: 1.25,
      snug: 1.375,
      normal: 1.5,
      relaxed: 1.625,
      loose: 2,
    },

    letterSpacing: {
      tighter: '-0.05em',
      tight: '-0.025em',
      normal: '0em',
      wide: '0.025em',
      wider: '0.05em',
      widest: '0.1em',
    },
  },

  // 间距令牌
  spacing: {
    0: '0px',
    1: '0.25rem',   // 4px
    2: '0.5rem',    // 8px
    3: '0.75rem',   // 12px
    4: '1rem',      // 16px
    5: '1.25rem',   // 20px
    6: '1.5rem',    // 24px
    7: '1.75rem',   // 28px
    8: '2rem',      // 32px
    9: '2.25rem',   // 36px
    10: '2.5rem',   // 40px
    11: '2.75rem',  // 44px
    12: '3rem',     // 48px
    14: '3.5rem',   // 56px
    16: '4rem',     // 64px
    20: '5rem',     // 80px
    24: '6rem',     // 96px
    28: '7rem',     // 112px
    32: '8rem',     // 128px
    36: '9rem',     // 144px
    40: '10rem',    // 160px
    44: '11rem',    // 176px
    48: '12rem',    // 192px
    52: '13rem',    // 208px
    56: '14rem',    // 224px
    60: '15rem',    // 240px
    64: '16rem',    // 256px
    72: '18rem',    // 288px
    80: '20rem',    // 320px
    96: '24rem',    // 384px
  },

  // 阴影令牌
  shadows: {
    none: 'none',
    sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    base: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
    xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
    '2xl': '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
    inner: 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
  },

  // 边框令牌
  borders: {
    width: {
      0: '0px',
      1: '1px',
      2: '2px',
      4: '4px',
      8: '8px',
    },

    radius: {
      none: '0px',
      sm: '0.125rem',   // 2px
      base: '0.25rem',  // 4px
      md: '0.375rem',   // 6px
      lg: '0.5rem',     // 8px
      xl: '0.75rem',    // 12px
      '2xl': '1rem',    // 16px
      '3xl': '1.5rem',  // 24px
      full: '9999px',
    },
  },

  // 动画令牌
  animations: {
    duration: {
      75: '75ms',
      100: '100ms',
      150: '150ms',
      200: '200ms',
      300: '300ms',
      500: '500ms',
      700: '700ms',
      1000: '1000ms',
    },

    easing: {
      linear: 'linear',
      in: 'cubic-bezier(0.4, 0, 1, 1)',
      out: 'cubic-bezier(0, 0, 0.2, 1)',
      'in-out': 'cubic-bezier(0.4, 0, 0.2, 1)',
    },
  },

  // 断点令牌
  breakpoints: {
    xs: '0px',
    sm: '640px',
    md: '768px',
    lg: '1024px',
    xl: '1280px',
    '2xl': '1536px',
  },

  // Z-index令牌
  zIndex: {
    hide: -1,
    auto: 'auto',
    base: 0,
    docked: 10,
    dropdown: 1000,
    sticky: 1100,
    banner: 1200,
    overlay: 1300,
    modal: 1400,
    popover: 1500,
    skipLink: 1600,
    toast: 1700,
    tooltip: 1800,
  },
};
```

---

## 🎯 面试常考知识点

### 1. 组件库架构设计

**Q: 如何设计一个可扩展的组件库架构？**

**A: 组件库架构设计要点：**

```typescript
// 组件库架构设计原则
const componentLibraryArchitecture = {
  // 1. 分层架构
  layeredArchitecture: {
    tokens: {
      layer: '设计令牌层',
      responsibility: '定义基础设计变量',
      examples: ['颜色', '字体', '间距', '阴影', '动画'],
      implementation: `
        // tokens/colors.ts
        export const colors = {
          primary: { 50: '#eff6ff', 500: '#3b82f6', 900: '#1e3a8a' },
          semantic: { success: '#10b981', error: '#ef4444' }
        };
      `
    },

    primitives: {
      layer: '原始组件层',
      responsibility: '基础UI组件',
      examples: ['Button', 'Input', 'Icon', 'Text'],
      characteristics: ['无业务逻辑', '高度可复用', 'API简单']
    },

    composite: {
      layer: '复合组件层',
      responsibility: '组合型组件',
      examples: ['Form', 'Table', 'Modal', 'Navigation'],
      characteristics: ['包含业务逻辑', '功能完整', '可配置性强']
    },

    patterns: {
      layer: '模式层',
      responsibility: '设计模式和最佳实践',
      examples: ['Layout', 'DataDisplay', 'Feedback', 'Navigation'],
      characteristics: ['解决特定问题', '提供指导', '标准化实践']
    }
  },

  // 2. 模块化设计
  modularDesign: {
    corePackage: {
      name: '@company/ui-core',
      contents: ['基础组件', '主题系统', '工具函数'],
      dependencies: ['react', 'react-dom']
    },

    iconPackage: {
      name: '@company/ui-icons',
      contents: ['SVG图标', '图标组件'],
      dependencies: ['@company/ui-core']
    },

    themePackage: {
      name: '@company/ui-themes',
      contents: ['预设主题', '主题工具'],
      dependencies: ['@company/ui-core']
    },

    utilsPackage: {
      name: '@company/ui-utils',
      contents: ['工具函数', '类型定义', 'Hooks'],
      dependencies: []
    }
  },

  // 3. 版本管理策略
  versioningStrategy: {
    semver: {
      major: '破坏性变更 (API不兼容)',
      minor: '新功能添加 (向后兼容)',
      patch: 'Bug修复 (向后兼容)'
    },

    deprecation: {
      process: [
        '标记为废弃 (console.warn)',
        '提供迁移指南',
        '保持向后兼容',
        '下个主版本移除'
      ],
      example: `
        // 废弃API示例
        const OldButton = (props) => {
          if (process.env.NODE_ENV !== 'production') {
            console.warn(
              'OldButton is deprecated. Use Button instead. ' +
              'See migration guide: https://ui.company.com/migration'
            );
          }
          return <Button {...props} />;
        };
      `
    }
  }
};

// 常见面试问题
const commonInterviewQuestions = {
  q1: {
    question: '如何保证组件库的一致性？',
    answer: {
      designTokens: '使用设计令牌统一设计变量',
      apiConsistency: '保持API设计的一致性',
      documentation: '完善的文档和使用指南',
      testing: '全面的测试覆盖',
      codeReview: '严格的代码审查流程',
      example: `
        // 一致的API设计
        <Button size="md" variant="primary" />
        <Input size="md" variant="outline" />
        <Select size="md" variant="filled" />
      `
    }
  },

  q2: {
    question: '如何处理组件库的主题定制？',
    answer: {
      designTokens: '基于设计令牌的主题系统',
      cssVariables: '使用CSS变量实现动态主题',
      themeProvider: '通过Context提供主题',
      runtimeSwitching: '支持运行时主题切换',
      implementation: `
        // 主题系统实现
        const ThemeProvider = ({ theme, children }) => {
          useEffect(() => {
            // 设置CSS变量
            Object.entries(theme.colors).forEach(([key, value]) => {
              document.documentElement.style.setProperty(
                \`--color-\${key}\`, value
              );
            });
          }, [theme]);

          return (
            <ThemeContext.Provider value={theme}>
              {children}
            </ThemeContext.Provider>
          );
        };
      `
    }
  },

  q3: {
    question: '如何优化组件库的性能？',
    answer: {
      strategies: [
        'Tree Shaking: 支持按需导入',
        'Code Splitting: 组件懒加载',
        'Bundle Optimization: 优化打包体积',
        'Runtime Performance: 减少重渲染',
        'Memory Management: 避免内存泄漏'
      ],
      implementation: `
        // 按需导入支持
        // babel-plugin-import配置
        {
          "libraryName": "@company/ui",
          "libraryDirectory": "es",
          "style": true
        }

        // 使用
        import { Button } from '@company/ui'; // 只导入Button

        // 懒加载组件
        const LazyModal = lazy(() => import('./Modal'));

        // 性能优化
        const OptimizedComponent = memo(({ data }) => {
          const memoizedData = useMemo(() =>
            processData(data), [data]
          );

          return <ExpensiveComponent data={memoizedData} />;
        });
      `
    }
  },

  q4: {
    question: '如何测试组件库？',
    answer: {
      testingLevels: {
        unit: '单个组件的功能测试',
        integration: '组件间的集成测试',
        visual: '视觉回归测试',
        accessibility: '可访问性测试',
        performance: '性能测试'
      },
      tools: [
        'Jest + React Testing Library: 单元测试',
        'Storybook: 组件文档和测试',
        'Chromatic: 视觉回归测试',
        'axe-core: 可访问性测试',
        'Lighthouse: 性能测试'
      ],
      example: `
        // 组件测试示例
        describe('Button Component', () => {
          it('renders correctly', () => {
            render(<Button>Click me</Button>);
            expect(screen.getByRole('button')).toBeInTheDocument();
          });

          it('handles click events', () => {
            const handleClick = jest.fn();
            render(<Button onClick={handleClick}>Click me</Button>);

            fireEvent.click(screen.getByRole('button'));
            expect(handleClick).toHaveBeenCalledTimes(1);
          });

          it('supports different variants', () => {
            const { rerender } = render(<Button variant="primary">Button</Button>);
            expect(screen.getByRole('button')).toHaveClass('button--primary');

            rerender(<Button variant="secondary">Button</Button>);
            expect(screen.getByRole('button')).toHaveClass('button--secondary');
          });

          it('is accessible', async () => {
            const { container } = render(<Button>Accessible Button</Button>);
            const results = await axe(container);
            expect(results).toHaveNoViolations();
          });
        });
      `
    }
  }
};
```

### 2. 跨框架组件库对比

**Q: 不同前端框架的组件库有什么区别？**

**A: 主流框架组件库对比：**

```typescript
// 跨框架组件库对比
const crossFrameworkComparison = {
  // React生态
  react: {
    popularLibraries: [
      'Ant Design - 企业级UI设计语言',
      'Material-UI - Google Material Design',
      'Chakra UI - 简单、模块化、可访问',
      'Mantine - 功能丰富的组件库',
      'React Bootstrap - Bootstrap的React实现'
    ],

    characteristics: {
      componentModel: 'Function Components + Hooks',
      stateManagement: 'useState, useReducer, Context',
      styling: 'CSS-in-JS, Styled Components, CSS Modules',
      typeScript: '优秀的TypeScript支持',
      ecosystem: '最丰富的生态系统'
    },

    example: `
      // React组件示例
      const Button = ({ variant, size, children, ...props }) => {
        const classes = clsx(
          'btn',
          \`btn--\${variant}\`,
          \`btn--\${size}\`
        );

        return (
          <button className={classes} {...props}>
            {children}
          </button>
        );
      };

      // 使用Hooks的复杂组件
      const DataTable = ({ data, columns }) => {
        const [sortConfig, setSortConfig] = useState(null);
        const [filterConfig, setFilterConfig] = useState({});

        const sortedData = useMemo(() => {
          if (!sortConfig) return data;

          return [...data].sort((a, b) => {
            if (a[sortConfig.key] < b[sortConfig.key]) {
              return sortConfig.direction === 'asc' ? -1 : 1;
            }
            if (a[sortConfig.key] > b[sortConfig.key]) {
              return sortConfig.direction === 'asc' ? 1 : -1;
            }
            return 0;
          });
        }, [data, sortConfig]);

        return (
          <table>
            <thead>
              {columns.map(column => (
                <th key={column.key} onClick={() => handleSort(column.key)}>
                  {column.title}
                </th>
              ))}
            </thead>
            <tbody>
              {sortedData.map(row => (
                <tr key={row.id}>
                  {columns.map(column => (
                    <td key={column.key}>{row[column.key]}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        );
      };
    `
  },

  // Vue生态
  vue: {
    popularLibraries: [
      'Element Plus - 基于Vue 3的桌面端组件库',
      'Ant Design Vue - Ant Design的Vue实现',
      'Vuetify - Material Design组件框架',
      'Quasar - 跨平台Vue组件库',
      'Naive UI - 较为完整的Vue 3组件库'
    ],

    characteristics: {
      componentModel: 'Composition API + Options API',
      stateManagement: 'ref, reactive, Pinia, Vuex',
      styling: 'Scoped CSS, CSS Modules, CSS-in-JS',
      typeScript: '良好的TypeScript支持',
      ecosystem: '快速发展的生态系统'
    },

    example: `
      <!-- Vue组件示例 -->
      <template>
        <button
          :class="buttonClasses"
          :disabled="disabled"
          @click="handleClick"
        >
          <slot />
        </button>
      </template>

      <script setup lang="ts">
      interface Props {
        variant?: 'primary' | 'secondary';
        size?: 'sm' | 'md' | 'lg';
        disabled?: boolean;
      }

      const props = withDefaults(defineProps<Props>(), {
        variant: 'primary',
        size: 'md',
        disabled: false
      });

      const emit = defineEmits<{
        click: [event: MouseEvent];
      }>();

      const buttonClasses = computed(() => [
        'btn',
        \`btn--\${props.variant}\`,
        \`btn--\${props.size}\`,
        { 'btn--disabled': props.disabled }
      ]);

      const handleClick = (event: MouseEvent) => {
        if (!props.disabled) {
          emit('click', event);
        }
      };
      </script>

      <style scoped>
      .btn {
        @apply px-4 py-2 rounded font-medium transition-colors;
      }

      .btn--primary {
        @apply bg-blue-500 text-white hover:bg-blue-600;
      }

      .btn--secondary {
        @apply bg-gray-200 text-gray-800 hover:bg-gray-300;
      }
      </style>
    `
  },

  // Angular生态
  angular: {
    popularLibraries: [
      'Angular Material - Google官方Material Design',
      'NG-ZORRO - Ant Design的Angular实现',
      'PrimeNG - 丰富的UI组件集合',
      'Clarity - VMware的设计系统',
      'Ionic - 移动端UI组件库'
    ],

    characteristics: {
      componentModel: 'Class Components + Decorators',
      stateManagement: 'Services, RxJS, NgRx',
      styling: 'Component Styles, Global Styles, CSS Modules',
      typeScript: '原生TypeScript支持',
      ecosystem: '企业级生态系统'
    },

    example: `
      // Angular组件示例
      @Component({
        selector: 'app-button',
        template: \`
          <button
            [class]="buttonClasses"
            [disabled]="disabled"
            (click)="handleClick($event)"
          >
            <ng-content></ng-content>
          </button>
        \`,
        styleUrls: ['./button.component.scss']
      })
      export class ButtonComponent {
        @Input() variant: 'primary' | 'secondary' = 'primary';
        @Input() size: 'sm' | 'md' | 'lg' = 'md';
        @Input() disabled: boolean = false;

        @Output() clicked = new EventEmitter<MouseEvent>();

        get buttonClasses(): string {
          return [
            'btn',
            \`btn--\${this.variant}\`,
            \`btn--\${this.size}\`,
            this.disabled ? 'btn--disabled' : ''
          ].filter(Boolean).join(' ');
        }

        handleClick(event: MouseEvent): void {
          if (!this.disabled) {
            this.clicked.emit(event);
          }
        }
      }

      // 复杂组件示例
      @Component({
        selector: 'app-data-table',
        template: \`
          <table class="data-table">
            <thead>
              <tr>
                <th
                  *ngFor="let column of columns"
                  (click)="sort(column.key)"
                  [class.sortable]="column.sortable"
                >
                  {{ column.title }}
                  <span *ngIf="sortConfig?.key === column.key">
                    {{ sortConfig.direction === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr *ngFor="let row of sortedData">
                <td *ngFor="let column of columns">
                  {{ row[column.key] }}
                </td>
              </tr>
            </tbody>
          </table>
        \`
      })
      export class DataTableComponent {
        @Input() data: any[] = [];
        @Input() columns: TableColumn[] = [];

        sortConfig: { key: string; direction: 'asc' | 'desc' } | null = null;

        get sortedData(): any[] {
          if (!this.sortConfig) return this.data;

          return [...this.data].sort((a, b) => {
            const aVal = a[this.sortConfig!.key];
            const bVal = b[this.sortConfig!.key];

            if (aVal < bVal) {
              return this.sortConfig!.direction === 'asc' ? -1 : 1;
            }
            if (aVal > bVal) {
              return this.sortConfig!.direction === 'asc' ? 1 : -1;
            }
            return 0;
          });
        }

        sort(key: string): void {
          if (this.sortConfig?.key === key) {
            this.sortConfig.direction =
              this.sortConfig.direction === 'asc' ? 'desc' : 'asc';
          } else {
            this.sortConfig = { key, direction: 'asc' };
          }
        }
      }
    `
  },

  // 框架对比总结
  comparison: {
    learningCurve: {
      react: 'Medium - Hooks概念需要理解',
      vue: 'Easy - 渐进式学习',
      angular: 'Hard - 完整的框架概念'
    },

    performance: {
      react: 'Good - 虚拟DOM + Fiber',
      vue: 'Excellent - 响应式系统 + 编译优化',
      angular: 'Good - Zone.js + AOT编译'
    },

    ecosystem: {
      react: 'Largest - 最丰富的第三方库',
      vue: 'Growing - 快速发展的生态',
      angular: 'Mature - 企业级完整解决方案'
    },

    typescript: {
      react: 'Excellent - 社区驱动的类型支持',
      vue: 'Good - 官方TypeScript支持',
      angular: 'Native - 原生TypeScript框架'
    }
  }
};
```

---

## 📚 实战练习

### 练习1：设计基础组件库

**任务**: 为Mall-Frontend设计一套基础组件库，包含Button、Input、Card等核心组件。

**要求**:
- 使用TypeScript编写
- 支持主题定制
- 提供完整的API文档
- 包含单元测试

### 练习2：实现复合组件

**任务**: 实现一个复合的Table组件，支持排序、筛选、分页等功能。

**要求**:
- 使用复合组件模式
- 支持自定义渲染
- 提供丰富的配置选项
- 优化性能表现

### 练习3：构建主题系统

**任务**: 构建一个完整的主题系统，支持多主题切换和自定义主题。

**要求**:
- 基于设计令牌
- 支持运行时切换
- 提供主题编辑器
- 兼容SSR

---

## 📚 本章总结

通过本章学习，我们全面掌握了组件库设计与开发的核心技术：

### 🎯 核心收获

1. **设计系统理论精通** 🎨
   - 掌握了设计系统的核心概念和价值
   - 理解了原子设计理论和实践方法
   - 学会了设计令牌的定义和使用

2. **组件库架构设计** 🏗️
   - 掌握了模块化架构设计原则
   - 学会了组件分层和依赖管理
   - 理解了版本管理和发布策略

3. **API设计能力** 🎯
   - 掌握了组件API设计原则
   - 学会了多种组件设计模式
   - 理解了类型系统和多态组件

4. **主题系统构建** 🎨
   - 掌握了设计令牌的定义和管理
   - 学会了主题系统的实现方法
   - 理解了运行时主题切换机制

5. **工程化实践** 🚀
   - 学会了组件库的构建和发布
   - 掌握了测试策略和质量保证
   - 理解了文档系统的重要性

### 🚀 技术进阶

- **下一步学习**: 微前端架构实践
- **实践建议**: 在项目中构建自己的组件库
- **深入方向**: 跨框架组件库和Web Components

组件库是现代前端开发的基础设施，掌握组件库的设计和开发是前端架构师的必备技能！ 🎉

---

*下一章我们将学习《微前端架构实践》，探索大型应用的架构拆分和治理！* 🚀
```
```
```
