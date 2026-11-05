export interface Paragraph {
  id: string
  text: string
}

export interface Document {
  id: string
  name: string
  kind: string
  contract_id: string
  status: string
  paragraphs: Paragraph[]
}
