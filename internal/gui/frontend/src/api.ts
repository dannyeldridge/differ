import type { Commit, FileChange, DiffLine } from './types'

interface GoAPI {
  CurrentBranch(): Promise<string>
  LoadCommits(): Promise<Commit[]>
  LoadFiles(hash: string): Promise<FileChange[]>
  LoadDiffLines(hash: string, file: string): Promise<DiffLine[]>
  LoadStagedFiles(): Promise<FileChange[]>
  LoadUnstagedFiles(): Promise<FileChange[]>
  LoadWorkingDiffLines(file: string, staged: boolean): Promise<DiffLine[]>
}

declare global {
  interface Window {
    go: { gui: { App: GoAPI } }
    runtime: {
      EventsOn(eventName: string, cb: (...data: unknown[]) => void): void
      Quit(): void
    }
  }
}

function app(): GoAPI {
  return window.go.gui.App
}

export const currentBranch = () => app().CurrentBranch()
export const loadCommits = () => app().LoadCommits()
export const loadFiles = (hash: string) => app().LoadFiles(hash).then((r) => r ?? [])
export const loadDiffLines = (hash: string, file: string) =>
  app().LoadDiffLines(hash, file).then((r) => r ?? [])
export const loadStagedFiles = () => app().LoadStagedFiles().then((r) => r ?? [])
export const loadUnstagedFiles = () => app().LoadUnstagedFiles().then((r) => r ?? [])
export const loadWorkingDiffLines = (file: string, staged: boolean) =>
  app().LoadWorkingDiffLines(file, staged).then((r) => r ?? [])
