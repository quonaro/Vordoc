const codeBlockRegex = /<pre><code[^>]*>.*?<\/code><\/pre>/gs

export function wrapCodeBlocksWithCopyButton(html: string): string {
  return html.replace(
    codeBlockRegex,
    (match) =>
      `<div class="vordoc-code-block relative group/copy">${match}<button type="button" class="vordoc-copy-btn absolute right-2 top-2 rounded border bg-card/90 px-2 py-1 text-xs text-muted-foreground opacity-0 shadow-sm transition-opacity group-hover/copy:opacity-100" aria-label="copy">Copy</button></div>`,
  )
}
