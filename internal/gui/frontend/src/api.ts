import type { Commit, FileChange, DiffLine, RepoList } from './types'

interface GoAPI {
  ListRepos(): Promise<RepoList>
  CloseRepo(repo: string): Promise<number>
  CurrentBranch(repo: string): Promise<string>
  LoadCommits(repo: string): Promise<Commit[]>
  LoadFiles(repo: string, hash: string): Promise<FileChange[]>
  LoadDiffLines(repo: string, hash: string, file: string): Promise<DiffLine[]>
  LoadStagedFiles(repo: string): Promise<FileChange[]>
  LoadUnstagedFiles(repo: string): Promise<FileChange[]>
  LoadWorkingDiffLines(repo: string, file: string, staged: boolean): Promise<DiffLine[]>
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

export const listRepos = () => app().ListRepos().then((r) => ({ Repos: r.Repos ?? [], Active: r.Active }))
export const closeRepo = (repo: string) => app().CloseRepo(repo)
export const currentBranch = (repo: string) => app().CurrentBranch(repo)
export const loadCommits = (repo: string) => app().LoadCommits(repo)
export const loadFiles = (repo: string, hash: string) => app().LoadFiles(repo, hash).then((r) => r ?? [])
export const loadDiffLines = (repo: string, hash: string, file: string) =>
  app().LoadDiffLines(repo, hash, file).then((r) => r ?? [])
export const loadStagedFiles = (repo: string) => app().LoadStagedFiles(repo).then((r) => r ?? [])
export const loadUnstagedFiles = (repo: string) => app().LoadUnstagedFiles(repo).then((r) => r ?? [])
export const loadWorkingDiffLines = (repo: string, file: string, staged: boolean) =>
  app().LoadWorkingDiffLines(repo, file, staged).then((r) => r ?? [])
