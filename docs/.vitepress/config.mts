import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Bingo",
  description: "A production-ready Go microservice scaffold framework for rapid development",

  // 使用干净的 URL（无 .html 后缀）
  cleanUrls: true,

  // 重写规则：将 zh/ 映射到根路径
  rewrites: {
    'zh/:rest*': ':rest*'
  },

  // SEO 优化：Head 标签配置
  head: [
    // 基础 SEO Meta 标签（中英文混合，提升国际搜索可见度）
    ['meta', { name: 'keywords', content: 'Go,Golang,Go framework,microservices,scaffold,backend,API,gRPC,Gin,GORM,Redis,微服务,脚手架,框架,后端开发,中后台,Go语言,微服务架构' }],
    ['meta', { name: 'author', content: 'Bingo Team' }],

    // Open Graph 标签（社交媒体分享优化）
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'Bingo' }],
    ['meta', { property: 'og:title', content: 'Bingo - 生产级 Go 中后台脚手架' }],
    ['meta', { property: 'og:description', content: 'Bingo 是一个生产级的 Go 中后台脚手架，提供了完整的微服务架构、核心组件和最佳实践，帮助团队快速搭建可扩展的后端服务。' }],
    ['meta', { property: 'og:url', content: 'https://bingoctl.dev' }],
    ['meta', { property: 'og:image', content: 'https://bingoctl.dev/og-image.png' }],
    ['meta', { property: 'og:locale', content: 'zh_CN' }],
    ['meta', { property: 'og:locale:alternate', content: 'en_US' }],

    // Twitter Card 标签
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'Bingo - 生产级 Go 中后台脚手架' }],
    ['meta', { name: 'twitter:description', content: 'Bingo 是一个生产级的 Go 中后台脚手架，提供了完整的微服务架构、核心组件和最佳实践' }],
    ['meta', { name: 'twitter:image', content: 'https://bingoctl.dev/og-image.png' }],

    // Canonical URL（避免重复内容）
    ['link', { rel: 'canonical', href: 'https://bingoctl.dev' }],

    // Google Analytics（获取 ID：https://analytics.google.com/）
    // 1. 访问 Google Analytics 创建媒体资源
    // 2. 获取衡量 ID（格式：G-XXXXXXXXXX）
    // 3. 取消下面两行的注释并替换 YOUR_GA_ID
    ['script', { async: '', src: 'https://www.googletagmanager.com/gtag/js?id=G-XEQGM96B19' }],
    ['script', {}, `window.dataLayer = window.dataLayer || [];function gtag(){dataLayer.push(arguments);}gtag('js', new Date());gtag('config', 'G-XEQGM96B19');`],

    // 百度统计（获取代码：https://tongji.baidu.com/）
    // 1. 访问百度统计注册并添加网站
    // 2. 获取统计代码中的 hm.js?后面的ID
    // 3. 取消下面一行的注释并替换 YOUR_BAIDU_ID
    // ['script', {}, `var _hmt = _hmt || [];(function() {var hm = document.createElement("script");hm.src = "https://hm.baidu.com/hm.js?YOUR_BAIDU_ID";var s = document.getElementsByTagName("script")[0];s.parentNode.insertBefore(hm, s);})();`],
  ],

  // 站点地图配置
  sitemap: {
    hostname: 'https://bingoctl.dev',
    transformItems: (items) => {
      // 为不同类型的页面设置不同的优先级和更新频率
      return items.map((item) => {
        const url = item.url || ''

        // 首页：最高优先级，每日更新（匹配 / 或 /en/）
        if (url === '/' || url === '/en/' || url === '/index.html' || url === '/en/index.html') {
          item.priority = 1.0
          item.changefreq = 'daily'
        }
        // 指南页面：高优先级，每周更新
        else if (url.includes('/guide/')) {
          item.priority = 0.9
          item.changefreq = 'weekly'
        }
        // 核心概念页面：较高优先级，每周更新
        else if (url.includes('/essentials/')) {
          item.priority = 0.8
          item.changefreq = 'weekly'
        }
        // 其他页面：标准优先级，每月更新
        else {
          item.priority = 0.7
          item.changefreq = 'monthly'
        }
        return item
      })
    }
  },

  // 结构化数据配置
  transformHead: ({ pageData }) => {
    const head = []

    // 为首页添加 JSON-LD 结构化数据
    if (pageData.relativePath === 'index.md' || pageData.relativePath === 'zh/index.md') {
      head.push([
        'script',
        { type: 'application/ld+json' },
        JSON.stringify({
          '@context': 'https://schema.org',
          '@type': 'SoftwareApplication',
          'name': 'Bingo',
          'applicationCategory': 'DeveloperApplication',
          'operatingSystem': 'Linux, macOS, Windows',
          'description': 'Bingo 是一个生产级的 Go 中后台脚手架，提供了完整的微服务架构、核心组件和最佳实践',
          'url': 'https://bingoctl.dev',
          'author': {
            '@type': 'Organization',
            'name': 'Bingo Team'
          },
          'offers': {
            '@type': 'Offer',
            'price': '0',
            'priceCurrency': 'USD'
          },
          'programmingLanguage': 'Go'
        })
      ])
    }

    return head
  },

  locales: {
    root: {
      label: '中文',
      lang: 'zh-CN',
      title: "Bingo",
      description: "生产级 Go 脚手架，开箱即用的微服务解决方案",
      themeConfig: {
        nav: [
          { text: '指南', link: '/guide/what-is-bingo' },
          { text: '核心概念', link: '/essentials/architecture' },
          { text: '组件', link: '/components/overview' },
          {
            text: '更多',
            items: [
              { text: '开发规范', link: '/development/standards' },
              { text: '测试指南', link: '/development/testing' },
              { text: 'Docker 部署', link: '/deployment/docker' },
              { text: '微服务拆分', link: '/advanced/microservices' }
            ]
          },
          { text: 'GitHub', link: 'https://github.com/bingo-project/bingo' },
        ],
        sidebar: {
          '/guide/': [
            {
              text: '新手入门',
              items: [
                { text: '什么是 Bingo', link: '/guide/what-is-bingo' },
                { text: '快速开始', link: '/guide/getting-started' },
                { text: '项目结构', link: '/guide/project-structure' },
                { text: '开发第一个功能', link: '/guide/first-feature' },
                { text: '使用 bingo CLI', link: '/guide/using-bingo' }
              ]
            }
          ],
          '/essentials/': [
            {
              text: '核心概念',
              items: [
                { text: '整体架构', link: '/essentials/architecture' },
                { text: '分层架构详解', link: '/essentials/layered-design' },
                { text: 'Store 包设计', link: '/essentials/store' }
              ]
            },
            {
              text: '核心服务',
              items: [
                { text: 'API Server', link: '/essentials/apiserver' },
                // { text: 'Admin Server', link: '/essentials/admserver' },
                { text: 'Scheduler 调度器', link: '/essentials/scheduler' },
                { text: 'Bot 机器人服务', link: '/essentials/bot' }
              ]
            }
          ],
          '/components/': [
            {
              text: '组件参考',
              items: [
                { text: '核心组件概览', link: '/components/overview' }
              ]
            }
          ],
          '/development/': [
            {
              text: '开发规范',
              items: [
                { text: '代码规范', link: '/development/standards' },
                { text: '测试指南', link: '/development/testing' }
              ]
            }
          ],
          '/deployment/': [
            {
              text: '部署指南',
              items: [
                { text: 'Docker 部署', link: '/deployment/docker' }
              ]
            }
          ],
          '/advanced/': [
            {
              text: '进阶主题',
              items: [
                { text: '微服务拆分', link: '/advanced/microservices' }
              ]
            }
          ]
        }
      }
    },
    en: {
      label: 'English',
      lang: 'en',
      title: "Bingo",
      description: "A production-ready Go microservice scaffold framework for rapid development",
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/en/guide/what-is-bingo' },
          { text: 'Essentials', link: '/en/essentials/architecture' },
          { text: 'Components', link: '/en/components/overview' },
          {
            text: 'More',
            items: [
              { text: 'Development Standards', link: '/en/development/standards' },
              { text: 'Testing Guide', link: '/en/development/testing' },
              { text: 'Docker Deployment', link: '/en/deployment/docker' },
              { text: 'Microservice Decomposition', link: '/en/advanced/microservices' }
            ]
          },
          { text: 'GitHub', link: 'https://github.com/bingo-project/bingo' },
        ],
        sidebar: {
          '/en/guide/': [
            {
              text: 'Getting Started',
              items: [
                { text: 'What is Bingo', link: '/en/guide/what-is-bingo' },
                { text: 'Getting Started', link: '/en/guide/getting-started' },
                { text: 'Project Structure', link: '/en/guide/project-structure' },
                { text: 'First Feature', link: '/en/guide/first-feature' },
                { text: 'Using bingo CLI', link: '/en/guide/using-bingo' }
              ]
            }
          ],
          '/en/essentials/': [
            {
              text: 'Core Concepts',
              items: [
                { text: 'Overall Architecture', link: '/en/essentials/architecture' },
                { text: 'Layered Design', link: '/en/essentials/layered-design' },
                { text: 'Store Package Design', link: '/en/essentials/store' }
              ]
            },
            {
              text: 'Core Services',
              items: [
                { text: 'API Server', link: '/en/essentials/apiserver' },
                // { text: 'Admin Server', link: '/en/essentials/admserver' },
                { text: 'Scheduler', link: '/en/essentials/scheduler' },
                { text: 'Bot Service', link: '/en/essentials/bot' }
              ]
            }
          ],
          '/en/components/': [
            {
              text: 'Component Reference',
              items: [
                { text: 'Core Components Overview', link: '/en/components/overview' }
              ]
            }
          ],
          '/en/development/': [
            {
              text: 'Development Standards',
              items: [
                { text: 'Coding Standards', link: '/en/development/standards' },
                { text: 'Testing Guide', link: '/en/development/testing' }
              ]
            }
          ],
          '/en/deployment/': [
            {
              text: 'Deployment Guide',
              items: [
                { text: 'Docker Deployment', link: '/en/deployment/docker' }
              ]
            }
          ],
          '/en/advanced/': [
            {
              text: 'Advanced Topics',
              items: [
                { text: 'Microservice Decomposition', link: '/en/advanced/microservices' }
              ]
            }
          ]
        }
      }
    }
  },

  themeConfig: {
    socialLinks: [
      { icon: 'github', link: 'https://github.com/bingo-project/bingo' }
    ],

    footer: {
      message: 'Released under the Apache 2.0 License.',
      copyright: 'Copyright © 2025-present Bingo Team'
    },

    search: {
      provider: 'local'
    }
  }
})
