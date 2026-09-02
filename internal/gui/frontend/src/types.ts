export interface Commit {
  Hash: string
  ShortHash: string
  Subject: string
  Author: string
  Date: string
  Body: string
}

export interface FileChange {
  Status: string
  Path: string
}

export type DiffLineType = 'header' | 'hunk' | 'add' | 'del' | 'context'

export interface DiffLine {
  Type: DiffLineType
  Content: string
}
