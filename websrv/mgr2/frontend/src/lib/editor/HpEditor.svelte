<script lang="ts">
  // hpEditor — CodeMirror 5 wrapper for a node `text` field. Ports editor.js.
  // Format tabs (text/html/shtml/md), markdown toolbar (preview/image), live
  // side-by-side preview with synchronized scroll, image insert via the s2
  // ObjSelector → ![FIG]({{hp_storage_service_endpoint}}/<path>?ipl=w960,h960).
  import { onMount, onDestroy } from 'svelte'
  import CodeMirror from 'codemirror'
  import 'codemirror/lib/codemirror.css'
  import 'codemirror/mode/markdown/markdown.js'
  import 'codemirror/mode/xml/xml.js'
  import 'codemirror/addon/selection/active-line.js'
  import { marked } from 'marked'
  import DOMPurify from 'dompurify'
  import { openModal } from '../modal'
  import { paths } from '../config'
  import ObjSelector from '../s2/ObjSelector.svelte'

  // value/format are two-way bound by the parent (bind:value / bind:format) so
  // they must be $bindable.
  let {
    value = $bindable(''),
    formats = ['md', 'html', 'shtml', 'text'],
    format = $bindable('md'),
  }: { value?: string; formats?: string[]; format?: string } = $props()

  let textarea: HTMLTextAreaElement
  let layoutEl: HTMLDivElement
  let previewEl = $state<HTMLDivElement | undefined>(undefined)
  // $state so the external-value-sync $effect re-runs when the editor instance
  // is ready (null → CodeMirror), not only on value changes.
  let cm = $state<any>(null)
  let showPreview = $state(false)
  let previewHtml = $state('')
  let resizeRo: ResizeObserver | undefined
  const s2_bucket_default = '/deft/'

  const isMd = $derived(format === 'md')

  function renderPreview() {
    if (!cm) return
    const html = marked.parse(cm.getValue()) as string
    previewHtml = DOMPurify.sanitize(html || '')
  }

  function applyFormat() {
    if (!cm) return
    const lineNumbers = format === 'md' || format === 'html' || format === 'shtml'
    // legacy keeps mode "markdown" for all formats; only lineNumbers toggles.
    cm.setOption('lineNumbers', lineNumbers)
    if (format !== 'md' && showPreview) {
      showPreview = false
      unbindScroll()
    }
  }

  onMount(() => {
    textarea.value = value
    const lineNumbers = format === 'md' || format === 'html' || format === 'shtml'
    cm = CodeMirror.fromTextArea(textarea, {
      mode: format === 'md' ? 'markdown' : 'xml',
      lineNumbers,
      theme: 'default',
      lineWrapping: true,
      styleActiveLine: true,
    })
    cm.setSize('100%', '100%')
    cm.on('change', () => {
      value = cm.getValue()
      if (showPreview) renderPreview()
    })
    // The editor lives in a flex column whose height is applied after mount
    // (viewport-driven). CodeMirror caches its measurements, so reflow it when
    // the container resizes, otherwise the scrollbar/last line misbehave.
    resizeRo = new ResizeObserver(() => {
      if (cm) cm.refresh()
    })
    resizeRo.observe(layoutEl)
  })

  onDestroy(() => {
    if (resizeRo) {
      resizeRo.disconnect()
      resizeRo = undefined
    }
    if (cm) {
      try {
        cm.toTextArea()
      } catch {
        /* ignore */
      }
      cm = null
    }
  })

  // external value updates (e.g. language switch) → push into the editor.
  // Tracks `value` (bindable prop); cm is $state so the first run waits for it.
  $effect(() => {
    if (cm && value !== cm.getValue()) {
      const cursor = cm.getCursor()
      cm.setValue(value || '')
      cm.setCursor(cursor)
    }
  })

  function setFormat(f: string) {
    format = f
    applyFormat()
  }

  function togglePreview() {
    showPreview = !showPreview
    if (showPreview) {
      renderPreview()
      setTimeout(bindScroll, 50)
    } else {
      unbindScroll()
    }
  }

  function insertImage() {
    if (!cm) return
    openModal({
      title: 'Select Images',
      width: 1000,
      height: 700,
      body: ObjSelector,
      props: {
        imageOnly: true,
        onselect: (p: string) => {
          if (!cm) return
          const path = p.replace(s2_bucket_default, '')
          let line = '\n![FIG](' + paths.storageEndpoint + '/' + path + '?ipl=w960,h960)\n\n'
          line = line.replace(/\/+/g, '/')
          const cs = cm.getCursor()
          cm.replaceRange(line, { line: cs.line, ch: cs.ch }, { line: cs.line, ch: cs.ch })
        },
      },
      buttons: [{ title: 'Cancel', click: () => {} }],
    })
  }

  // synchronized scroll (percentage-based, snapping at top/bottom)
  let scrollSource: 'editor' | 'preview' | null = null
  function onEditorScroll() {
    if (scrollSource === 'preview' || !cm || !previewEl) return
    scrollSource = 'editor'
    const scroller = cm.getScrollerElement ? cm.getScrollerElement() : null
    sync(scroller, previewEl)
    setTimeout(() => (scrollSource = null), 0)
  }
  function onPreviewScroll() {
    if (scrollSource === 'editor' || !previewEl) return
    scrollSource = 'preview'
    const scroller = cm && cm.getScrollerElement ? cm.getScrollerElement() : null
    sync(previewEl, scroller)
    setTimeout(() => (scrollSource = null), 0)
  }
  function sync(src: any, dst: any) {
    if (!src || !dst) return
    const scrollTop = src.scrollTop
    const height = src.clientHeight
    const scrollH = src.scrollHeight
    if (scrollTop === 0) {
      dst.scrollTop = 0
    } else if (scrollTop + height >= scrollH) {
      dst.scrollTop = dst.scrollHeight
    } else {
      const percent = scrollTop / scrollH
      dst.scrollTop = dst.scrollHeight * percent
    }
  }
  function bindScroll() {
    if (cm && cm.getScrollerElement) {
      CodeMirror.on(cm.getScrollerElement(), 'scroll', onEditorScroll)
    }
  }
  function unbindScroll() {
    if (cm && cm.getScrollerElement) {
      CodeMirror.off(cm.getScrollerElement(), 'scroll', onEditorScroll)
    }
  }
</script>

<div class="hpm-editor">
  <div class="hpm-editor-toolbar">
    <div class="btn-group btn-group-sm">
      {#each formats as f (f)}
        <button
          type="button"
          class={'btn btn-outline-dark editor-nav-' + f + (format === f ? ' active' : '')}
          onclick={() => setFormat(f)}
        >
          {f === 'md' ? 'Makedown' : f}
        </button>
      {/each}
    </div>
    {#if isMd}
      <div class="btn-group btn-group-sm ms-2">
        {#if !showPreview}
          <button type="button" class="btn btn-outline-dark preview_open" onclick={togglePreview}
            >Preview</button
          >
        {:else}
          <button type="button" class="btn btn-outline-dark preview_close" onclick={togglePreview}
            >Close Preview</button
          >
        {/if}
        <button type="button" class="btn btn-outline-dark" onclick={insertImage}>Image</button>
      </div>
    {/if}
  </div>

  <div class="hpm-editor-layout" bind:this={layoutEl}>
    <div class="hpm-editor-col" style={'width:' + (showPreview ? '50%' : '100%')}>
      <textarea bind:this={textarea} style="display:none"></textarea>
    </div>
    {#if showPreview}
      <div
        class="hpm-editor-preview lynkui-scroll"
        bind:this={previewEl}
        onscroll={onPreviewScroll}
      >
        {@html previewHtml}
      </div>
    {/if}
  </div>
</div>

<style>
  .hpm-editor {
    width: 100%;
    display: flex;
    flex-direction: column;
    /* fill the text-field column (basis 0 + grow); shrinkable so it never
       overflows its allotted space. A no-op outside a flex parent. */
    flex: 1 1 0;
    min-height: 0;
  }
  .hpm-editor-toolbar {
    display: flex;
    align-items: center;
    margin-bottom: 6px;
  }
  .hpm-editor-layout {
    display: flex;
    border: 1px solid #ddd;
    /* fills the available height and scrolls internally. The modest floor keeps
       a new/empty editor usable; it must stay below the typical available space
       so the editor shrinks to fit instead of overflowing onto the Save row. */
    flex: 1 1 0;
    min-height: 200px;
  }
  .hpm-editor-col {
    display: flex;
    min-height: 0;
  }
  .hpm-editor-col :global(.CodeMirror) {
    width: 100%;
    height: 100%;
  }
  .hpm-editor-preview {
    width: 50%;
    padding: 8px 12px;
    overflow: auto;
    border-left: 1px solid #ddd;
    background: #fafafa;
  }
</style>
