<script lang="ts">
  // File IDE (spec-editor). Route: spec-editor/<modname>. Ports spec-editor.js +
  // tablet.js + lc-editor.js core: file tree (mod-set-fs/list), open files in
  // tabs, CodeMirror edit (LcEditor), save with md5 sumcheck (mod-set-fs/put),
  // new / delete / rename. Open-file state is held in memory (the legacy's
  // IndexedDB offline cache is omitted; online operation is fully preserved).
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { confirmModal } from '../../lib/modal'
  import { md5 } from '../../lib/util'
  import LcEditor from './LcEditor.svelte'
  import type { FsFile } from '../../lib/types'

  export let route = 'spec-editor/index'
  $: modname = route.startsWith('spec-editor/') ? route.slice('spec-editor/'.length) : ''

  let treePath = '/'
  let treeItems: (FsFile & { _abspath: string })[] = []
  let openFiles: { path: string; name: string; content: string; origSum: string }[] = []
  let activePath = ''

  $: active = openFiles.find((f) => f.path === activePath)
  $: dirty = (f: { content: string; origSum: string }) => md5(f.content) !== f.origSum

  async function loadTree(p: string = treePath) {
    treePath = p
    try {
      const data = await api.get<{ items?: FsFile[] }>('mod-set-fs/list', { modname, path: p })
      treeItems = (data?.items || []).map((it) => ({ ...it, _abspath: joinPath(p, it.name || '') }))
    } catch {
      treeItems = []
    }
  }

  function joinPath(dir: string, name: string): string {
    return (dir.replace(/\/+$/, '') + '/' + name).replace(/\/+/g, '/')
  }

  $: breadcrumbs = treePath.split('/').filter(Boolean)

  function navTo(p: string) {
    loadTree(p === '' ? '/' : '/' + p)
  }

  async function openFile(path: string, name: string) {
    if (openFiles.find((f) => f.path === path)) {
      activePath = path
      return
    }
    try {
      const f = await api.get<FsFile>('mod-set-fs/get', { modname, path })
      const content = f?.body || ''
      openFiles = [...openFiles, { path, name, content, origSum: md5(content) }]
      activePath = path
    } catch (e) {
      if (e instanceof ApiError) alert(e.message)
    }
  }

  function switchTab(path: string) {
    activePath = path
  }

  function closeTab(path: string) {
    const f = openFiles.find((x) => x.path === path)
    if (!f) return
    const doClose = () => {
      openFiles = openFiles.filter((x) => x.path !== path)
      if (activePath === path) activePath = openFiles[openFiles.length - 1]?.path || ''
    }
    if (dirty(f)) {
      confirmModal({
        title: 'Save',
        html: '<div class="alert alert-warning">Save changes before closing?</div>',
        buttons: [
          { title: 'Save', class: 'btn-primary', click: () => save(path).then(doClose) },
          { title: 'Close without saving', class: 'btn-danger', click: doClose },
          { title: 'Cancel', click: () => {} },
        ],
      })
    } else {
      doClose()
    }
  }

  async function save(path: string) {
    const f = openFiles.find((x) => x.path === path)
    if (!f) return
    try {
      const sumcheck = md5(f.content)
      const rsp = await api.post<FsFile>('mod-set-fs/put', {
        path,
        body: f.content,
        encode: 'text',
        sumcheck,
      }, { modname })
      if (rsp && rsp.kind === 'FsFile') {
        f.origSum = sumcheck
        openFiles = openFiles
      }
    } catch (e) {
      if (e instanceof ApiError) alert(e.message)
    }
  }

  function newFile() {
    const name = prompt('New file name')
    if (!name) return
    const path = joinPath(treePath, name)
    api.post('mod-set-fs/put', { path, body: '', encode: 'text' }, { modname }).then(() => {
      loadTree()
      openFile(path, name)
    })
  }

  function deleteFile(path: string, name: string) {
    if (!window.confirm('Delete ' + name + ' ?')) return
    api.post('mod-set-fs/del', { path }, { modname }).then(() => {
      loadTree()
      openFiles = openFiles.filter((x) => x.path !== path)
      if (activePath === path) activePath = openFiles[openFiles.length - 1]?.path || ''
    })
  }

  function renameFile(path: string, name: string) {
    const nn = prompt('Rename ' + name, name)
    if (!nn || nn === name) return
    const newPath = joinPath(treePath, nn)
    api.post('mod-set-fs/rename', { path, pathset: newPath }, { modname }).then(() => loadTree())
  }

  $: if (modname) loadTree('/')
  onMount(() => loadTree('/'))
</script>

<div class="lcide">
  <div class="lcide-fsnav lynkui-scroll">
    <div class="lcide-toolbar">
      <button class="btn btn-sm btn-primary" on:click={newFile}>New File</button>
    </div>
    <div class="lcide-breadcrumb">
      <a href="javascript:void(0)" on:click={() => loadTree('/')}>root</a>
      {#each breadcrumbs as crumb, i}
        <span>/</span>
        <a href="javascript:void(0)" on:click={() => navTo(breadcrumbs.slice(0, i + 1).join('/'))}>{crumb}</a>
      {/each}
    </div>
    <ul class="lcide-tree">
      {#each treeItems as it (it._abspath)}
        <li>
          {#if it.isdir}
            <a href="javascript:void(0)" on:click={() => loadTree(it._abspath)}>
              <i class="bi bi-folder"></i> {it.name}
            </a>
          {:else}
            <span class="lcide-file">
              <a href="javascript:void(0)" on:click={() => openFile(it._abspath, it.name || '')}>
                <i class="bi bi-file-earmark"></i> {it.name}
              </a>
              <span class="lcide-file-actions">
                <button class="btn btn-sm btn-link py-0" on:click={() => renameFile(it._abspath, it.name || '')}
                  >ren</button
                >
                <button class="btn btn-sm btn-link py-0 text-danger" on:click={() => deleteFile(it._abspath, it.name || '')}
                  >del</button
                >
              </span>
            </span>
          {/if}
        </li>
      {/each}
    </ul>
  </div>

  <div class="lcide-main">
    <div class="lcide-tabs">
      {#each openFiles as f (f.path)}
        <div class={'lcide-tab' + (f.path === activePath ? ' active' : '')}>
          <a href="javascript:void(0)" on:click={() => switchTab(f.path)}>
            {f.name}{#if dirty(f)}<span class="lcide-dirty">*</span>{/if}
          </a>
          <button class="lcide-tab-close" on:click={() => closeTab(f.path)}>×</button>
        </div>
      {/each}
    </div>

    <div class="lcide-editor">
      {#if active}
        <LcEditor bind:value={active.content} path={active.path} onSave={() => save(active.path)} />
        <div class="lcide-savebar">
          <button class="btn btn-sm btn-primary" on:click={() => save(active.path)}>Save</button>
          {#if dirty(active)}<span class="text-warning">unsaved</span>{/if}
        </div>
      {:else}
        <div class="text-muted p-4">Select a file to edit.</div>
      {/if}
    </div>
  </div>
</div>

<style>
  .lcide {
    display: flex;
    height: calc(100vh - 120px);
    border: 1px solid #ddd;
  }
  .lcide-fsnav {
    width: 240px;
    border-right: 1px solid #ddd;
    overflow: auto;
    padding: 8px;
    background: #fafafa;
  }
  .lcide-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .lcide-tabs {
    display: flex;
    border-bottom: 1px solid #ddd;
    background: #f1f1f1;
    overflow-x: auto;
  }
  .lcide-tab {
    display: flex;
    align-items: center;
    padding: 6px 10px;
    border-right: 1px solid #ddd;
    white-space: nowrap;
    font-size: 13px;
  }
  .lcide-tab.active {
    background: #fff;
  }
  .lcide-tab-close {
    border: none;
    background: none;
    margin-left: 6px;
    cursor: pointer;
    color: #888;
  }
  .lcide-dirty {
    color: #d33;
    margin-left: 2px;
  }
  .lcide-editor {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  .lcide-editor :global(.CodeMirror) {
    flex: 1;
  }
  .lcide-savebar {
    padding: 4px 8px;
    border-top: 1px solid #ddd;
    background: #fafafa;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .lcide-tree {
    list-style: none;
    padding-left: 0;
    font-size: 13px;
  }
  .lcide-tree li {
    padding: 2px 0;
  }
  .lcide-file {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .lcide-file-actions {
    visibility: hidden;
  }
  .lcide-file:hover .lcide-file-actions {
    visibility: visible;
  }
  .lcide-breadcrumb {
    font-size: 12px;
    margin-bottom: 8px;
  }
  .lcide-toolbar {
    margin-bottom: 8px;
  }
</style>
