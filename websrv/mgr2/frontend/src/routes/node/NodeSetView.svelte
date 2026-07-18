<script lang="ts">
  // node set (create/edit) view. Ports node/set.tpl + node.js Set/SetCommit.
  // Field dispatch by type (string/text/int*), term options (tag input /
  // taxonomy select), extensions (comment_perentry/permalink/node_refer),
  // status, multi-language fields, main/side layout heuristic, Ctrl/Cmd+S.
  import { onMount, onDestroy } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { innerShow } from '../../lib/alert'
  import { nodeReferActive, hotkeyCtrlS } from '../../lib/store'
  import HpEditor from '../../lib/editor/HpEditor.svelte'
  import type { Node, NodeModel, SpecLangItem } from '../../lib/types'

  export let modname: string
  export let modelid: string
  export let nodeid: string | null = null
  export let oncancel: () => void = () => {}
  export let onsaved: () => void = () => {}

  const statusDef = [
    { type: 1, name: 'Publish' },
    { type: 2, name: 'Draft' },
    { type: 3, name: 'Private' },
  ]
  const onoff = [
    { type: 'true', name: 'ON' },
    { type: 'false', name: 'OFF' },
  ]
  const textFormats = [
    { name: 'text', value: 'Text' },
    { name: 'html', value: 'Html' },
    { name: 'shtml', value: 'Script Html' },
    { name: 'md', value: 'Makedown' },
  ]

  let loaded = false
  let langs: SpecLangItem[] = []
  let data: any = {} // the node being edited
  let model: NodeModel
  let fields: any[] = []
  let terms: any[] = []
  let termOptions: Record<string, any[]> = {} // termName -> flattened tree
  let status = 1
  let ext_comment_perentry: string = 'true'
  let ext_permalink_name = ''
  let ext_node_refer = ''

  $: ext = model?.extensions || ({} as any)

  function langList(csv?: string): SpecLangItem[] | null {
    if (!csv) return null
    const ids = csv.split(',')
    const list = ids
      .map((id) => langs.find((l) => l.id === id))
      .filter(Boolean) as SpecLangItem[]
    // legacy bug: checks `.lengh` (typo) so list is always kept when >=2 entries
    return list.length >= 2 ? list : null
  }

  async function load() {
    try {
      const langRsp = await api.get<{ items?: SpecLangItem[] }>('mod-set/spec-lang-list')
      langs = langRsp.items || []

      let node: any
      if (nodeid) {
        node = await api.get<Node>('node/entry', { modname, modelid, id: nodeid })
        if (!node || node.kind !== 'Node') {
          innerShow('hpm-node-alert', 'info', 'Item Not Found')
          return
        }
      } else {
        const m = await api.get<NodeModel>('node-model/entry', { modname, modelid })
        node = {
          kind: 'Node',
          model: m,
          id: '',
          title: '',
          status: 1,
          ext_comment_perentry: true,
          fields: [],
          terms: [],
          create_new: true,
        }
      }

      data = node
      model = node.model
      const mfields: any[] = (model.fields as any[]) || []
      const mterms: any[] = (model.terms as any[]) || []
      model.fields = mfields
      model.terms = mterms
      status = node.status || 1
      ext_comment_perentry = String(node.ext_comment_perentry ?? false)
      ext_permalink_name = node.ext_permalink_name || ''
      ext_node_refer = node.ext_node_refer || ''

      // build editable fields
      const out: any[] = []
      for (const field of mfields) {
        if (field.edit_disable) continue
        field.attrs = field.attrs || []
        for (const a of field.attrs) field['attr_' + a.key] = a.value

        const entry = (node.fields || []).find((f: any) => f.name === field.name)
        let value = entry?.value || ''
        let value_langs: any = entry?.langs || null

        field.attr_lang_list = langList(field.attr_langs)
        field.attr_lang_active = field.attr_lang_list ? field.attr_lang_list[0].id : null

        if (field.type === 'text') {
          // text attrs come from the entry (per-field), not the model
          for (const a of entry?.attrs || []) field['attr_' + a.key] = a.value
          field.attr_format = field.attr_format || 'md'
          field.attr_formats = field.attr_formats || 'text,html,md'
          const fmts = field.attr_formats.split(',')
          field._formats = textFormats.filter((f) => fmts.indexOf(f.name) > -1)
          if (field._formats.length < 1) field._formats = [textFormats[0]]
          if (!fmts.includes(field.attr_format)) field.attr_format = field._formats[0].name
          field._format = field.attr_format
        }

        field.value = value
        field.value_langs = field.attr_lang_active && !value_langs ? { items: [] } : value_langs
        field._display = value
        field._lang = field.attr_lang_active
        out.push(field)
      }
      fields = out

      // terms
      const tms: any[] = []
      for (const t of mterms) {
        const tv = (node.terms || []).find((x: any) => x.name === t.meta?.name)
        t.value = tv?.value || (t.type === 'taxonomy' ? '0' : '')
        tms.push(t)
        if (t.type === 'taxonomy') {
          const tl = await api.get<{ items?: any[] }>('term/list', {
            modname,
            modelid: t.meta!.name,
          })
          termOptions[t.meta!.name] = flattenTree(tl.items || [])
        }
      }
      terms = tms

      if ((!ext_node_refer || ext_node_refer.length < 12) && $nodeReferActive) {
        ext_node_refer = $nodeReferActive
      }

      innerShow('hpm-node-alert', '', '')
      loaded = true
      hotkeyCtrlS.set(() => commit({ save: true }))
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        innerShow('hpm-node-alert', 'danger', 'Failed to load')
      }
    }
  }

  function flattenTree(items: any[]): any[] {
    for (const it of items) if (!it.pid) it.pid = 0
    const out: any[] = []
    for (const it of items) {
      if (it.pid == 0) {
        out.push({ ...it, _dp: 0 })
        for (const sub of childrenOf(items, it.id, 1)) out.push(sub)
      }
    }
    return out
  }
  function childrenOf(items: any[], pid: string, dp: number): any[] {
    const out: any[] = []
    for (const it of items) {
      if (it.pid == pid) {
        out.push({ ...it, _dp: dp })
        out.push(...childrenOf(items, it.id, dp + 1))
      }
    }
    return out
  }

  // flush a field's displayed value into value / value_langs
  function flushField(field: any) {
    if (!field.attr_lang_list) {
      field.value = field._display
      return
    }
    const primary = field.attr_lang_list[0].id
    if (field._lang === primary) {
      field.value = field._display
    } else {
      field.value_langs = field.value_langs || { items: [] }
      const items = field.value_langs.items || []
      const found = items.find((x: any) => x.key === field._lang)
      if (found) found.value = field._display
      else items.push({ key: field._lang, value: field._display })
    }
  }
  function switchLang(field: any, lang: string) {
    if (!field.attr_lang_list || field._lang === lang) return
    flushField(field)
    field._lang = lang
    const primary = field.attr_lang_list[0].id
    if (lang === primary) {
      field._display = field.value || ''
    } else {
      const found = (field.value_langs?.items || []).find((x: any) => x.key === lang)
      field._display = found?.value || ''
    }
    field._display = field._display // trigger reactivity
  }

  $: sprintIndent = (n: number) => '    '.repeat(n)

  // main/side layout heuristic (node.js:790-834)
  $: layout = (() => {
    let mainLen = 0
    let sideLen = 0
    for (const f of fields) {
      if (f.type === 'string') mainLen += 1
      else if (f.type === 'text') mainLen += 5
      else sideLen += 1
    }
    sideLen += terms.length
    if (ext.comment_perentry) sideLen += 1
    if (ext.permalink) mainLen += 1
    if (ext.node_refer) mainLen += 1
    const useSide = sideLen > 0 && mainLen > sideLen
    return { useSide, target: useSide ? 'side' : 'main' }
  })()

  async function commit(opts: { save?: boolean } = {}) {
    // flush all fields
    for (const f of fields) flushField(f)

    const reqFields: any[] = []
    for (const f of fields) {
      const fs: any = { name: f.name, value: null, attrs: [] }
      if (f.type === 'text') {
        fs.attrs.push({ key: 'format', value: f._format || 'text' })
      }
      if (f.attr_lang_list && f._lang !== f.attr_lang_list[0].id) {
        fs.value = f.value
      } else {
        fs.value = f._display
      }
      if (f.value_langs) fs.langs = f.value_langs
      if (fs.value) reqFields.push(fs)
    }

    const reqTerms: any[] = []
    for (const t of terms) {
      if (t.value) reqTerms.push({ name: t.meta.name, value: String(t.value) })
    }

    const req: any = {
      id: data.id || '',
      status,
      fields: reqFields,
      terms: reqTerms,
      ext_comment_perentry: ext_comment_perentry === 'false' ? false : true,
      ext_permalink_name,
      ext_node_refer,
    }

    try {
      const rsp = await api.post<Node>('node/set', req, { modname, modelid })
      if (!rsp || rsp.kind !== 'Node') return
      data.id = rsp.id
      innerShow('hpm-node-alert', 'success', 'Successful operation')
      if (opts.save) return
      setTimeout(onsaved, 500)
    } catch (e) {
      if (e instanceof ApiError) innerShow('hpm-node-alert', 'danger', e.message)
    }
  }

  onDestroy(() => hotkeyCtrlS.set(null))
  onMount(load)
</script>

{#if loaded}
  <div class="hpm-block-gap-column">
    <div class="hpm-block-gap-row">
      <div class="hpm-nodeset-laymain" style={'width:' + (layout.useSide ? '75%' : '100%')}>
        <!-- title field -->
        {#each fields as f (f.name)}
          {#if f.name === 'title'}
            <div class="hpm-nodeset-tplx">
              <label class="form-label"><span>{f.title}</span></label>
              <input type="text" class="form-control" bind:value={f._display} />
            </div>
          {/if}
        {/each}

        {#if ext.node_refer}
          <div class="hpm-nodeset-tplx">
            <label class="form-label">Refer ID</label>
            <input type="text" class="form-control" bind:value={ext_node_refer} />
          </div>
        {/if}
        {#if ext.permalink}
          <div class="hpm-nodeset-tplx">
            <label class="form-label">Permalink Name</label>
            <input type="text" class="form-control" bind:value={ext_permalink_name} />
          </div>
        {/if}

        {#if layout.target === 'main'}
          {@render genericFields('main')}
        {/if}
      </div>

      {#if layout.useSide}
        <div class="hpm-nodeset-layside" style="width:25%">
          {@render genericFields('side')}
        </div>
      {/if}
    </div>

    <div class="hpm-block-gap-row-sm">
      <button class="btn btn-primary" on:click={() => commit()}>Save</button>
      <button class="btn btn-outline-primary" on:click={oncancel}>Cancel</button>
    </div>
  </div>
{/if}

{#snippet genericFields(col: string)}
  {#each fields as f (f.name)}
    {#if f.name === 'title'}
      <!-- title rendered above -->
    {:else if f.type === 'string'}
      <div class="hpm-nodeset-tplx">
        <label class="form-label">
          <span>{f.title}</span>
          {#if f.attr_lang_list}
            <select class="form-select form-select-sm d-inline-block w-auto" on:change={(e) => switchLang(f, e.currentTarget.value)}>
              {#each f.attr_lang_list as l}<option value={l.id} selected={f._lang === l.id}>{l.name}</option>{/each}
            </select>
          {/if}
        </label>
        <input type="text" class="form-control" bind:value={f._display} />
      </div>
    {:else if f.type === 'text'}
      <div class="hpm-nodeset-tplx">
        <label class="form-label">
          <span>{f.title}</span>
          {#if f.attr_lang_list}
            <select class="form-select form-select-sm d-inline-block w-auto" on:change={(e) => switchLang(f, e.currentTarget.value)}>
              {#each f.attr_lang_list as l}<option value={l.id} selected={f._lang === l.id}>{l.name}</option>{/each}
            </select>
          {/if}
        </label>
        <HpEditor
          bind:value={f._display}
          formats={f._formats.map((x: any) => x.name)}
          bind:format={f._format}
        />
      </div>
    {:else if f.type.startsWith('int') || f.type.startsWith('uint')}
      <div class="hpm-nodeset-tplx">
        <label class="form-label">{f.title}</label>
        <input type="text" class="form-control" bind:value={f._display} />
      </div>
    {/if}
  {/each}

  {#each terms as t (t.meta.name)}
    <div class="hpm-nodeset-tplx">
      <label class="form-label">{t.title}</label>
      {#if t.type === 'tag'}
        <input type="text" class="form-control" bind:value={t.value} />
      {:else}
        <select class="form-select" bind:value={t.value}>
          <option value="0">ROOT</option>
          {#each termOptions[t.meta.name] || [] as o (o.id)}
            <option value={o.id}>{sprintIndent(o._dp)}{o.title}</option>
          {/each}
        </select>
      {/if}
    </div>
  {/each}

  {#if ext.comment_enable && ext.comment_perentry}
    <div class="hpm-nodeset-tplx">
      <label class="form-label">Comment On/Off</label>
      <select class="form-select" bind:value={ext_comment_perentry}>
        {#each onoff as o}<option value={o.type} selected={ext_comment_perentry === o.type}>{o.name}</option>{/each}
      </select>
    </div>
  {/if}

  <div class="hpm-nodeset-tplx">
    <label class="form-label">Status</label>
    <select class="form-select" bind:value={status}>
      {#each statusDef as s}<option value={s.type} selected={status === s.type}>{s.name}</option>{/each}
    </select>
  </div>
{/snippet}

<style>
  .hpm-nodeset-tplx {
    margin-bottom: 12px;
  }
</style>
