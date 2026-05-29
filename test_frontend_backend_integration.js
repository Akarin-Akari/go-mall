// 前后端集成测试脚本
// 测试前端(localhost:3000)与后端(localhost:8081)的API通信

// 使用Node.js内置的fetch API (Node 18+)

// 配置
const FRONTEND_URL = 'http://localhost:3000';
const BACKEND_URL = 'http://localhost:8081';

// 测试结果统计
let testResults = {
  total: 0,
  passed: 0,
  failed: 0,
  errors: []
};

// 测试工具函数
function logTest(testName, status, message = '') {
  testResults.total++;
  if (status === 'PASS') {
    testResults.passed++;
    console.log(`✅ ${testName}: PASS ${message}`);
  } else {
    testResults.failed++;
    console.log(`❌ ${testName}: FAIL ${message}`);
    testResults.errors.push(`${testName}: ${message}`);
  }
}

// 测试前端服务可访问性
async function testFrontendAccessibility() {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);

    const response = await fetch(FRONTEND_URL, {
      signal: controller.signal
    });
    clearTimeout(timeoutId);

    if (response.ok) {
      logTest('前端服务可访问性', 'PASS', `状态码: ${response.status}`);
      return true;
    } else {
      logTest('前端服务可访问性', 'FAIL', `状态码: ${response.status}`);
      return false;
    }
  } catch (error) {
    logTest('前端服务可访问性', 'FAIL', `错误: ${error.message}`);
    return false;
  }
}

// 测试后端服务可访问性
async function testBackendAccessibility() {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);

    const response = await fetch(`${BACKEND_URL}/api/v1/products?page=1&page_size=1`, {
      signal: controller.signal
    });
    clearTimeout(timeoutId);

    if (response.ok) {
      logTest('后端服务可访问性', 'PASS', `状态码: ${response.status}`);
      return true;
    } else {
      logTest('后端服务可访问性', 'FAIL', `状态码: ${response.status}`);
      return false;
    }
  } catch (error) {
    logTest('后端服务可访问性', 'FAIL', `错误: ${error.message}`);
    return false;
  }
}

// 测试CORS配置
async function testCORSConfiguration() {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);

    // 模拟前端发起的跨域请求
    const response = await fetch(`${BACKEND_URL}/api/v1/products?page=1&page_size=1`, {
      headers: {
        'Origin': FRONTEND_URL,
        'Access-Control-Request-Method': 'GET',
        'Access-Control-Request-Headers': 'Content-Type,Authorization'
      },
      signal: controller.signal
    });
    clearTimeout(timeoutId);

    if (response.ok) {
      logTest('CORS配置', 'PASS', '跨域请求成功');
      return true;
    } else {
      logTest('CORS配置', 'FAIL', `状态码: ${response.status}`);
      return false;
    }
  } catch (error) {
    logTest('CORS配置', 'FAIL', `错误: ${error.message}`);
    return false;
  }
}

// 测试用户认证API
async function testUserAuthenticationAPI() {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000);

    // 测试登录API
    const loginResponse = await fetch(`${BACKEND_URL}/api/v1/users/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Origin': FRONTEND_URL
      },
      body: JSON.stringify({
        username: 'admin',
        password: 'admin123'
      }),
      signal: controller.signal
    });
    clearTimeout(timeoutId);

    if (loginResponse.ok) {
      const loginData = await loginResponse.json();
      if (loginData.code === 200) {
        const token = loginData.data.token;
        logTest('用户登录API', 'PASS', `Token获取成功: ${token.substring(0, 20)}...`);

        // 测试需要认证的API
        const profileController = new AbortController();
        const profileTimeoutId = setTimeout(() => profileController.abort(), 5000);

        const profileResponse = await fetch(`${BACKEND_URL}/api/v1/users/profile`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Origin': FRONTEND_URL
          },
          signal: profileController.signal
        });
        clearTimeout(profileTimeoutId);

        if (profileResponse.ok) {
          logTest('认证API访问', 'PASS', '用户信息获取成功');
          return token;
        } else {
          logTest('认证API访问', 'FAIL', `状态码: ${profileResponse.status}`);
          return null;
        }
      } else {
        logTest('用户登录API', 'FAIL', `登录失败: ${loginData.message || '未知错误'}`);
        return null;
      }
    } else {
      logTest('用户登录API', 'FAIL', `状态码: ${loginResponse.status}`);
      return null;
    }
  } catch (error) {
    logTest('用户认证API', 'FAIL', `错误: ${error.message}`);
    return null;
  }
}

// 测试商品API
async function testProductAPI(token) {
  try {
    const headers = {
      'Origin': FRONTEND_URL
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    // 测试商品列表API
    const listController = new AbortController();
    const listTimeoutId = setTimeout(() => listController.abort(), 5000);

    const listResponse = await fetch(`${BACKEND_URL}/api/v1/products?page=1&page_size=10`, {
      headers,
      signal: listController.signal
    });
    clearTimeout(listTimeoutId);

    if (listResponse.ok) {
      const listData = await listResponse.json();
      logTest('商品列表API', 'PASS', `获取到 ${listData.data?.items?.length || 0} 个商品`);

      // 测试商品详情API
      const detailController = new AbortController();
      const detailTimeoutId = setTimeout(() => detailController.abort(), 5000);

      const detailResponse = await fetch(`${BACKEND_URL}/api/v1/products/1`, {
        headers,
        signal: detailController.signal
      });
      clearTimeout(detailTimeoutId);

      if (detailResponse.ok) {
        const detailData = await detailResponse.json();
        logTest('商品详情API', 'PASS', `商品名称: ${detailData.data?.name || '未知'}`);
      } else {
        logTest('商品详情API', 'FAIL', `状态码: ${detailResponse.status}`);
      }
    } else {
      logTest('商品列表API', 'FAIL', `状态码: ${listResponse.status}`);
    }
  } catch (error) {
    logTest('商品API', 'FAIL', `错误: ${error.message}`);
  }
}

// 测试购物车API
async function testCartAPI(token) {
  if (!token) {
    logTest('购物车API', 'FAIL', '需要认证Token');
    return;
  }

  try {
    const headers = {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      'Origin': FRONTEND_URL
    };

    // 测试获取购物车
    const cartController = new AbortController();
    const cartTimeoutId = setTimeout(() => cartController.abort(), 5000);

    const cartResponse = await fetch(`${BACKEND_URL}/api/v1/cart`, {
      headers,
      signal: cartController.signal
    });
    clearTimeout(cartTimeoutId);

    if (cartResponse.ok) {
      const cartData = await cartResponse.json();
      logTest('购物车获取API', 'PASS', `购物车商品数: ${cartData.data?.items?.length || 0}`);

      // 测试添加商品到购物车
      const addController = new AbortController();
      const addTimeoutId = setTimeout(() => addController.abort(), 5000);

      const addResponse = await fetch(`${BACKEND_URL}/api/v1/cart/add`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          product_id: 1,
          quantity: 1
        }),
        signal: addController.signal
      });
      clearTimeout(addTimeoutId);

      if (addResponse.ok) {
        logTest('购物车添加API', 'PASS', '商品添加成功');
      } else {
        logTest('购物车添加API', 'FAIL', `状态码: ${addResponse.status}`);
      }
    } else {
      logTest('购物车获取API', 'FAIL', `状态码: ${cartResponse.status}`);
    }
  } catch (error) {
    logTest('购物车API', 'FAIL', `错误: ${error.message}`);
  }
}

// 测试订单API
async function testOrderAPI(token) {
  if (!token) {
    logTest('订单API', 'FAIL', '需要认证Token');
    return;
  }

  try {
    const headers = {
      'Authorization': `Bearer ${token}`,
      'Origin': FRONTEND_URL
    };

    // 测试订单列表API
    const orderController = new AbortController();
    const orderTimeoutId = setTimeout(() => orderController.abort(), 5000);

    const orderResponse = await fetch(`${BACKEND_URL}/api/v1/orders?page=1&page_size=10`, {
      headers,
      signal: orderController.signal
    });
    clearTimeout(orderTimeoutId);

    if (orderResponse.ok) {
      const orderData = await orderResponse.json();
      logTest('订单列表API', 'PASS', `订单数量: ${orderData.data?.total || 0}`);
    } else {
      logTest('订单列表API', 'FAIL', `状态码: ${orderResponse.status}`);
    }
  } catch (error) {
    logTest('订单API', 'FAIL', `错误: ${error.message}`);
  }
}

// 主测试函数
async function runIntegrationTests() {
  console.log('🚀 开始前后端集成测试...\n');
  
  // 1. 测试服务可访问性
  console.log('📡 测试服务可访问性...');
  const frontendOK = await testFrontendAccessibility();
  const backendOK = await testBackendAccessibility();
  
  if (!frontendOK || !backendOK) {
    console.log('\n❌ 基础服务不可用，停止测试');
    return;
  }
  
  // 2. 测试CORS配置
  console.log('\n🌐 测试CORS配置...');
  await testCORSConfiguration();
  
  // 3. 测试用户认证
  console.log('\n🔐 测试用户认证...');
  const token = await testUserAuthenticationAPI();
  
  // 4. 测试商品API
  console.log('\n🛍️ 测试商品API...');
  await testProductAPI(token);
  
  // 5. 测试购物车API
  console.log('\n🛒 测试购物车API...');
  await testCartAPI(token);
  
  // 6. 测试订单API
  console.log('\n📦 测试订单API...');
  await testOrderAPI(token);
  
  // 输出测试结果
  console.log('\n' + '='.repeat(50));
  console.log('📊 测试结果统计:');
  console.log(`总测试数: ${testResults.total}`);
  console.log(`通过: ${testResults.passed} ✅`);
  console.log(`失败: ${testResults.failed} ❌`);
  console.log(`成功率: ${((testResults.passed / testResults.total) * 100).toFixed(1)}%`);
  
  if (testResults.errors.length > 0) {
    console.log('\n❌ 失败的测试:');
    testResults.errors.forEach(error => console.log(`  - ${error}`));
  }
  
  console.log('\n🎉 前后端集成测试完成!');
}

// 运行测试
runIntegrationTests().catch(error => {
  console.error('测试运行出错:', error);
  process.exit(1);
});
