// ***********************************************************
// This example support/e2e.ts is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.js using ES2015 syntax:
import './commands';

// Alternatively you can use CommonJS syntax:
// require('./commands')

// 导入代码覆盖率支持
import '@cypress/code-coverage/support';

// 全局配置
Cypress.on('uncaught:exception', (err, runnable) => {
  // 忽略某些预期的错误，防止测试失败
  if (err.message.includes('ResizeObserver loop limit exceeded')) {
    return false;
  }

  if (err.message.includes('Non-Error promise rejection captured')) {
    return false;
  }

  // 让其他错误正常失败
  return true;
});

// 在每个测试前清理状态
beforeEach(() => {
  // 清理localStorage
  cy.clearLocalStorage();

  // 清理sessionStorage
  cy.clearCookies();

  // 清理indexedDB
  cy.clearAllSessionStorage();

  // 设置默认视口
  cy.viewport(1280, 720);
});

// 在每个测试后进行清理
afterEach(() => {
  // 清理任何残留的网络拦截
  cy.window().then(win => {
    // 清理任何全局状态
    if (win.localStorage) {
      win.localStorage.clear();
    }
    if (win.sessionStorage) {
      win.sessionStorage.clear();
    }
  });
});

// 添加自定义命令类型声明
declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      /**
       * 登录用户
       * @param email 邮箱
       * @param password 密码
       */
      login(email: string, password: string): Chainable<void>;

      /**
       * 等待页面加载完成
       */
      waitForPageLoad(): Chainable<void>;

      /**
       * 检查无障碍性
       */
      checkA11y(): Chainable<void>;

      /**
       * 模拟网络延迟
       * @param delay 延迟时间（毫秒）
       */
      simulateNetworkDelay(delay: number): Chainable<void>;

      /**
       * 检查性能指标
       */
      checkPerformance(): Chainable<void>;
    }
  }
}
