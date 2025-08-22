export interface RequestOptions extends RequestInit {
  baseUrl?: string
}

export async function http<T>(path: string, { baseUrl = '/api', headers, ...init }: RequestOptions = {}): Promise<T> {
  const res = await fetch(`${baseUrl}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(headers || {}),
    },
    credentials: 'include',
  })
  if (!res.ok) {
    const msg = await res.text()
    throw new Error(msg || `HTTP ${res.status}`)
  }
  return (await res.json()) as T
}

