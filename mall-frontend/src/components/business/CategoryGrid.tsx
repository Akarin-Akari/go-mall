'use client';

import React from 'react';
import { Row, Col, Card, Typography, Space } from 'antd';
import { Category } from '@/types';

const { Text, Title } = Typography;

interface CategoryGridProps {
  categories: Category[];
  onCategoryClick?: (categoryId: number) => void;
  loading?: boolean;
  columns?: {
    xs?: number;
    sm?: number;
    md?: number;
    lg?: number;
    xl?: number;
  };
}

const CategoryGrid: React.FC<CategoryGridProps> = ({
  categories,
  onCategoryClick,
  loading = false,
  columns = { xs: 2, sm: 3, md: 4, lg: 6, xl: 6 },
}) => {
  // 分类图标映射
  const getCategoryIcon = (categoryName: string) => {
    const iconMap: { [key: string]: string } = {
      电子产品: '📱',
      服装鞋帽: '👕',
      家居用品: '🏠',
      美妆护肤: '💄',
      运动户外: '⚽',
      图书音像: '📚',
      食品饮料: '🍎',
      母婴用品: '👶',
      汽车用品: '🚗',
      办公用品: '📝',
      数码配件: '🔌',
      家用电器: '🔌',
      手机通讯: '📞',
      电脑办公: '💻',
      家装建材: '🔨',
      珠宝首饰: '💎',
      钟表眼镜: '⌚',
      玩具乐器: '🎮',
      宠物用品: '🐕',
      医疗保健: '💊',
    };

    return iconMap[categoryName] || '📦';
  };

  // 处理分类点击
  const handleCategoryClick = (category: Category) => {
    onCategoryClick?.(category.id);
  };

  return (
    <Row gutter={[16, 16]}>
      {categories.map(category => (
        <Col
          key={category.id}
          xs={columns.xs ? 24 / columns.xs : 12}
          sm={columns.sm ? 24 / columns.sm : 8}
          md={columns.md ? 24 / columns.md : 6}
          lg={columns.lg ? 24 / columns.lg : 4}
          xl={columns.xl ? 24 / columns.xl : 4}
        >
          <Card
            hoverable
            loading={loading}
            style={{
              textAlign: 'center',
              borderRadius: 8,
              border: '1px solid #f0f0f0',
              transition: 'all 0.3s ease',
              cursor: 'pointer',
            }}
            bodyStyle={{
              padding: '20px 16px',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: 120,
            }}
            onClick={() => handleCategoryClick(category)}
          >
            {/* 分类图标 */}
            <div
              style={{
                fontSize: 32,
                marginBottom: 8,
                transition: 'transform 0.3s ease',
              }}
              className='category-icon'
            >
              {category.icon || getCategoryIcon(category.name)}
            </div>

            {/* 分类名称 */}
            <Title
              level={5}
              style={{
                margin: 0,
                marginBottom: 4,
                fontSize: 14,
                fontWeight: 500,
              }}
            >
              {category.name}
            </Title>

            {/* 商品数量 */}
            {category.product_count !== undefined && (
              <Text
                type='secondary'
                style={{
                  fontSize: 12,
                  opacity: 0.8,
                }}
              >
                {category.product_count} 件商品
              </Text>
            )}

            {/* 分类描述 */}
            {category.description && (
              <Text
                type='secondary'
                style={{
                  fontSize: 11,
                  marginTop: 4,
                  textAlign: 'center',
                  lineHeight: 1.2,
                  display: '-webkit-box',
                  WebkitLineClamp: 2,
                  WebkitBoxOrient: 'vertical',
                  overflow: 'hidden',
                }}
              >
                {category.description}
              </Text>
            )}
          </Card>
        </Col>
      ))}

      <style jsx>{`
        :global(.ant-card:hover .category-icon) {
          transform: scale(1.1);
        }

        :global(.ant-card:hover) {
          border-color: #1890ff;
          box-shadow: 0 4px 12px rgba(24, 144, 255, 0.15);
        }
      `}</style>
    </Row>
  );
};

export default CategoryGrid;
