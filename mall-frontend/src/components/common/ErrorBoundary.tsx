'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Result, Button } from 'antd';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.setState({
      error,
      errorInfo,
    });

    // 这里可以将错误信息发送到错误监控服务
    // reportError(error, errorInfo);
  }

  handleReload = () => {
    window.location.reload();
  };

  handleGoHome = () => {
    window.location.href = '/';
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div style={{ 
          padding: '50px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '400px'
        }}>
          <Result
            status="error"
            title="页面出现错误"
            subTitle="抱歉，页面遇到了一些问题。请尝试刷新页面或返回首页。"
            extra={[
              <Button type="primary" key="reload" onClick={this.handleReload}>
                刷新页面
              </Button>,
              <Button key="home" onClick={this.handleGoHome}>
                返回首页
              </Button>,
            ]}
          >
            {process.env.NODE_ENV === 'development' && (
              <div style={{ 
                textAlign: 'left', 
                marginTop: 20,
                padding: 16,
                backgroundColor: '#f5f5f5',
                borderRadius: 4,
                fontSize: 12,
                fontFamily: 'monospace'
              }}>
                <details>
                  <summary style={{ cursor: 'pointer', marginBottom: 8 }}>
                    错误详情 (开发模式)
                  </summary>
                  <div>
                    <strong>错误信息:</strong>
                    <pre style={{ whiteSpace: 'pre-wrap', margin: '8px 0' }}>
                      {this.state.error?.toString()}
                    </pre>
                  </div>
                  {this.state.errorInfo && (
                    <div>
                      <strong>组件堆栈:</strong>
                      <pre style={{ whiteSpace: 'pre-wrap', margin: '8px 0' }}>
                        {this.state.errorInfo.componentStack}
                      </pre>
                    </div>
                  )}
                </details>
              </div>
            )}
          </Result>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
