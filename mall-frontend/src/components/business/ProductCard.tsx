'use client';

import React, { useState, useCallback } from 'react';
import {
  Card,
  Typography,
  Button,
  Space,
  Tag,
  Rate,
  Tooltip,
  Image,
  Badge,
  message,
} from 'antd';
import {
  ShoppingCartOutlined,
  HeartOutlined,
  HeartFilled,
  EyeOutlined,
  ShareAltOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import { Product } from '@/types';
import { ROUTES } from '@/constants';
import { formatPrice, formatNumber } from '@/utils';

const { Text, Title } = Typography;
const { Meta } = Card;

interface ProductCardProps {
  product: Product;
  loading?: boolean;
  showActions?: boolean;
  showBadge?: boolean | string;
  badgeColor?: string;
  size?: 'small' | 'default' | 'large';
  onAddToCart?: (product: Product) => void;
  onToggleFavorite?: (product: Product) => void;
  onShare?: (product: Product) => void;
  onViewDetail?: (productId: number) => void;
  className?: string;
  style?: React.CSSProperties;
}

const ProductCard: React.FC<ProductCardProps> = ({
  product,
  loading = false,
  showActions = true,
  showBadge = true,
  badgeColor,
  size = 'default',
  onAddToCart,
  onToggleFavorite,
  onShare,
  onViewDetail,
  className,
  style,
}) => {
  const [imageLoading, setImageLoading] = useState(true);
  const [imageError, setImageError] = useState(false);
  const [isFavorited, setIsFavorited] = useState(false);
  const [addingToCart, setAddingToCart] = useState(false);

  const router = useRouter();

  // 处理图片加载
  const handleImageLoad = useCallback(() => {
    setImageLoading(false);
  }, []);

  const handleImageError = useCallback(() => {
    setImageLoading(false);
    setImageError(true);
  }, []);

  // 处理点击商品卡片
  const handleCardClick = useCallback(() => {
    if (onViewDetail) {
      onViewDetail(product.id);
    } else {
      router.push(ROUTES.PRODUCT_DETAIL(product.id));
    }
  }, [router, product.id, onViewDetail]);

  // 处理添加到购物车
  const handleAddToCart = useCallback(
    async (e: React.MouseEvent) => {
      e.stopPropagation(); // 阻止事件冒泡

      if (!onAddToCart) return;

      try {
        setAddingToCart(true);
        await onAddToCart(product);
        message.success('已添加到购物车');
      } catch (error) {
        message.error('添加失败，请重试');
      } finally {
        setAddingToCart(false);
      }
    },
    [onAddToCart, product]
  );

  // 处理收藏切换
  const handleToggleFavorite = useCallback(
    async (e: React.MouseEvent) => {
      e.stopPropagation();

      try {
        setIsFavorited(!isFavorited);
        if (onToggleFavorite) {
          await onToggleFavorite(product);
        }
        message.success(isFavorited ? '已取消收藏' : '已添加到收藏');
      } catch (error) {
        setIsFavorited(isFavorited); // 回滚状态
        message.error('操作失败，请重试');
      }
    },
    [isFavorited, onToggleFavorite, product]
  );

  // 处理分享
  const handleShare = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();

      if (onShare) {
        onShare(product);
      } else {
        // 默认分享逻辑
        if (navigator.share) {
          navigator.share({
            title: product.name,
            text: product.description,
            url: window.location.origin + ROUTES.PRODUCT_DETAIL(product.id),
          });
        } else {
          // 复制链接到剪贴板
          navigator.clipboard.writeText(
            window.location.origin + ROUTES.PRODUCT_DETAIL(product.id)
          );
          message.success('链接已复制到剪贴板');
        }
      }
    },
    [onShare, product]
  );

  // 获取商品状态标签
  const getStatusBadge = () => {
    if (!showBadge) return null;

    // 如果showBadge是字符串，直接使用自定义标签
    if (typeof showBadge === 'string') {
      return (
        <Badge.Ribbon
          key='custom'
          text={showBadge}
          color={badgeColor || 'blue'}
        >
          <div />
        </Badge.Ribbon>
      );
    }

    const badges = [];

    // 热销标签
    if (product.sales_count && product.sales_count > 1000) {
      badges.push(
        <Badge.Ribbon key='hot' text='热销' color='red'>
          <div />
        </Badge.Ribbon>
      );
    }

    // 新品标签
    const isNew =
      product.created_at &&
      new Date(product.created_at).getTime() >
        Date.now() - 7 * 24 * 60 * 60 * 1000;
    if (isNew) {
      badges.push(
        <Badge.Ribbon key='new' text='新品' color='green'>
          <div />
        </Badge.Ribbon>
      );
    }

    // 限时优惠标签
    if (
      product.discount_price &&
      parseFloat(product.discount_price) < parseFloat(product.price)
    ) {
      badges.push(
        <Badge.Ribbon key='discount' text='限时优惠' color='orange'>
          <div />
        </Badge.Ribbon>
      );
    }

    return badges[0]; // 只显示第一个标签
  };

  // 获取卡片尺寸配置
  const getSizeConfig = () => {
    switch (size) {
      case 'small':
        return {
          cardStyle: { width: 200 },
          imageHeight: 150,
          titleLevel: 5 as const,
          showDescription: false,
        };
      case 'large':
        return {
          cardStyle: { width: 320 },
          imageHeight: 240,
          titleLevel: 4 as const,
          showDescription: true,
        };
      default:
        return {
          cardStyle: { width: 260 },
          imageHeight: 200,
          titleLevel: 5 as const,
          showDescription: true,
        };
    }
  };

  const sizeConfig = getSizeConfig();

  // 渲染操作按钮
  const renderActions = () => {
    if (!showActions) return [];

    return [
      <Tooltip key='view' title='查看详情'>
        <Button type='text' icon={<EyeOutlined />} onClick={handleCardClick} />
      </Tooltip>,
      <Tooltip key='favorite' title={isFavorited ? '取消收藏' : '添加收藏'}>
        <Button
          type='text'
          icon={
            isFavorited ? (
              <HeartFilled style={{ color: '#ff4d4f' }} />
            ) : (
              <HeartOutlined />
            )
          }
          onClick={handleToggleFavorite}
        />
      </Tooltip>,
      <Tooltip key='share' title='分享商品'>
        <Button type='text' icon={<ShareAltOutlined />} onClick={handleShare} />
      </Tooltip>,
    ];
  };

  return (
    <div className={`product-card ${className || ''}`} style={style}>
      {getStatusBadge()}
      <Card
        loading={loading}
        hoverable
        style={{
          ...sizeConfig.cardStyle,
          cursor: 'pointer',
          transition: 'all 0.3s ease',
        }}
        styles={{
          body: { padding: '12px' },
        }}
        cover={
          <div
            style={{
              position: 'relative',
              height: sizeConfig.imageHeight,
              overflow: 'hidden',
            }}
          >
            <Image
              alt={product.name}
              src={
                imageError
                  ? '/images/product-placeholder.svg'
                  : product.images?.[0] || '/images/product-placeholder.svg'
              }
              style={{
                width: '100%',
                height: '100%',
                objectFit: 'cover',
                transition: 'transform 0.3s ease',
              }}
              preview={false}
              loading='lazy'
              onLoad={handleImageLoad}
              onError={handleImageError}
              placeholder={
                <div
                  style={{
                    width: '100%',
                    height: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    backgroundColor: '#f5f5f5',
                  }}
                >
                  加载中...
                </div>
              }
            />

            {/* 悬浮操作按钮 */}
            <div
              style={{
                position: 'absolute',
                top: 8,
                right: 8,
                opacity: 0,
                transition: 'opacity 0.3s ease',
              }}
              className='product-card-actions'
            >
              <Space direction='vertical' size='small'>
                <Button
                  type='primary'
                  shape='circle'
                  size='small'
                  icon={isFavorited ? <HeartFilled /> : <HeartOutlined />}
                  onClick={handleToggleFavorite}
                  style={{
                    backgroundColor: isFavorited
                      ? '#ff4d4f'
                      : 'rgba(0,0,0,0.6)',
                    borderColor: 'transparent',
                  }}
                />
                <Button
                  type='primary'
                  shape='circle'
                  size='small'
                  icon={<ShareAltOutlined />}
                  onClick={handleShare}
                  style={{
                    backgroundColor: 'rgba(0,0,0,0.6)',
                    borderColor: 'transparent',
                  }}
                />
              </Space>
            </div>
          </div>
        }
        actions={renderActions()}
        onClick={handleCardClick}
      >
        <Meta
          title={
            <Tooltip title={product.name}>
              <Title
                level={sizeConfig.titleLevel}
                ellipsis={{ rows: 2 }}
                style={{ margin: 0, minHeight: size === 'small' ? 32 : 44 }}
              >
                {product.name}
              </Title>
            </Tooltip>
          }
          description={
            sizeConfig.showDescription && product.description ? (
              <Text
                type='secondary'
                ellipsis
                style={{ fontSize: 12, minHeight: 32 }}
              >
                {product.description}
              </Text>
            ) : null
          }
        />

        {/* 评分和销量 */}
        <div style={{ margin: '8px 0' }}>
          <Space split={<span style={{ color: '#d9d9d9' }}>|</span>}>
            <Space size='small'>
              <Rate
                disabled
                allowHalf
                value={product.rating || 0}
                style={{ fontSize: 12 }}
              />
              <Text type='secondary' style={{ fontSize: 12 }}>
                {product.rating?.toFixed(1) || '暂无评分'}
              </Text>
            </Space>
            {product.sales_count && (
              <Text type='secondary' style={{ fontSize: 12 }}>
                已售 {formatNumber(product.sales_count)}
              </Text>
            )}
          </Space>
        </div>

        {/* 价格区域 */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginTop: 8,
          }}
        >
          <Space direction='vertical' size={0}>
            <Space align='baseline'>
              <Text strong style={{ color: '#ff4d4f', fontSize: 18 }}>
                ¥{formatPrice(product.discount_price || product.price)}
              </Text>
              {product.discount_price &&
                parseFloat(product.discount_price) <
                  parseFloat(product.price) && (
                  <Text delete type='secondary' style={{ fontSize: 12 }}>
                    ¥{formatPrice(product.price)}
                  </Text>
                )}
            </Space>
            {product.discount_price &&
              parseFloat(product.discount_price) <
                parseFloat(product.price) && (
                <Tag color='red'>
                  省¥
                  {formatPrice(
                    (
                      parseFloat(product.price) -
                      parseFloat(product.discount_price)
                    ).toString()
                  )}
                </Tag>
              )}
          </Space>

          {/* 添加到购物车按钮 */}
          <Button
            type='primary'
            icon={<ShoppingCartOutlined />}
            loading={addingToCart}
            onClick={handleAddToCart}
            size={size === 'small' ? 'small' : 'middle'}
            style={{
              borderRadius: '50%',
              width: size === 'small' ? 32 : 40,
              height: size === 'small' ? 32 : 40,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          />
        </div>
      </Card>

      <style jsx>{`
        .product-card:hover .product-card-actions {
          opacity: 1 !important;
        }

        .product-card:hover img {
          transform: scale(1.05);
        }
      `}</style>
    </div>
  );
};

export default ProductCard;
