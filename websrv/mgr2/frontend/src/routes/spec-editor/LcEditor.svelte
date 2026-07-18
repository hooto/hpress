<script lang="ts">
  // lc-editor — CodeMirror 5 code editor for the file IDE. Ports lc-editor.js
  // (mode-by-extension dispatch, monokai theme, line numbers, fold gutter,
  // auto-close brackets/tags). binds Ctrl/Cmd-S to the parent save handler.
  import { onMount, onDestroy } from 'svelte'
  import CodeMirror from 'codemirror'
  import 'codemirror/lib/codemirror.css'
  import 'codemirror/theme/monokai.css'
  import 'codemirror/mode/javascript/javascript.js'
  import 'codemirror/mode/xml/xml.js'
  import 'codemirror/mode/css/css.js'
  import 'codemirror/mode/htmlmixed/htmlmixed.js'
  import 'codemirror/mode/clike/clike.js'
  import 'codemirror/mode/go/go.js'
  import 'codemirror/mode/php/php.js'
  import 'codemirror/mode/python/python.js'
  import 'codemirror/mode/yaml/yaml.js'
  import 'codemirror/mode/sql/sql.js'
  import 'codemirror/mode/lua/lua.js'
  import 'codemirror/mode/markdown/markdown.js'
  import 'codemirror/addon/edit/closebrackets.js'
  import 'codemirror/addon/edit/closetag.js'
  import 'codemirror/addon/fold/foldcode.js'
  import 'codemirror/addon/fold/foldgutter.js'
  import 'codemirror/addon/fold/foldgutter.css'

  export let value = ''
  export let path = ''
  export let onSave: () => void = () => {}

  let textarea: HTMLTextAreaElement
  let cm: any = null

  function modeFor(name: string): string {
    const ext = (name.split('.').pop() || '').toLowerCase()
    const map: Record<string, string> = {
      js: 'javascript', json: 'javascript',
      xml: 'xml', html: 'htmlmixed', htm: 'htmlmixed', tpl: 'htmlmixed',
      css: 'css',
      md: 'markdown',
      c: 'clike', h: 'clike', cpp: 'clike', java: 'clike',
      go: 'go', php: 'php', py: 'python', rb: 'ruby',
      yml: 'yaml', yaml: 'yaml', sql: 'sql', lua: 'lua', sh: 'shell',
    }
    return map[ext] || 'htmlmixed'
  }

  onMount(() => {
    textarea.value = value
    cm = CodeMirror.fromTextArea(textarea, {
      mode: modeFor(path),
      theme: 'monokai',
      lineNumbers: true,
      lineWrapping: true,
      styleActiveLine: true,
      matchBrackets: true,
      indentUnit: 4,
      tabSize: 4,
      smartIndent: true,
      autoCloseBrackets: true,
      autoCloseTags: true,
      foldGutter: true,
      gutters: ['CodeMirror-linenumbers', 'CodeMirror-foldgutter'],
    })
    cm.setSize('100%', '100%')
    cm.on('change', () => {
      value = cm.getValue()
    })
    cm.setOption('extraKeys', {
      'Ctrl-S': () => onSave(),
      'Cmd-S': () => onSave(),
      Tab: (cmm: any) => {
        if (cmm.somethingSelected()) cmm.indentSelection('add')
        else cmm.replaceSelection('    ', 'end')
      },
    })
  })

  onDestroy(() => {
    if (cm) {
      try {
        cm.toTextArea()
      } catch {
        /* ignore */
      }
      cm = null
    }
  })

  // external value update (switching file content) → reload into the editor
  let lastPath = path
  $: if (cm) {
    if (path !== lastPath) {
      lastPath = path
      cm.setOption('mode', modeFor(path))
      const cursor = cm.getCursor()
      cm.setValue(value || '')
      cm.setCursor(cursor)
    } else if (value !== cm.getValue()) {
      const cursor = cm.getCursor()
      cm.setValue(value || '')
      cm.setCursor(cursor)
    }
  }
</script>

<textarea bind:this={textarea} style="display:none" />

<style>
  :global(.CodeMirror) {
    height: 100% !important;
    font-size: 13px;
  }
</style>
