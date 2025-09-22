'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import {
  Image,
  Card,
  Button,
  Space,
  Modal,
  Spin,
  message,
  Tooltip,
} from 'antd';
import {
  LeftOutlined,
  RightOutlined,
  ExpandOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
  RotateLeftOutlined,
  RotateRightOutlined,
  DownloadOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
} from '@ant-design/icons';

interface ProductImageGalleryProps {
  images: string[];
  productName: string;
  className?: string;
  style?: React.CSSProperties;
}

const ProductImageGallery: React.FC<ProductImageGalleryProps> = ({
  images,
  productName,
  className,
  style,
}) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [previewVisible, setPreviewVisible] = useState(false);
  const [previewIndex, setPreviewIndex] = useState(0);
  const [imageLoading, setImageLoading] = useState<Record<number, boolean>>({});
  const [imageError, setImageError] = useState<Record<number, boolean>>({});
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [zoom, setZoom] = useState(1);
  const [rotation, setRotation] = useState(0);

  const mainImageRef = useRef<HTMLDivElement>(null);
  const thumbnailsRef = useRef<HTMLDivElement>(null);

  // 处理图片加载
  const handleImageLoad = useCallback((index: number) => {
    setImageLoading(prev => ({ ...prev, [index]: false }));
  }, []);

  const handleImageError = useCallback((index: number) => {
    setImageLoading(prev => ({ ...prev, [index]: false }));
    setImageError(prev => ({ ...prev, [index]: true }));
  }, []);

  // 切换主图
  const handleImageChange = useCallback(
    (index: number) => {
      if (index >= 0 && index < images.length) {
        setCurrentIndex(index);
        setImageLoading(prev => ({ ...prev, [index]: true }));
      }
    },
    [images.length]
  );

  // 上一张图片
  const handlePrevImage = useCallback(() => {
    const newIndex = currentIndex > 0 ? currentIndex - 1 : images.length - 1;
    handleImageChange(newIndex);
  }, [currentIndex, images.length, handleImageChange]);

  // 下一张图片
  const handleNextImage = useCallback(() => {
    const newIndex = currentIndex < images.length - 1 ? currentIndex + 1 : 0;
    handleImageChange(newIndex);
  }, [currentIndex, images.length, handleImageChange]);

  // 打开预览
  const handlePreview = useCallback(
    (index?: number) => {
      setPreviewIndex(index ?? currentIndex);
      setPreviewVisible(true);
      setZoom(1);
      setRotation(0);
    },
    [currentIndex]
  );

  // 关闭预览
  const handlePreviewClose = useCallback(() => {
    setPreviewVisible(false);
    setZoom(1);
    setRotation(0);
  }, []);

  // 预览中切换图片
  const handlePreviewChange = useCallback(
    (index: number) => {
      if (index >= 0 && index < images.length) {
        setPreviewIndex(index);
        setZoom(1);
        setRotation(0);
      }
    },
    [images.length]
  );

  // 预览中的上一张
  const handlePreviewPrev = useCallback(() => {
    const newIndex = previewIndex > 0 ? previewIndex - 1 : images.length - 1;
    handlePreviewChange(newIndex);
  }, [previewIndex, images.length, handlePreviewChange]);

  // 预览中的下一张
  const handlePreviewNext = useCallback(() => {
    const newIndex = previewIndex < images.length - 1 ? previewIndex + 1 : 0;
    handlePreviewChange(newIndex);
  }, [previewIndex, images.length, handlePreviewChange]);

  // 缩放控制
  const handleZoomIn = useCallback(() => {
    setZoom(prev => Math.min(prev + 0.5, 3));
  }, []);

  const handleZoomOut = useCallback(() => {
    setZoom(prev => Math.max(prev - 0.5, 0.5));
  }, []);

  // 旋转控制
  const handleRotateLeft = useCallback(() => {
    setRotation(prev => prev - 90);
  }, []);

  const handleRotateRight = useCallback(() => {
    setRotation(prev => prev + 90);
  }, []);

  // 下载图片
  const handleDownload = useCallback(async () => {
    try {
      const imageUrl = images[previewIndex];
      const response = await fetch(imageUrl);
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${productName}-${previewIndex + 1}.jpg`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      message.success('图片下载成功');
    } catch (error) {
      message.error('图片下载失败');
    }
  }, [images, previewIndex, productName]);

  // 全屏控制
  const handleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen();
      setIsFullscreen(true);
    } else {
      document.exitFullscreen();
      setIsFullscreen(false);
    }
  }, []);

  // 监听全屏状态变化
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
    };
  }, []);

  // 键盘事件处理
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (previewVisible) {
        switch (e.key) {
          case 'ArrowLeft':
            e.preventDefault();
            handlePreviewPrev();
            break;
          case 'ArrowRight':
            e.preventDefault();
            handlePreviewNext();
            break;
          case 'Escape':
            e.preventDefault();
            handlePreviewClose();
            break;
          case '+':
          case '=':
            e.preventDefault();
            handleZoomIn();
            break;
          case '-':
            e.preventDefault();
            handleZoomOut();
            break;
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [
    previewVisible,
    handlePreviewPrev,
    handlePreviewNext,
    handlePreviewClose,
    handleZoomIn,
    handleZoomOut,
  ]);

  if (!images || images.length === 0) {
    return (
      <Card className={className} style={style}>
        <div
          style={{
            height: 400,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: '#f5f5f5',
          }}
        >
          <span style={{ color: '#999' }}>暂无图片</span>
        </div>
      </Card>
    );
  }

  return (
    <div className={className} style={style}>
      {/* 主图区域 */}
      <Card
        style={{ marginBottom: 16 }}
        bodyStyle={{ padding: 0, position: 'relative' }}
      >
        <div
          ref={mainImageRef}
          style={{
            position: 'relative',
            height: 400,
            overflow: 'hidden',
            backgroundColor: '#f5f5f5',
            cursor: 'zoom-in',
          }}
          onClick={() => handlePreview()}
        >
          <Image
            src={
              imageError[currentIndex]
                ? '/images/product-placeholder.svg'
                : images[currentIndex]
            }
            alt={`${productName} - 图片 ${currentIndex + 1}`}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'contain',
              transition: 'transform 0.3s ease',
            }}
            preview={false}
            loading='lazy'
            onLoad={() => handleImageLoad(currentIndex)}
            onError={() => handleImageError(currentIndex)}
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
                <Spin size='large' />
              </div>
            }
          />

          {/* 图片导航按钮 */}
          {images.length > 1 && (
            <>
              <Button
                type='text'
                icon={<LeftOutlined />}
                onClick={e => {
                  e.stopPropagation();
                  handlePrevImage();
                }}
                style={{
                  position: 'absolute',
                  left: 16,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  color: 'white',
                  border: 'none',
                  borderRadius: '50%',
                  width: 40,
                  height: 40,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              />
              <Button
                type='text'
                icon={<RightOutlined />}
                onClick={e => {
                  e.stopPropagation();
                  handleNextImage();
                }}
                style={{
                  position: 'absolute',
                  right: 16,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  color: 'white',
                  border: 'none',
                  borderRadius: '50%',
                  width: 40,
                  height: 40,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              />
            </>
          )}

          {/* 放大按钮 */}
          <Tooltip title='点击查看大图'>
            <Button
              type='text'
              icon={<ExpandOutlined />}
              onClick={e => {
                e.stopPropagation();
                handlePreview();
              }}
              style={{
                position: 'absolute',
                top: 16,
                right: 16,
                backgroundColor: 'rgba(0, 0, 0, 0.5)',
                color: 'white',
                border: 'none',
                borderRadius: '50%',
                width: 36,
                height: 36,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            />
          </Tooltip>

          {/* 图片指示器 */}
          {images.length > 1 && (
            <div
              style={{
                position: 'absolute',
                bottom: 16,
                left: '50%',
                transform: 'translateX(-50%)',
                display: 'flex',
                gap: 8,
              }}
            >
              {images.map((_, index) => (
                <div
                  key={index}
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: '50%',
                    backgroundColor:
                      index === currentIndex
                        ? 'white'
                        : 'rgba(255, 255, 255, 0.5)',
                    cursor: 'pointer',
                    transition: 'all 0.3s ease',
                  }}
                  onClick={e => {
                    e.stopPropagation();
                    handleImageChange(index);
                  }}
                />
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* 缩略图区域 */}
      {images.length > 1 && (
        <div
          ref={thumbnailsRef}
          style={{
            display: 'flex',
            gap: 8,
            overflowX: 'auto',
            paddingBottom: 8,
            scrollbarWidth: 'thin',
          }}
        >
          {images.map((image, index) => (
            <div
              key={index}
              style={{
                minWidth: 80,
                height: 80,
                border:
                  index === currentIndex
                    ? '2px solid #1890ff'
                    : '2px solid transparent',
                borderRadius: 8,
                overflow: 'hidden',
                cursor: 'pointer',
                transition: 'all 0.3s ease',
                backgroundColor: '#f5f5f5',
              }}
              onClick={() => handleImageChange(index)}
            >
              <Image
                src={
                  imageError[index] ? '/images/product-placeholder.svg' : image
                }
                alt={`${productName} - 缩略图 ${index + 1}`}
                style={{
                  width: '100%',
                  height: '100%',
                  objectFit: 'cover',
                }}
                preview={false}
                loading='lazy'
                onLoad={() => handleImageLoad(index)}
                onError={() => handleImageError(index)}
              />
            </div>
          ))}
        </div>
      )}

      {/* 图片预览模态框 */}
      <Modal
        open={previewVisible}
        onCancel={handlePreviewClose}
        footer={null}
        width='90vw'
        style={{ top: 20 }}
        styles={{
          body: {
            padding: 0,
            height: '80vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: '#000',
          },
        }}
        destroyOnClose
      >
        <div
          style={{
            position: 'relative',
            width: '100%',
            height: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {/* 预览图片 */}
          <img
            src={
              imageError[previewIndex]
                ? '/images/product-placeholder.svg'
                : images[previewIndex]
            }
            alt={`${productName} - 预览 ${previewIndex + 1}`}
            style={{
              maxWidth: '100%',
              maxHeight: '100%',
              transform: `scale(${zoom}) rotate(${rotation}deg)`,
              transition: 'transform 0.3s ease',
              objectFit: 'contain',
            }}
          />

          {/* 预览控制按钮 */}
          <div
            style={{
              position: 'absolute',
              top: 20,
              right: 20,
              display: 'flex',
              gap: 8,
            }}
          >
            <Space>
              <Tooltip title='放大'>
                <Button
                  type='text'
                  icon={<ZoomInOutlined />}
                  onClick={handleZoomIn}
                  style={{
                    color: 'white',
                    backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  }}
                />
              </Tooltip>
              <Tooltip title='缩小'>
                <Button
                  type='text'
                  icon={<ZoomOutOutlined />}
                  onClick={handleZoomOut}
                  style={{
                    color: 'white',
                    backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  }}
                />
              </Tooltip>
              <Tooltip title='向左旋转'>
                <Button
                  type='text'
                  icon={<RotateLeftOutlined />}
                  onClick={handleRotateLeft}
                  style={{
                    color: 'white',
                    backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  }}
                />
              </Tooltip>
              <Tooltip title='向右旋转'>
                <Button
                  type='text'
                  icon={<RotateRightOutlined />}
                  onClick={handleRotateRight}
                  style={{
                    color: 'white',
                    backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  }}
                />
              </Tooltip>
              <Tooltip title='下载图片'>
                <Button
                  type='text'
                  icon={<DownloadOutlined />}
                  onClick={handleDownload}
                  style={{
                    color: 'white',
                    backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  }}
                />
              </Tooltip>
              <Tooltip title={isFullscreen ? '退出全屏' : '全屏'}>
                <Button
                  type='text'
                  icon={
                    isFullscreen ? (
                      <FullscreenExitOutlined />
                    ) : (
                      <FullscreenOutlined />
                    )
                  }
                  onClick={handleFullscreen}
                  style={{
                    color: 'white',
                    backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  }}
                />
              </Tooltip>
            </Space>
          </div>

          {/* 预览导航按钮 */}
          {images.length > 1 && (
            <>
              <Button
                type='text'
                icon={<LeftOutlined />}
                onClick={handlePreviewPrev}
                style={{
                  position: 'absolute',
                  left: 20,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  color: 'white',
                  border: 'none',
                  borderRadius: '50%',
                  width: 48,
                  height: 48,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              />
              <Button
                type='text'
                icon={<RightOutlined />}
                onClick={handlePreviewNext}
                style={{
                  position: 'absolute',
                  right: 20,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  color: 'white',
                  border: 'none',
                  borderRadius: '50%',
                  width: 48,
                  height: 48,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              />
            </>
          )}

          {/* 预览缩略图导航 */}
          {images.length > 1 && (
            <div
              style={{
                position: 'absolute',
                bottom: 20,
                left: '50%',
                transform: 'translateX(-50%)',
                display: 'flex',
                gap: 8,
                backgroundColor: 'rgba(0, 0, 0, 0.5)',
                padding: '8px 16px',
                borderRadius: 20,
              }}
            >
              {images.map((image, index) => (
                <div
                  key={index}
                  style={{
                    width: 40,
                    height: 40,
                    border:
                      index === previewIndex
                        ? '2px solid #1890ff'
                        : '2px solid transparent',
                    borderRadius: 4,
                    overflow: 'hidden',
                    cursor: 'pointer',
                    transition: 'all 0.3s ease',
                  }}
                  onClick={() => handlePreviewChange(index)}
                >
                  <img
                    src={
                      imageError[index]
                        ? '/images/product-placeholder.svg'
                        : image
                    }
                    alt={`缩略图 ${index + 1}`}
                    style={{
                      width: '100%',
                      height: '100%',
                      objectFit: 'cover',
                    }}
                  />
                </div>
              ))}
            </div>
          )}

          {/* 图片信息 */}
          <div
            style={{
              position: 'absolute',
              bottom: 20,
              left: 20,
              color: 'white',
              backgroundColor: 'rgba(0, 0, 0, 0.5)',
              padding: '8px 12px',
              borderRadius: 4,
              fontSize: 14,
            }}
          >
            {previewIndex + 1} / {images.length}
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default ProductImageGallery;
