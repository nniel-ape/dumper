import {
  init,
  miniApp,
  themeParams,
  viewport,
  backButton,
} from '@telegram-apps/sdk-react'

export function initTelegramApp() {
  try {
    init()

    // Expand viewport
    if (viewport.mount.isAvailable()) {
      viewport.mount()
      viewport.expand()
    }

    // Mount theme params
    if (themeParams.mount.isAvailable()) {
      themeParams.mount()
    }

    // Mount mini app
    if (miniApp.mount.isAvailable()) {
      miniApp.mount()
      miniApp.ready()
    }

    // Mount back button
    if (backButton.mount.isAvailable()) {
      backButton.mount()
    }

    return true
  } catch (e) {
    console.warn('Not running in Telegram:', e)
    return false
  }
}

export function getInitData(): string {
  try {
    // @ts-expect-error - Telegram WebApp global
    return window.Telegram?.WebApp?.initData || ''
  } catch {
    return ''
  }
}

export function hapticFeedback(type: 'light' | 'medium' | 'heavy' | 'success' | 'error') {
  try {
    // @ts-expect-error - Telegram WebApp global
    const haptic = window.Telegram?.WebApp?.HapticFeedback
    if (haptic) {
      if (type === 'success' || type === 'error') {
        haptic.notificationOccurred(type)
      } else {
        haptic.impactOccurred(type)
      }
    }
  } catch {
    // Not in Telegram
  }
}

export function openLink(url: string) {
  try {
    // @ts-expect-error - Telegram WebApp global
    window.Telegram?.WebApp?.openLink(url)
  } catch {
    window.open(url, '_blank')
  }
}
