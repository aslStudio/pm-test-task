import type { ContractField as Field, Evidence } from '../model/types'
import styles from './ContractTable.module.scss'

interface Props {
  field: Field
  currency?: boolean
  onEvidence: (evidence: Evidence) => void
}

export function ContractField({ field, currency = false, onEvidence }: Props) {
  if (!field.value) return <span className={styles.missing}>нет данных</span>

  const display = currency ? `${Number(field.value).toLocaleString('ru-RU')} ₽` : field.value
  return <button className={styles.value} type="button" onClick={() => field.evidence && onEvidence(field.evidence)}>
    {display}
    {field.changed && <small>изменено соглашением</small>}
  </button>
}
