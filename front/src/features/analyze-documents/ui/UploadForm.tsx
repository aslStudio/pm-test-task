import { useState, type FormEvent } from 'react'
import { analyzeDocuments, type AnalysisResult } from '../api/analyzeDocuments'
import styles from './UploadForm.module.scss'

interface Props {
  onResult: (result: AnalysisResult) => void
}

export function UploadForm({ onResult }: Props) {
  const [files, setFiles] = useState<File[]>([])
  const [asOf, setAsOf] = useState('2026-03-10')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setBusy(true)
    setError('')
    try {
      onResult(await analyzeDocuments(files, asOf))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Не удалось проверить документы')
    } finally {
      setBusy(false)
    }
  }

  return <>
    <form className={styles.form} onSubmit={submit}>
      <label className={styles.field}>
        TXT-документы
        <input type="file" multiple accept=".txt,text/plain" required onChange={(event) => setFiles(Array.from(event.target.files ?? []))} />
      </label>
      <label className={styles.field}>
        Дата проверки
        <input type="date" required value={asOf} onChange={(event) => setAsOf(event.target.value)} />
      </label>
      <button className={styles.submit} disabled={busy || files.length === 0} type="submit">
        {busy ? 'Проверяем…' : 'Проверить документы'}
      </button>
      <span className={styles.count}>Выбрано файлов: {files.length}</span>
    </form>
    {error && <p className={styles.error} role="alert">{error}</p>}
  </>
}
