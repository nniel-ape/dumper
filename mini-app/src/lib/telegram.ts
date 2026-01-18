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

  console.info('[Dev] Mocked Telegram environment for user:', devUserId)
  return true
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
