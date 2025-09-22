import { defineConfig } from 'cypress';

export default defineConfig({
  e2e: {
    baseUrl: 'http://localhost:3001',
    supportFile: 'cypress/support/e2e.ts',
    specPattern: 'cypress/e2e/**/*.cy.{js,jsx,ts,tsx}',
    viewportWidth: 1280,
    viewportHeight: 720,
    video: true,
    screenshotOnRunFailure: true,
    defaultCommandTimeout: 10000,
    requestTimeout: 10000,
    responseTimeout: 10000,
    pageLoadTimeout: 30000,

    env: {
      // 测试环境变量
      apiUrl: 'http://localhost:8080',
      testUser: {
        email: 'test@example.com',
        password: 'password123',
      },
    },

    setupNodeEvents(on, config) {
      // 实现node事件监听器
      on('task', {
        log(message) {
          console.log(message);
          return null;
        },

        // 清理测试数据
        clearTestData() {
          // 这里可以添加清理测试数据的逻辑
          return null;
        },

        // 设置测试数据
        seedTestData() {
          // 这里可以添加设置测试数据的逻辑
          return null;
        },
      });

      // 代码覆盖率配置（如果安装了@cypress/code-coverage）
      // require('@cypress/code-coverage/task')(on, config);

      return config;
    },
  },

  component: {
    devServer: {
      framework: 'next',
      bundler: 'webpack',
    },
    specPattern: 'src/**/*.cy.{js,jsx,ts,tsx}',
    supportFile: 'cypress/support/component.ts',
  },

  // 全局配置
  retries: {
    runMode: 2,
    openMode: 0,
  },

  // 实验性功能
  experimentalStudio: true,
  experimentalMemoryManagement: true,
});
