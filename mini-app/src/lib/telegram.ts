import {
  init,
  miniApp,
  themeParams,
  viewport,
  backButton,
  hapticFeedback as hapticFeedbackModule,
  openLink as sdkOpenLink,
  requestFullscreen,
} from '@telegram-apps/sdk-react'
import {
  retrieveRawInitData,
  mockTelegramEnv,
  emitEvent,
} from '@telegram-apps/bridge'

let initialized = false

// Check if running inside Telegram (sync check via URL params)
function isInTelegram(): boolean {
  try {
    const hash = window.location.hash
    const search = window.location.search
    return hash.includes('tgWebAppData') || search.includes('tgWebAppData')
  } catch {
    return false
  }
}

// Mock Telegram environment for local development
function setupDevMock() {
  const devUserId = import.meta.env.VITE_DEV_USER_ID
  if (!devUserId) return false

  if (isInTelegram()) return false // Already in Telegram

  mockTelegramEnv({
    launchParams: {
      tgWebAppData: new URLSearchParams([
        ['user', JSON.stringify({ id: Number(devUserId), first_name: 'Dev', username: 'devuser' })],
        ['hash', 'dev_mock_hash'],
        ['auth_date', Math.floor(Date.now() / 1000).toString()],
      ]),
      tgWebAppVersion: '8',
      tgWebAppPlatform: 'tdesktop',
      tgWebAppThemeParams: {},
    },
    onEvent(e) {
      if (e[0] === 'web_app_request_viewport') {
        emitEvent('viewport_changed', {
          height: window.innerHeight,
          width: window.innerWidth,
          is_expanded: true,
          is_state_stable: true,
        })
      }
    },
  })

  // Set dev safe area CSS variables to simulate fullscreen mode
  // Telegram's Close button overlay is ~56px on iOS
  const root = document.documentElement
  root.style.setProperty('--tg-safe-area-inset-top', '0px')
  root.style.setProperty('--tg-safe-area-inset-bottom', '0px')
  root.style.setProperty('--tg-content-safe-area-inset-top', '56px')
  root.style.setProperty('--tg-content-safe-area-inset-bottom', '0px')
  root.style.setProperty('--tg-total-safe-area-top', '56px')
  root.style.setProperty('--tg-total-safe-area-bottom', '0px')

  console.info('[Dev] Mocked Telegram environment for user:', devUserId)
  return true
}

// Bind Telegram safe area insets to CSS custom properties
// This handles both device safe areas (notch, Dynamic Island) and
// Telegram content safe areas (Close button overlay in fullscreen mode)
function bindSafeAreaCssVars() {
  const updateVars = () => {
    try {
      const device = viewport.safeAreaInsets?.() ?? { top: 0, bottom: 0, left: 0, right: 0 }
      const content = viewport.contentSafeAreaInsets?.() ?? { top: 0, bottom: 0, left: 0, right: 0 }

      const root = document.documentElement
      root.style.setProperty('--tg-safe-area-inset-top', `${device.top}px`)
      root.style.setProperty('--tg-safe-area-inset-bottom', `${device.bottom}px`)
      root.style.setProperty('--tg-content-safe-area-inset-top', `${content.top}px`)
      root.style.setProperty('--tg-content-safe-area-inset-bottom', `${content.bottom}px`)
      root.style.setProperty('--tg-total-safe-area-top', `${device.top + content.top}px`)
      root.style.setProperty('--tg-total-safe-area-bottom', `${device.bottom + content.bottom}px`)
    } catch {
      // Safe area APIs not available
    }
  }

  updateVars()
  // Poll for changes (fullscreen toggle, orientation changes)
  setInterval(updateVars, 100)
}

export async function initTelegramApp(): Promise<boolean> {
  if (initialized) return true

  try {
    // Setup dev mock before init (only in dev mode outside Telegram)
    setupDevMock()

    init()
    initialized = true

    // Mount viewport (async with guard)
    if (viewport.mount.isAvailable() && !viewport.isMounting()) {
      await viewport.mount()
      if (viewport.expand.isAvailable()) {
        viewport.expand()
      }
    }

    // Request fullscreen mode
    if (requestFullscreen.isAvailable()) {
      try {
        await requestFullscreen()
      } catch {
        // Fullscreen may fail on some platforms
      }
    }

    // Bind safe area insets to CSS custom properties
    bindSafeAreaCssVars()

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
    // Use retrieveRawInitData() from bridge - directly reads launch params
    // This is more reliable than initData.raw() which is a reactive signal
    return retrieveRawInitData() || ''
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
