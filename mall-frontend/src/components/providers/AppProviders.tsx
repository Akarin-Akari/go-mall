'use client';

import React from 'react';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { ConfigProvider, App, theme } from 'antd';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
// import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';

import { store, persistor } from '@/store';
import { useAppSelector } from '@/store';
import { selectTheme } from '@/store/slices/appSlice';
import ErrorBoundary from '@/components/common/ErrorBoundary';
import Loading from '@/components/common/Loading';

// 设置dayjs中文
dayjs.locale('zh-cn');

// 创建QueryClient实例
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      retryDelay: attemptIndex => Math.min(1000 * 2 ** attemptIndex, 30000),
      staleTime: 5 * 60 * 1000, // 5分钟
      gcTime: 10 * 60 * 1000, // 10分钟 (原cacheTime)
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 1,
    },
  },
});

// Ant Design主题配置组件
const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const currentTheme = useAppSelector(selectTheme);

  const antdTheme = {
    algorithm: currentTheme === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: {
      colorPrimary: '#1890ff',
      borderRadius: 6,
      colorBgContainer: currentTheme === 'dark' ? '#141414' : '#ffffff',
    },
    components: {
      Layout: {
        siderBg: currentTheme === 'dark' ? '#001529' : '#001529',
        triggerBg: currentTheme === 'dark' ? '#002140' : '#002140',
      },
      Menu: {
        darkItemBg: 'transparent',
        darkSubMenuItemBg: 'transparent',
      },
    },
  };

  return (
    <ConfigProvider
      locale={zhCN}
      theme={antdTheme}
      componentSize="middle"
    >
      <App>
        {children}
      </App>
    </ConfigProvider>
  );
};

// 应用提供者组件
interface AppProvidersProps {
  children: React.ReactNode;
}

const AppProviders: React.FC<AppProvidersProps> = ({ children }) => {
  return (
    <ErrorBoundary>
      <Provider store={store}>
        <PersistGate 
          loading={<Loading fullScreen text="正在加载应用..." />} 
          persistor={persistor}
        >
          <QueryClientProvider client={queryClient}>
            <ThemeProvider>
              {children}
            </ThemeProvider>
            {/* {process.env.NODE_ENV === 'development' && (
              <ReactQueryDevtools initialIsOpen={false} />
            )} */}
          </QueryClientProvider>
        </PersistGate>
      </Provider>
    </ErrorBoundary>
  );
};

export default AppProviders;
