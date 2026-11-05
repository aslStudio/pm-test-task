import { useState } from 'react'
import { ContractTable } from '@/entities/contract/ui/ContractTable'
import { SourceViewer } from '@/entities/document/ui/SourceViewer'
import { UploadForm } from '@/features/analyze-documents/ui/UploadForm'
import type { AnalysisResult } from '@/features/analyze-documents/api/analyzeDocuments'
import { DeadlineFilter } from '@/features/filter-deadlines/ui/DeadlineFilter'
import type { Evidence } from '@/shared/model/evidence'
import styles from './ContractCheckPage.module.scss'

export function ContractCheckPage() {
  const [result, setResult] = useState<AnalysisResult | null>(null)
  const [selectedEvidence, setSelectedEvidence] = useState<Evidence | null>(null)
  const sourceDocument = result?.documents.find((document) => document.id === selectedEvidence?.document_id)

  function acceptResult(next: AnalysisResult) {
    setResult(next)
    setSelectedEvidence(null)
  }

  return <main className={styles.page}>
    <header className={styles.header}>
      <div className={styles.eyebrow}>Semantic TextNavigator · прототип</div>
      <h1>Проверка условий поставки</h1>
      <p>Загрузите договоры и связанные документы. Каждое найденное условие можно проверить по исходному тексту.</p>
    </header>

    <UploadForm onResult={acceptResult} />

    {result && <>
      <ContractTable contracts={result.contracts} asOf={result.as_of} onEvidence={setSelectedEvidence} />
      <DeadlineFilter contracts={result.contracts} />
      {result.warnings.length > 0 && <section className={styles.warnings}>
        <h2>Предупреждения</h2>
        {result.warnings.map((warning, index) => <p key={index}>{warning}</p>)}
      </section>}
      {selectedEvidence && sourceDocument && <SourceViewer
        evidence={selectedEvidence}
        document={sourceDocument}
        onClose={() => setSelectedEvidence(null)}
      />}
    </>}
  </main>
}
