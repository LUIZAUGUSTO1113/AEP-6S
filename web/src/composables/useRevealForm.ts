import { nextTick, useTemplateRef } from 'vue'

function scrollBehavior(): ScrollBehavior {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth'
}

export async function revealElement(
  root: HTMLElement | null | undefined,
  block: ScrollLogicalPosition = 'start',
) {
  await nextTick()
  if (!root) {
    return
  }
  root.scrollIntoView({
    block,
    inline: 'nearest',
    behavior: scrollBehavior(),
  })
  root
    .querySelector<HTMLElement>('input:not([type="hidden"]), select, textarea')
    ?.focus({ preventScroll: true })
}

export function useRevealForm() {
  const editor = useTemplateRef<HTMLElement>('editor')

  async function reveal() {
    await revealElement(editor.value)
  }

  return { reveal }
}
