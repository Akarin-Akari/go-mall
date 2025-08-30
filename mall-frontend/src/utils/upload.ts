import { message } from 'antd';
import { http } from './request';
import { UPLOAD_CONFIG, API_ENDPOINTS } from '@/constants';
import { UploadFile } from '@/types';

// 上传配置接口
export interface UploadOptions {
  maxSize?: number;
  allowedTypes?: string[];
  multiple?: boolean;
  compress?: boolean;
  quality?: number;
  maxWidth?: number;
  maxHeight?: number;
  onProgress?: (percent: number) => void;
  onSuccess?: (response: any) => void;
  onError?: (error: Error) => void;
}

// 文件验证结果
interface ValidationResult {
  valid: boolean;
  error?: string;
}

// 文件上传管理器
export class UploadManager {
  private static instance: UploadManager;

  private constructor() {}

  public static getInstance(): UploadManager {
    if (!UploadManager.instance) {
      UploadManager.instance = new UploadManager();
    }
    return UploadManager.instance;
  }

  // 验证文件
  validateFile(file: File, options: UploadOptions = {}): ValidationResult {
    const {
      maxSize = UPLOAD_CONFIG.MAX_SIZE,
      allowedTypes = UPLOAD_CONFIG.ALLOWED_IMAGE_TYPES,
    } = options;

    // 检查文件大小
    if (file.size > maxSize) {
      return {
        valid: false,
        error: `文件大小不能超过 ${this.formatFileSize(maxSize)}`,
      };
    }

    // 检查文件类型
    if (!allowedTypes.includes(file.type)) {
      return {
        valid: false,
        error: `不支持的文件类型，支持的类型：${allowedTypes.join(', ')}`,
      };
    }

    return { valid: true };
  }

  // 格式化文件大小
  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  // 压缩图片
  async compressImage(
    file: File,
    options: {
      quality?: number;
      maxWidth?: number;
      maxHeight?: number;
    } = {}
  ): Promise<File> {
    const { quality = 0.8, maxWidth = 1920, maxHeight = 1080 } = options;

    return new Promise((resolve, reject) => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      const img = new Image();

      img.onload = () => {
        // 计算压缩后的尺寸
        let { width, height } = img;
        
        if (width > maxWidth || height > maxHeight) {
          const ratio = Math.min(maxWidth / width, maxHeight / height);
          width *= ratio;
          height *= ratio;
        }

        canvas.width = width;
        canvas.height = height;

        // 绘制压缩后的图片
        ctx?.drawImage(img, 0, 0, width, height);

        canvas.toBlob(
          (blob) => {
            if (blob) {
              const compressedFile = new File([blob], file.name, {
                type: file.type,
                lastModified: Date.now(),
              });
              resolve(compressedFile);
            } else {
              reject(new Error('图片压缩失败'));
            }
          },
          file.type,
          quality
        );
      };

      img.onerror = () => reject(new Error('图片加载失败'));
      img.src = URL.createObjectURL(file);
    });
  }

  // 单文件上传
  async uploadFile(
    file: File,
    options: UploadOptions = {}
  ): Promise<{ url: string; filename: string }> {
    // 验证文件
    const validation = this.validateFile(file, options);
    if (!validation.valid) {
      throw new Error(validation.error);
    }

    // 压缩图片（如果需要）
    let uploadFile = file;
    if (options.compress && file.type.startsWith('image/')) {
      try {
        uploadFile = await this.compressImage(file, options);
      } catch (error) {
        console.warn('图片压缩失败，使用原文件上传:', error);
      }
    }

    // 创建FormData
    const formData = new FormData();
    formData.append('file', uploadFile);

    try {
      const response = await http.upload(API_ENDPOINTS.UPLOAD.IMAGE, formData, {
        showLoading: true,
        onUploadProgress: (progressEvent) => {
          if (options.onProgress && progressEvent.total) {
            const percent = Math.round(
              (progressEvent.loaded * 100) / progressEvent.total
            );
            options.onProgress(percent);
          }
        },
      });

      const result = {
        url: response.data.url,
        filename: response.data.filename,
      };

      options.onSuccess?.(result);
      return result;
    } catch (error) {
      options.onError?.(error as Error);
      throw error;
    }
  }

  // 多文件上传
  async uploadFiles(
    files: File[],
    options: UploadOptions = {}
  ): Promise<{ url: string; filename: string }[]> {
    const results: { url: string; filename: string }[] = [];
    const errors: Error[] = [];

    for (let i = 0; i < files.length; i++) {
      try {
        const result = await this.uploadFile(files[i], {
          ...options,
          onProgress: (percent) => {
            // 计算总体进度
            const totalPercent = Math.round(
              ((i * 100 + percent) / files.length)
            );
            options.onProgress?.(totalPercent);
          },
        });
        results.push(result);
      } catch (error) {
        errors.push(error as Error);
      }
    }

    if (errors.length > 0 && results.length === 0) {
      throw new Error(`所有文件上传失败: ${errors[0].message}`);
    }

    if (errors.length > 0) {
      message.warning(`${errors.length} 个文件上传失败`);
    }

    return results;
  }

  // Base64上传
  async uploadBase64(
    base64: string,
    filename: string,
    options: UploadOptions = {}
  ): Promise<{ url: string; filename: string }> {
    try {
      const response = await http.post(API_ENDPOINTS.UPLOAD.IMAGE, {
        file: base64,
        filename,
      });

      const result = {
        url: response.data.url,
        filename: response.data.filename,
      };

      options.onSuccess?.(result);
      return result;
    } catch (error) {
      options.onError?.(error as Error);
      throw error;
    }
  }

  // 删除文件
  async deleteFile(fileId: string): Promise<void> {
    await http.delete(API_ENDPOINTS.UPLOAD.DELETE(fileId));
  }
}

// 创建全局上传管理器实例
export const uploadManager = UploadManager.getInstance();

// 便捷的上传函数
export const uploadImage = (
  file: File,
  options?: UploadOptions
): Promise<{ url: string; filename: string }> => {
  return uploadManager.uploadFile(file, {
    allowedTypes: UPLOAD_CONFIG.ALLOWED_IMAGE_TYPES,
    compress: true,
    ...options,
  });
};

export const uploadImages = (
  files: File[],
  options?: UploadOptions
): Promise<{ url: string; filename: string }[]> => {
  return uploadManager.uploadFiles(files, {
    allowedTypes: UPLOAD_CONFIG.ALLOWED_IMAGE_TYPES,
    compress: true,
    ...options,
  });
};

// 文件选择器
export const selectFiles = (options: {
  accept?: string;
  multiple?: boolean;
}): Promise<File[]> => {
  return new Promise((resolve) => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = options.accept || 'image/*';
    input.multiple = options.multiple || false;

    input.onchange = (event) => {
      const files = Array.from((event.target as HTMLInputElement).files || []);
      resolve(files);
    };

    input.click();
  });
};

// 拖拽上传处理
export const handleDrop = (
  event: DragEvent,
  callback: (files: File[]) => void
): void => {
  event.preventDefault();
  event.stopPropagation();

  const files = Array.from(event.dataTransfer?.files || []);
  callback(files);
};

export const handleDragOver = (event: DragEvent): void => {
  event.preventDefault();
  event.stopPropagation();
};

// 粘贴上传处理
export const handlePaste = (
  event: ClipboardEvent,
  callback: (files: File[]) => void
): void => {
  const items = event.clipboardData?.items;
  if (!items) return;

  const files: File[] = [];
  for (let i = 0; i < items.length; i++) {
    const item = items[i];
    if (item.type.indexOf('image') !== -1) {
      const file = item.getAsFile();
      if (file) {
        files.push(file);
      }
    }
  }

  if (files.length > 0) {
    callback(files);
  }
};

// 图片预览
export const previewImage = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      resolve(e.target?.result as string);
    };
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
};

// 获取图片信息
export const getImageInfo = (file: File): Promise<{
  width: number;
  height: number;
  size: number;
  type: string;
}> => {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => {
      resolve({
        width: img.width,
        height: img.height,
        size: file.size,
        type: file.type,
      });
    };
    img.onerror = reject;
    img.src = URL.createObjectURL(file);
  });
};

// Ant Design Upload组件的自定义上传函数
export const customUploadRequest = (options: any) => {
  const { file, onProgress, onSuccess, onError } = options;

  uploadManager
    .uploadFile(file, {
      onProgress,
      onSuccess,
      onError,
    })
    .then((result) => {
      onSuccess(result, file);
    })
    .catch((error) => {
      onError(error);
    });
};

// 转换为Ant Design Upload组件需要的文件列表格式
export const transformToUploadFileList = (
  urls: string[]
): UploadFile[] => {
  return urls.map((url, index) => ({
    uid: `${index}`,
    name: `image-${index}`,
    status: 'done',
    url,
  }));
};

// 从Ant Design Upload组件的文件列表提取URL
export const extractUrlsFromFileList = (
  fileList: UploadFile[]
): string[] => {
  return fileList
    .filter((file) => file.status === 'done' && file.url)
    .map((file) => file.url!);
};
