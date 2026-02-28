/**
 * Tracely SDK 测试脚本 - 用于填充测试数据
 * 使用 Bun 运行：bun run test-data.ts
 */

import CryptoJS from 'crypto-js'

/**
 * 生成随机 Nonce（32 位十六进制，与 Go SDK 保持一致）
 */
function generateNonce(): string {
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  return Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('')
}

/**
 * 生成 HMAC-SHA256 签名
 * 算法：HMAC-SHA256(appId + timestamp + nonce, appSecret)
 */
function generateSignature(
  appID: string,
  appSecret: string,
  timestamp: string,
  nonce: string
): string {
  const raw = appID + timestamp + nonce
  return CryptoJS.HmacSHA256(raw, appSecret).toString(CryptoJS.enc.Hex)
}

/**
 * 生成认证请求头（使用秒级时间戳，与 Go SDK 保持一致）
 */
function buildHeaders(appID: string, appSecret: string): Record<string, string> {
  // 使用秒级时间戳（与 Go 的 time.Now().Unix() 一致）
  const timestamp = Math.floor(Date.now() / 1000).toString()
  const nonce = generateNonce()
  const signature = generateSignature(appID, appSecret, timestamp, nonce)

  return {
    'X-App-Id': appID,
    'X-Timestamp': timestamp,
    'X-Nonce': nonce,
    'X-Signature': signature,
  }
}

/**
 * 上报错误数据
 */
async function reportError(
  host: string,
  appId: string,
  appSecret: string,
  errorData: {
    type: string
    message: string
    stack?: string
    url: string
  }
): Promise<void> {
  const headers = buildHeaders(appId, appSecret)
  
  // 添加 appId 到请求体中
  const payload = {
    ...errorData,
    appId: appId,
  }
  
  try {
    const response = await fetch(host + '/report/error', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: JSON.stringify(payload),
    })
    
    if (response.ok) {
      console.log('✓ 错误上报成功:', errorData.message)
    } else {
      const errorText = await response.text()
      console.error('✗ 错误上报失败:', response.status, response.statusText, errorText)
    }
  } catch (error) {
    console.error('✗ 错误上报异常:', error)
  }
}

/**
 * 上报活跃数据
 */
async function reportActive(
  host: string,
  appId: string,
  appSecret: string,
  activeData: {
    appId: string
    userId: string
    page: string
    duration: number
  }
): Promise<void> {
  const headers = buildHeaders(appId, appSecret)
  
  try {
    const response = await fetch(host + '/report/active', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: JSON.stringify(activeData),
    })
    
    if (response.ok) {
      console.log('✓ 活跃数据上报成功:', activeData.page, `(${activeData.duration}s)`)
    } else {
      console.error('✗ 活跃数据上报失败:', response.status, response.statusText)
    }
  } catch (error) {
    console.error('✗ 活跃数据上报异常:', error)
  }
}

/**
 * 生成随机用户 ID
 */
function generateUserId(): string {
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15)
}

/**
 * 主函数 - 生成测试数据
 */
async function main() {
  // 配置信息（请根据实际情况修改）
  const config = {
    appId: 'my-app-id',
    appSecret: 'my-app-secret-please-change-this-to-32-chars',
    host: 'http://localhost:3001',
  }

  console.log('🚀 开始生成测试数据...\n')
  console.log('配置:', config)
  console.log('')

  // 测试页面列表
  const pages = [
    '/dashboard',
    '/dashboard/overview',
    '/dashboard/errors',
    '/dashboard/active',
    '/dashboard/settings',
    '/dashboard/users',
    '/dashboard/reports',
    '/dashboard/analytics',
  ]

  // 错误类型列表
  const errorTypes = [
    { type: 'jsError', message: 'Uncaught TypeError: Cannot read property "name" of undefined' },
    { type: 'jsError', message: 'Uncaught ReferenceError: variable is not defined' },
    { type: 'jsError', message: 'Uncaught SyntaxError: Unexpected token' },
    { type: 'promiseError', message: 'Promise rejected: Network timeout' },
    { type: 'promiseError', message: 'Promise rejected: API response 500' },
    { type: 'promiseError', message: 'Promise rejected: Connection refused' },
    { type: 'manualError', message: 'User action failed: submit form' },
    { type: 'manualError', message: 'Validation failed: email format invalid' },
    { type: 'resourceError', message: 'Failed to load resource: net::ERR_FAILED' },
    { type: 'apiError', message: 'API Error: /api/users returned 403' },
  ]

  // 生成活跃数据（每个页面 20 条，共 160 条）
  console.log('📊 生成活跃数据...')
  for (let i = 0; i < pages.length; i++) {
    for (let j = 0; j < 20; j++) {
      const userId = generateUserId()
      const duration = Math.floor(Math.random() * 300) + 10 // 10-310 秒
      await reportActive(config.host, config.appId, config.appSecret, {
        appId: config.appId,
        userId,
        page: pages[i],
        duration,
      })
      // 延迟避免触发限流（60 次/分钟）
      await new Promise(resolve => setTimeout(resolve, 120))
    }
  }

  console.log('')

  // 生成错误数据（每种类型 10 条，共 100 条）
  console.log('❌ 生成错误数据...')
  for (let i = 0; i < errorTypes.length; i++) {
    for (let j = 0; j < 10; j++) {
      const error = errorTypes[i]
      await reportError(config.host, config.appId, config.appSecret, {
        type: error.type,
        message: error.message,
        stack: error.type.includes('Error') ? `Error: ${error.message}\n    at test.js:1:1` : undefined,
        url: `http://localhost:3000${pages[Math.floor(Math.random() * pages.length)]}`,
      })
      // 延迟避免触发限流（60 次/分钟）
      await new Promise(resolve => setTimeout(resolve, 1200))
    }
  }

  console.log('')
  console.log('✅ 测试数据生成完成！')
  console.log(`共计：${pages.length * 20} 条活跃数据，${errorTypes.length * 10} 条错误数据`)
}

// 运行测试
main().catch(console.error)