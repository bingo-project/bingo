// VitePress 配置示例
// 使用时重命名为 .vitepress/config.js

export default {
  title: 'Bingo',
  description: '生产级 Go 中后台脚手架',

  themeConfig: {
    nav: [
      { text: '指南', link: '/guide/what-is-bingo' },
      { text: '核心概念', link: '/essentials/architecture' },
      { text: '组件', link: '/components/overview' },
      { text: 'GitHub', link: 'https://github.com/your-repo/bingo' }
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
            { text: '分层架构详解', link: '/essentials/layered-design' }
          ]
        }
      ],

      '/development/': [
        {
          text: '开发指南',
          items: [
            { text: '开发规范', link: '/development/standards' }
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

      '/deployment/': [
        {
          text: '部署运维',
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
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/your-repo/bingo' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present Bingo Team'
    },

    search: {
      provider: 'local'
    }
  }
}
