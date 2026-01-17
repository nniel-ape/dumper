import { getInitData } from '@/lib/telegram'
import type { ApiError } from './types'

const BASE_URL = '/api'

// For dev mode without Telegram - set VITE_DEV_USER_ID in .env
const DEV_USER_ID = import.meta.env.VITE_DEV_USER_ID || null

class ApiClient {
  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }

    const initData = getInitData()
    if (initData) {
      headers['X-Telegram-Init-Data'] = initData
    }

    return headers
  }

  private getUrl(path: string): string {
    const url = new URL(BASE_URL + path, window.location.origin)

    // Add dev user_id if no init data
    if (DEV_USER_ID && !getInitData()) {
      url.searchParams.set('user_id', DEV_USER_ID)
    }

    return url.toString()
  }

  async get<T>(path: string): Promise<T> {
    const response = await fetch(this.getUrl(path), {
      method: 'GET',
      headers: this.getHeaders(),
    })

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.error)
    }

    return response.json()
  }

  async post<T, B = unknown>(path: string, body: B): Promise<T> {
    const response = await fetch(this.getUrl(path), {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(body),
    })

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.error)
    }

    return response.json()
  }

  async delete(path: string): Promise<void> {
    const response = await fetch(this.getUrl(path), {
      method: 'DELETE',
      headers: this.getHeaders(),
    })

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.error)
    }
  }

  getExportUrl(): string {
    return this.getUrl('/export')
  }
}

export const apiClient = new ApiClient()
