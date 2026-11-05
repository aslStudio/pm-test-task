import { useEffect, useRef } from 'react'
import type { Evidence } from '@/shared/model/evidence'
import type { Document } from '../model/types'
import styles from './SourceViewer.module.scss'

interface Props {
  evidence: Evidence
  document: Document
  onClose: () => void
}

export function SourceViewer({ evidence, document, onClose }: Props) {
  const sectionRef = useRef<HTMLElement>(null)
  const selectedRef = useRef<HTMLParagraphElement>(null)

  useEffect(() => {
    sectionRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    selectedRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, [evidence])

  return <section className={styles.section} ref={sectionRef}>
    <div className={styles.heading}>
      <h2>Источник: {document.name}</h2>
      <button type="button" onClick={onClose}>Закрыть</button>
    </div>
    <p className={styles.hint}>Выделен абзац, из которого взято значение.</p>
    <div className={styles.paragraphs}>
      {document.paragraphs.map((paragraph) => <p
        key={paragraph.id}
        ref={paragraph.id === evidence.paragraph_id ? selectedRef : undefined}
        className={paragraph.id === evidence.paragraph_id ? styles.selected : undefined}
      ><span>{paragraph.id}</span>{paragraph.text}</p>)}
    </div>
  </section>
}
