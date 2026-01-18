import { useState, useEffect, useCallback } from 'react'
import { themeParams } from '@/lib/telegram'

export type Theme = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'dumper-theme'

function getSystemTheme(): 'light' | 'dark' {
  // Check Telegram SDK first
  try {
    if (themeParams.isMounted() && themeParams.isDark()) {
      return 'dark'
    }
    if (themeParams.isMounted() && !themeParams.isDark()) {
      return 'light'
    }
  } catch { /* SDK not available */ }

  // Fallback to system preference
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return 'light'
}

function getStoredTheme(): Theme {
  if (typeof window === 'undefined') return 'system'
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'light' || stored === 'dark' || stored === 'system') {
    return stored
  }
  return 'system'
}

function applyTheme(theme: Theme) {
  const effectiveTheme = theme === 'system' ? getSystemTheme() : theme
  document.documentElement.classList.toggle('dark', effectiveTheme === 'dark')
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(getStoredTheme)

  const setTheme = useCallback((newTheme: Theme) => {
    setThemeState(newTheme)
    localStorage.setItem(STORAGE_KEY, newTheme)
    applyTheme(newTheme)
  }, [])

  // Apply theme on mount and when system theme changes
  useEffect(() => {
    applyTheme(theme)

    // Listen for system theme changes when in system mode
    if (theme === 'system') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      const handleChange = () => applyTheme('system')
      mediaQuery.addEventListener('change', handleChange)
      return () => mediaQuery.removeEventListener('change', handleChange)
    }
  }, [theme])

  const effectiveTheme = theme === 'system' ? getSystemTheme() : theme

  return { theme, setTheme, effectiveTheme }
}

// Initialize theme immediately on load (before React hydration)
export function initializeTheme() {
  applyTheme(getStoredTheme())
}
