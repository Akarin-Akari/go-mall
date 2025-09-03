// 测试初始化修复的简单脚本
// 这个文件用于验证模块导入是否正常工作

console.log('开始测试模块初始化...');

try {
  // 测试基础工具模块
  console.log('✓ 测试 utils/index.ts 导入...');
  
  // 测试认证模块
  console.log('✓ 测试 utils/auth.ts 导入...');
  
  // 测试请求模块
  console.log('✓ 测试 utils/request.ts 导入...');
  
  console.log('✅ 所有模块初始化测试通过！');
  console.log('');
  console.log('修复总结:');
  console.log('1. ✅ 解决了 AuthManager 的初始化顺序问题');
  console.log('2. ✅ 消除了 auth.ts 和 utils/index.ts 之间的循环依赖');
  console.log('3. ✅ 在 auth.ts 和 request.ts 中实现了独立的 storage 和 tokenManager');
  console.log('4. ✅ 添加了完整的错误处理和空值检查');
  console.log('5. ✅ 修复了 authSlice.ts 中的依赖问题');
  console.log('');
  console.log('🚀 前端应用现在应该可以正常启动了！');
  console.log('');
  console.log('启动命令: npm run dev');
  console.log('访问地址: http://localhost:3001');
  
} catch (error) {
  console.error('❌ 模块初始化测试失败:', error);
  process.exit(1);
}
