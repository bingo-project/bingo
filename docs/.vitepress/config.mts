import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Bingo",
  description: "Production-grade Go backend scaffold with complete microservice architecture",

  locales: {
    root: {
      label: 'English',
      lang: 'en',
      title: "Bingo",
      description: "Production-grade Go backend scaffold with complete microservice architecture, core components, and best practices",
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/guide/what-is-bingo' },
          { text: 'Essentials', link: '/essentials/architecture' },
          { text: 'Components', link: '/components/overview' },
          { text: 'GitHub', link: 'https://github.com/bingo-project/bingo' },
          {
            text: '中文',
            link: '/zh/'
          }
        ],
        sidebar: {
          '/guide/': [
            {
              text: 'Getting Started',
              items: [
                { text: 'What is Bingo', link: '/guide/what-is-bingo' },
                { text: 'Getting Started', link: '/guide/getting-started' },
                { text: 'Project Structure', link: '/guide/project-structure' },
                { text: 'First Feature', link: '/guide/first-feature' }
              ]
            }
          ],
          '/essentials/': [
            {
              text: 'Core Concepts',
              items: [
                { text: 'Overall Architecture', link: '/essentials/architecture' },
                { text: 'Layered Design', link: '/essentials/layered-design' }
              ]
            }
          ],
          '/components/': [
            {
              text: 'Component Reference',
              items: [
                { text: 'Core Components Overview', link: '/components/overview' }
              ]
            }
          ]
        }
      }
    },
    zh: {
      label: '中文',
      lang: 'zh-CN',
      title: "Bingo",
      description: "Bingo 是一个生产级的 Go 中后台脚手架,提供了完整的微服务架构、核心组件和最佳实践,帮助团队快速搭建可扩展的后端服务。",
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '指南', link: '/zh/guide/what-is-bingo' },
          { text: '核心概念', link: '/zh/essentials/architecture' },
          { text: '组件', link: '/zh/components/overview' },
          { text: 'GitHub', link: 'https://github.com/bingo-project/bingo' },
          {
            text: 'English',
            link: '/'
          }
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '新手入门',
              items: [
                { text: '什么是 Bingo', link: '/zh/guide/what-is-bingo' },
                { text: '快速开始', link: '/zh/guide/getting-started' },
                { text: '项目结构', link: '/zh/guide/project-structure' },
                { text: '开发第一个功能', link: '/zh/guide/first-feature' }
              ]
            }
          ],
          '/zh/essentials/': [
            {
              text: '核心概念',
              items: [
                { text: '整体架构', link: '/zh/essentials/architecture' },
                { text: '分层架构详解', link: '/zh/essentials/layered-design' }
              ]
            }
          ],
          '/zh/components/': [
            {
              text: '组件参考',
              items: [
                { text: '核心组件概览', link: '/zh/components/overview' }
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
