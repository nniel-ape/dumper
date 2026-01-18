import {
  init,
  miniApp,
  themeParams,
  viewport,
  backButton,
  initData,
  hapticFeedback as hapticFeedbackModule,
  openLink as sdkOpenLink,
} from '@telegram-apps/sdk-react'

let initialized = false

export async function initTelegramApp(): Promise<boolean> {
  if (initialized) return true

  try {
    init()
    initialized = true

    // Mount viewport (async with guard)
    if (viewport.mount.isAvailable() && !viewport.isMounting()) {
      await viewport.mount()
      if (viewport.expand.isAvailable()) {
        viewport.expand()
      }
    }

    // Mount theme params (sync version available)
    if (themeParams.mountSync.isAvailable() && !themeParams.isMounted()) {
      themeParams.mountSync()
    }

    // Mount mini app
    if (miniApp.mount.isAvailable() && !miniApp.isMounting()) {
      await miniApp.mount()
      miniApp.ready()
    }

    // Mount back button
    if (backButton.mount.isAvailable() && !backButton.isMounted()) {
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
    return initData.raw() || ''
  } catch {
    return ''
  }
}

export function hapticFeedback(type: 'light' | 'medium' | 'heavy' | 'success' | 'error') {
  try {
    if (type === 'success' || type === 'error') {
      if (hapticFeedbackModule.notificationOccurred.isAvailable()) {
        hapticFeedbackModule.notificationOccurred(type)
      }
    } else {
      if (hapticFeedbackModule.impactOccurred.isAvailable()) {
        hapticFeedbackModule.impactOccurred(type)
      }
    }
  } catch {
    // Not in Telegram
  }
}

export function openLink(url: string) {
  try {
    if (sdkOpenLink.isAvailable()) {
      sdkOpenLink(url)
      return
    }
  } catch {
    // Fall through to window.open
  }
  window.open(url, '_blank')
}

export { backButton, themeParams }
