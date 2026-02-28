import CryptoJS from 'crypto-js'

/**
 * 生成随机 Nonce（UUID 去掉横线）
 */
export function generateNonce(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID().replace(/-/g, '')
  }
  // 降级方案
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15)
}

/**
 * 生成 HMAC-SHA256 签名
 * 算法：HMAC-SHA256(appId + timestamp + nonce, appSecret)
 */
export function generateSignature(
  appID: string,
  appSecret: string,
  timestamp: string,
  nonce: string
): string {
  const raw = appID + timestamp + nonce
  return CryptoJS.HmacSHA256(raw, appSecret).toString(CryptoJS.enc.Hex)
}

/**
 * 生成认证请求头
 */
export function buildHeaders(appID: string, appSecret: string): Record<string, string> {
  const timestamp = Date.now().toString()
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
 * 带签名的 fetch 请求
 */
export async function signedFetch(
  host: string,
  path: string,
  body: unknown,
  appID: string,
  appSecret: string
): Promise<void> {
  const headers = buildHeaders(appID, appSecret)
  
  try {
    await fetch(host + path, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: JSON.stringify(body),
      keepalive: true, // 确保页面关闭时请求不丢失
    })
  } catch (error) {
    // 上报失败不影响业务，静默失败
    console.warn('[Tracely] Failed to send report:', error)
  }
}
