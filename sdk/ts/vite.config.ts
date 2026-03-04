import { defineConfig } from 'vite'

export default defineConfig({
  build: {
    lib: {
      entry: 'src/index.ts',
      name: 'Tracely',
      // 生成多种格式以支持不同的使用场景
      formats: ['es', 'umd'],
      fileName: (format) => {
        if (format === 'es') {
          return 'tracely-sdk.mjs'
        }
        return 'tracely-sdk.js'
      },
    },
    outDir: 'dist',
    sourcemap: true,
    // 确保生成正确的 ES6 模块
    target: 'es2020',
  },
})