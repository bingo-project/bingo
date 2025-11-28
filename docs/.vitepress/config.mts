import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Bingo",
  description: "Production-grade Go backend scaffold with complete microservice architecture",

  // 重写规则：将 zh/ 映射到根路径
  rewrites: {
    'zh/:rest*': ':rest*'
  },

  // SEO 优化：Head 标签配置
  head: [
    // 基础 SEO Meta 标签
    ['meta', { name: 'keywords', content: 'Go,Golang,微服务,脚手架,框架,后端开发,中后台,API,gRPC,Gin,GORM,Redis' }],
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
  ],

  // 站点地图配置
  sitemap: {
    hostname: 'https://bingoctl.dev'
  },

  locales: {
    root: {
      label: '中文',
      lang: 'zh-CN',
      title: "Bingo",
      description: "Bingo 是一个生产级的 Go 中后台脚手架,提供了完整的微服务架构、核心组件和最佳实践,帮助团队快速搭建可扩展的后端服务。",
      themeConfig: {
        nav: [
          { text: '指南', link: '/guide/what-is-bingo' },
          { text: '核心概念', link: '/essentials/architecture' },
          { text: '组件', link: '/components/overview' },
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
                { text: '开发第一个功能', link: '/guide/first-feature' }
              ]
            }
          ],
          '/essentials/': [
            {
              text: '核心概念',
              items: [
                { text: '整体架构', link: '/essentials/architecture' },
                { text: 'API Server 详解', link: '/essentials/apiserver' },
                { text: '分层架构详解', link: '/essentials/layered-design' },
                { text: 'Store 包设计', link: '/essentials/store' }
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
          ]
        }
      }
    },
    en: {
      label: 'English',
      lang: 'en',
      title: "Bingo",
      description: "Production-grade Go backend scaffold with complete microservice architecture, core components, and best practices",
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/en/guide/what-is-bingo' },
          { text: 'Essentials', link: '/en/essentials/architecture' },
          { text: 'Components', link: '/en/components/overview' },
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
                { text: 'Using bingoctl', link: '/en/guide/using-bingoctl' }
              ]
            }
          ],
          '/en/essentials/': [
            {
              text: 'Core Concepts',
              items: [
                { text: 'Overall Architecture', link: '/en/essentials/architecture' },
                { text: 'API Server', link: '/en/essentials/apiserver' },
                { text: 'Layered Design', link: '/en/essentials/layered-design' },
                { text: 'Store Package Design', link: '/en/essentials/store' }
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
                { text: 'Coding Standards', link: '/en/development/standards' }
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
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025-present Bingo Team'
    },

    search: {
      provider: 'local'
    }
  }
})
