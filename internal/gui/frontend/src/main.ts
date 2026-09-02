interface Commit {
  Hash: string
  ShortHash: string
  Subject: string
  Author: string
  Date: string
  Body: string
}

interface GoAPI {
  CurrentBranch(): Promise<string>
  LoadCommits(): Promise<Commit[]>
}

declare global {
  interface Window {
    go: { gui: { App: GoAPI } }
  }
}

async function main() {
  const app = document.getElementById('app')!
  const branch = await window.go.gui.App.CurrentBranch()
  const commits = await window.go.gui.App.LoadCommits()
  app.innerHTML = `
    <p>branch: ${branch}</p>
    <ul>${commits.map((c) => `<li>${c.ShortHash} ${c.Subject}</li>`).join('')}</ul>
  `
}

main()
