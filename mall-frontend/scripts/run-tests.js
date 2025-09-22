#!/usr/bin/env node

/**
 * 测试运行脚本
 * 自动化运行各种类型的测试并生成报告
 */

// eslint-disable-next-line @typescript-eslint/no-require-imports
const { execSync, spawn } = require('child_process');
// eslint-disable-next-line @typescript-eslint/no-require-imports
const fs = require('fs');
// eslint-disable-next-line @typescript-eslint/no-require-imports
const path = require('path');

// 颜色输出
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function logSection(title) {
  log('\n' + '='.repeat(60), 'cyan');
  log(`  ${title}`, 'bright');
  log('='.repeat(60), 'cyan');
}

function logSuccess(message) {
  log(`✅ ${message}`, 'green');
}

function logError(message) {
  log(`❌ ${message}`, 'red');
}

function logWarning(message) {
  log(`⚠️  ${message}`, 'yellow');
}

function logInfo(message) {
  log(`ℹ️  ${message}`, 'blue');
}

// 检查依赖
function checkDependencies() {
  logSection('检查依赖');

  try {
    // 检查Node.js版本
    const nodeVersion = process.version;
    logInfo(`Node.js版本: ${nodeVersion}`);

    // 检查npm版本
    const npmVersion = execSync('npm --version', { encoding: 'utf8' }).trim();
    logInfo(`npm版本: ${npmVersion}`);

    // 检查是否安装了依赖
    if (!fs.existsSync('node_modules')) {
      logWarning('node_modules不存在，正在安装依赖...');
      execSync('npm ci', { stdio: 'inherit' });
    }

    logSuccess('依赖检查完成');
    return true;
  } catch (error) {
    logError(`依赖检查失败: ${error.message}`);
    return false;
  }
}

// 运行单元测试
function runUnitTests() {
  logSection('运行单元测试');

  try {
    execSync('npm run test:unit -- --verbose', { stdio: 'inherit' });
    logSuccess('单元测试通过');
    return true;
  } catch (error) {
    logError('单元测试失败');
    return false;
  }
}

// 运行集成测试
function runIntegrationTests() {
  logSection('运行集成测试');

  try {
    execSync('npm run test:integration -- --verbose', { stdio: 'inherit' });
    logSuccess('集成测试通过');
    return true;
  } catch (error) {
    logError('集成测试失败');
    return false;
  }
}

// 生成覆盖率报告
function generateCoverageReport() {
  logSection('生成覆盖率报告');

  try {
    execSync('npm run test:coverage', { stdio: 'inherit' });

    // 检查覆盖率文件
    const coveragePath = path.join(process.cwd(), 'coverage');
    if (fs.existsSync(coveragePath)) {
      logSuccess('覆盖率报告生成成功');
      logInfo(`报告位置: ${coveragePath}/lcov-report/index.html`);

      // 读取覆盖率摘要
      const summaryPath = path.join(coveragePath, 'coverage-summary.json');
      if (fs.existsSync(summaryPath)) {
        const summary = JSON.parse(fs.readFileSync(summaryPath, 'utf8'));
        const total = summary.total;

        logInfo(`代码覆盖率统计:`);
        logInfo(`  语句覆盖率: ${total.statements.pct}%`);
        logInfo(`  分支覆盖率: ${total.branches.pct}%`);
        logInfo(`  函数覆盖率: ${total.functions.pct}%`);
        logInfo(`  行覆盖率: ${total.lines.pct}%`);

        // 检查是否达到目标覆盖率
        const threshold = 80;
        const overallCoverage =
          (total.statements.pct +
            total.branches.pct +
            total.functions.pct +
            total.lines.pct) /
          4;

        if (overallCoverage >= threshold) {
          logSuccess(
            `总体覆盖率 ${overallCoverage.toFixed(1)}% 达到目标 ${threshold}%`
          );
        } else {
          logWarning(
            `总体覆盖率 ${overallCoverage.toFixed(1)}% 未达到目标 ${threshold}%`
          );
        }
      }

      return true;
    } else {
      logWarning('覆盖率报告目录不存在');
      return false;
    }
  } catch (error) {
    logError(`覆盖率报告生成失败: ${error.message}`);
    return false;
  }
}

// 运行E2E测试
function runE2ETests() {
  logSection('运行E2E测试');

  return new Promise(resolve => {
    // 启动开发服务器
    logInfo('启动开发服务器...');
    const server = spawn('npm', ['run', 'dev'], {
      stdio: 'pipe',
      detached: false,
    });

    let serverReady = false;

    server.stdout.on('data', data => {
      const output = data.toString();
      if (output.includes('Ready') || output.includes('localhost:3001')) {
        serverReady = true;
        logSuccess('开发服务器启动成功');

        // 等待服务器完全启动
        setTimeout(() => {
          try {
            logInfo('运行Cypress E2E测试...');
            execSync('npm run test:e2e', { stdio: 'inherit' });
            logSuccess('E2E测试通过');
            resolve(true);
          } catch (error) {
            logError('E2E测试失败');
            resolve(false);
          } finally {
            // 关闭服务器
            server.kill('SIGTERM');
          }
        }, 3000);
      }
    });

    server.stderr.on('data', data => {
      const output = data.toString();
      if (!output.includes('Warning') && !output.includes('deprecated')) {
        logError(`服务器错误: ${output}`);
      }
    });

    // 超时处理
    setTimeout(() => {
      if (!serverReady) {
        logError('服务器启动超时');
        server.kill('SIGTERM');
        resolve(false);
      }
    }, 30000);
  });
}

// 生成测试报告
function generateTestReport(results) {
  logSection('生成测试报告');

  const report = {
    timestamp: new Date().toISOString(),
    results: results,
    summary: {
      total: Object.keys(results).length,
      passed: Object.values(results).filter(r => r === true).length,
      failed: Object.values(results).filter(r => r === false).length,
    },
  };

  // 创建报告目录
  const reportsDir = path.join(process.cwd(), 'test-reports');
  if (!fs.existsSync(reportsDir)) {
    fs.mkdirSync(reportsDir, { recursive: true });
  }

  // 写入报告文件
  const reportFile = path.join(reportsDir, `test-report-${Date.now()}.json`);
  fs.writeFileSync(reportFile, JSON.stringify(report, null, 2));

  logInfo(`测试报告已保存: ${reportFile}`);

  // 输出摘要
  logInfo('测试摘要:');
  Object.entries(results).forEach(([test, result]) => {
    if (result) {
      logSuccess(`  ${test}: 通过`);
    } else {
      logError(`  ${test}: 失败`);
    }
  });

  const successRate = (
    (report.summary.passed / report.summary.total) *
    100
  ).toFixed(1);
  logInfo(`总体成功率: ${successRate}%`);

  return report;
}

// 主函数
async function main() {
  log('🚀 开始运行Mall-Frontend测试套件', 'bright');

  const results = {};

  // 检查依赖
  if (!checkDependencies()) {
    process.exit(1);
  }

  // 运行测试
  results.unitTests = runUnitTests();
  results.integrationTests = runIntegrationTests();
  results.coverageReport = generateCoverageReport();

  // 运行E2E测试（可选）
  const runE2E =
    process.argv.includes('--e2e') || process.argv.includes('--all');
  if (runE2E) {
    results.e2eTests = await runE2ETests();
  }

  // 生成报告
  const report = generateTestReport(results);

  // 确定退出码
  const hasFailures = Object.values(results).some(result => result === false);

  if (hasFailures) {
    logError('部分测试失败');
    process.exit(1);
  } else {
    logSuccess('所有测试通过！');
    process.exit(0);
  }
}

// 处理命令行参数
if (require.main === module) {
  main().catch(error => {
    logError(`脚本执行失败: ${error.message}`);
    process.exit(1);
  });
}

module.exports = {
  checkDependencies,
  runUnitTests,
  runIntegrationTests,
  generateCoverageReport,
  runE2ETests,
  generateTestReport,
};
