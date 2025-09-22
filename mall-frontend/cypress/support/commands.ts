/// <reference types="cypress" />

// ***********************************************
// This example commands.ts shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************

/**
 * 登录用户
 */
Cypress.Commands.add('login', (email: string, password: string) => {
  cy.session([email, password], () => {
    cy.visit('/login');
    cy.get('[data-testid="email-input"]').type(email);
    cy.get('[data-testid="password-input"]').type(password);
    cy.get('[data-testid="submit-login"]').click();

    // 等待登录成功
    cy.url().should('not.include', '/login');
    cy.get('[data-testid="user-menu"]').should('be.visible');
  });
});

/**
 * 等待页面加载完成
 */
Cypress.Commands.add('waitForPageLoad', () => {
  // 等待页面加载完成
  cy.get('[data-testid="app-loaded"]', { timeout: 10000 }).should('exist');

  // 等待所有图片加载完成
  cy.get('img')
    .should('be.visible')
    .and($imgs => {
      $imgs.each((index, img) => {
        expect(img.naturalWidth).to.be.greaterThan(0);
      });
    });

  // 等待网络请求完成
  cy.window().then(win => {
    return new Cypress.Promise(resolve => {
      const checkPendingRequests = () => {
        // 检查是否有pending的fetch请求
        if (win.fetch && (win.fetch as any).__pendingRequests === 0) {
          resolve();
        } else {
          setTimeout(checkPendingRequests, 100);
        }
      };
      checkPendingRequests();
    });
  });
});

/**
 * 检查无障碍性
 */
Cypress.Commands.add('checkA11y', () => {
  // 检查基本的无障碍性要求

  // 检查所有图片都有alt属性
  cy.get('img').each($img => {
    cy.wrap($img).should('have.attr', 'alt');
  });

  // 检查表单标签
  cy.get('input, textarea, select').each($input => {
    const id = $input.attr('id');
    const ariaLabel = $input.attr('aria-label');
    const ariaLabelledby = $input.attr('aria-labelledby');

    if (id) {
      cy.get(`label[for="${id}"]`).should('exist');
    } else {
      expect(ariaLabel || ariaLabelledby).to.exist;
    }
  });

  // 检查按钮有可访问的名称
  cy.get('button').each($button => {
    const text = $button.text().trim();
    const ariaLabel = $button.attr('aria-label');
    const ariaLabelledby = $button.attr('aria-labelledby');

    expect(text || ariaLabel || ariaLabelledby).to.exist;
  });

  // 检查标题层级
  cy.get('h1, h2, h3, h4, h5, h6').then($headings => {
    const headings = Array.from($headings).map(h =>
      parseInt(h.tagName.charAt(1))
    );

    // 检查是否有h1
    expect(headings).to.include(1);

    // 检查标题层级是否合理（不跳级）
    for (let i = 1; i < headings.length; i++) {
      const diff = headings[i] - headings[i - 1];
      expect(diff).to.be.at.most(1);
    }
  });
});

/**
 * 模拟网络延迟
 */
Cypress.Commands.add('simulateNetworkDelay', (delay: number) => {
  cy.intercept('**', req => {
    req.reply(res => {
      return new Promise(resolve => {
        setTimeout(() => resolve(res), delay);
      });
    });
  });
});

/**
 * 检查性能指标
 */
Cypress.Commands.add('checkPerformance', () => {
  cy.window().then(win => {
    // 检查Performance API是否可用
    if (!win.performance) {
      cy.log('Performance API not available');
      return;
    }

    const navigation = win.performance.getEntriesByType(
      'navigation'
    )[0] as PerformanceNavigationTiming;

    if (navigation) {
      // 计算关键性能指标
      const metrics = {
        // 首次内容绘制时间
        fcp: navigation.responseStart - navigation.navigationStart,
        // DOM内容加载完成时间
        domContentLoaded:
          navigation.domContentLoadedEventEnd - navigation.navigationStart,
        // 页面完全加载时间
        loadComplete: navigation.loadEventEnd - navigation.navigationStart,
        // DNS查询时间
        dnsLookup: navigation.domainLookupEnd - navigation.domainLookupStart,
        // TCP连接时间
        tcpConnect: navigation.connectEnd - navigation.connectStart,
        // 服务器响应时间
        serverResponse: navigation.responseEnd - navigation.requestStart,
      };

      // 记录性能指标
      cy.log('Performance Metrics:', metrics);

      // 性能断言
      expect(
        metrics.loadComplete,
        'Page load time should be less than 3 seconds'
      ).to.be.lessThan(3000);
      expect(
        metrics.domContentLoaded,
        'DOM content loaded should be less than 2 seconds'
      ).to.be.lessThan(2000);
      expect(
        metrics.serverResponse,
        'Server response time should be less than 1 second'
      ).to.be.lessThan(1000);
    }

    // 检查内存使用情况
    if ('memory' in win.performance) {
      const memory = (win.performance as any).memory;
      const memoryUsageMB = memory.usedJSHeapSize / 1024 / 1024;

      cy.log(`Memory usage: ${memoryUsageMB.toFixed(2)} MB`);

      // 内存使用不应超过50MB
      expect(memoryUsageMB, 'Memory usage should be reasonable').to.be.lessThan(
        50
      );
    }

    // 检查资源加载性能
    const resources = win.performance.getEntriesByType('resource');
    const slowResources = resources.filter(
      (resource: any) => resource.duration > 1000
    );

    if (slowResources.length > 0) {
      cy.log(
        'Slow resources detected:',
        slowResources.map((r: any) => r.name)
      );
    }

    // 慢资源不应超过总资源的10%
    expect(
      slowResources.length / resources.length,
      'Slow resources ratio should be low'
    ).to.be.lessThan(0.1);
  });
});

/**
 * 添加截图命令增强
 */
Cypress.Commands.overwrite(
  'screenshot',
  (originalFn, subject, name, options) => {
    // 在截图前等待一下，确保页面稳定
    cy.wait(500);

    return originalFn(subject, name, {
      capture: 'viewport',
      clip: { x: 0, y: 0, width: 1280, height: 720 },
      ...options,
    });
  }
);

/**
 * 添加类型检查的get命令
 */
Cypress.Commands.add(
  'getByTestId',
  (
    testId: string,
    options?: Partial<
      Cypress.Loggable &
        Cypress.Timeoutable &
        Cypress.Withinable &
        Cypress.Shadow
    >
  ) => {
    return cy.get(`[data-testid="${testId}"]`, options);
  }
);

/**
 * 等待元素可见并可交互
 */
Cypress.Commands.add('waitForInteractable', (selector: string) => {
  cy.get(selector)
    .should('be.visible')
    .should('not.be.disabled')
    .should('not.have.attr', 'aria-disabled', 'true');
});

// 扩展Cypress命令类型
declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      getByTestId(
        testId: string,
        options?: Partial<
          Cypress.Loggable &
            Cypress.Timeoutable &
            Cypress.Withinable &
            Cypress.Shadow
        >
      ): Chainable<JQuery<HTMLElement>>;
      waitForInteractable(selector: string): Chainable<JQuery<HTMLElement>>;
    }
  }
}
