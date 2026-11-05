import type { Contract } from '@/entities/contract/model/types'
import type { Document } from '@/entities/document/model/types'

export interface AnalysisResult {
  as_of: string
  contracts: Contract[]
  documents: Document[]
  warnings: string[]
}

export async function analyzeDocuments(files: File[], asOf: string): Promise<AnalysisResult> {
  const payload = new FormData()
  payload.append('as_of', asOf)
  files.forEach((file) => payload.append('files', file))

  const response = await fetch('/api/analyze', { method: 'POST', body: payload })
  const body: unknown = await response.json()
  if (!response.ok) {
    const message = typeof body === 'object' && body !== null && 'error' in body
      ? String(body.error)
      : 'Не удалось проверить документы'
    throw new Error(message)
  }
  return body as AnalysisResult
}
