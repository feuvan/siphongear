import { defineStore } from 'pinia'

const THEME_KEY = 'sg-theme'
const MOBILE_QUERY = '(max-width: 768px)'

type Theme = 'light' | 'dark'

function applyTheme(theme: Theme) {
  const root = document.documentElement
  if (theme === 'dark') root.classList.add('dark')
  else root.classList.remove('dark')
}

function initialTheme(): Theme {
  const stored = localStorage.getItem(THEME_KEY) as Theme | null
  if (stored === 'light' || stored === 'dark') return stored
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark'
  return 'light'
}

let listenerInstalled = false

export const useUiStore = defineStore('ui', {
  state: () => ({
    theme: 'light' as Theme,
    isMobile: false,
    asideOpen: true
  }),
  actions: {
    init() {
      this.theme = initialTheme()
      applyTheme(this.theme)

      const mq = window.matchMedia(MOBILE_QUERY)
      this.isMobile = mq.matches
      this.asideOpen = !mq.matches
      if (!listenerInstalled) {
        const handler = (e: MediaQueryListEvent) => {
          this.isMobile = e.matches
          this.asideOpen = !e.matches
        }
        if (mq.addEventListener) mq.addEventListener('change', handler)
        else mq.addListener(handler)
        listenerInstalled = true
      }
    },
    setTheme(t: Theme) {
      this.theme = t
      localStorage.setItem(THEME_KEY, t)
      applyTheme(t)
    },
    toggleTheme() {
      this.setTheme(this.theme === 'dark' ? 'light' : 'dark')
    },
    toggleAside() {
      this.asideOpen = !this.asideOpen
    },
    openAside() {
      this.asideOpen = true
    },
    closeAside() {
      this.asideOpen = false
    }
  }
})
