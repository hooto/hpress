<script lang="ts">
  // File IDE (spec-editor). Route: spec-editor/<modname>. Ports spec-editor.js +
  // tablet.js + lc-editor.js core: file tree (mod-set-fs/list), open files in
  // tabs, CodeMirror edit (LcEditor), save with md5 sumcheck (mod-set-fs/put),
  // new / delete / rename. Open-file state is held in memory (the legacy's
  // IndexedDB offline cache is omitted; online operation is fully preserved).
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { confirmModal, openModal } from '../../lib/modal'
  import { navigate } from '../../lib/router'
  import { md5 } from '../../lib/util'
  import { uploadModFsFile, joinFsPath } from '../../lib/fsupload'
  import LcEditor from './LcEditor.svelte'
  import FsUpload from './FsUpload.svelte'
  import type { FsFile } from '../../lib/types'

  let { route = 'spec-editor/index' }: { route?: string } = $props()
  const modname = $derived(
    route.startsWith('spec-editor/') ? route.slice('spec-editor/'.length) : '',
  )

  interface OpenFile {
    path: string
    name: string
    content: string
    origSum: string
  }

  let treePath = $state('/')
  let treeItems: (FsFile & { _abspath: string })[] = $state([])
  // openFiles is reassigned and its rows are deep-mutated (origSum, and content
  // via the LcEditor bind), so it carries deep reactivity.
  let openFiles: OpenFile[] = $state([])
  let activePath = $state('')
  let menuOpen = $state(false)
  let menuWrap: HTMLElement
  let singleInput: HTMLInputElement
  let ctxMenu: { x: number; y: number; path: string; name: string } | null = $state(null)

  const active = $derived(openFiles.find((f) => f.path === activePath))
  const dirty = (f: { content: string; origSum: string }) => md5(f.content) !== f.origSum

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

  const breadcrumbs = $derived(treePath.split('/').filter(Boolean))

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
      // f is a row of the $state array, so this mutation is reactive on its own.
      if (rsp && rsp.kind === 'FsFile') {
        f.origSum = sumcheck
      }
    } catch (e) {
      if (e instanceof ApiError) alert(e.message)
    }
  }

  function closeMenu() {
    menuOpen = false
  }

  // Close the New dropdown on any click outside of it (the toggle button and
  // menu live inside menuWrap; their own handlers manage open/close). Also
  // dismiss the right-click context menu on any click outside its own entries.
  function onWindowClick(e: MouseEvent) {
    if (menuOpen && menuWrap && !menuWrap.contains(e.target as Node)) {
      menuOpen = false
    }
    if (ctxMenu && !(e.target as HTMLElement)?.closest('.lcide-ctxmenu')) {
      ctxMenu = null
    }
  }

  // Right-click context menu on a tree item (Rename / Delete). Position at the
  // cursor, clamped to the viewport so it never overflows the edge.
  function openCtx(e: MouseEvent, path: string, name: string) {
    const mw = 156
    const mh = 86
    const x = Math.max(4, Math.min(e.clientX, window.innerWidth - mw - 4))
    const y = Math.max(4, Math.min(e.clientY, window.innerHeight - mh - 4))
    ctxMenu = { x, y, path, name }
  }

  function ctxRename() {
    if (!ctxMenu) return
    const { path, name } = ctxMenu
    ctxMenu = null
    renameFile(path, name)
  }

  function ctxDelete() {
    if (!ctxMenu) return
    const { path, name } = ctxMenu
    ctxMenu = null
    deleteFile(path, name)
  }

  function rootRefresh() {
    closeMenu()
    loadTree('/')
  }

  function backToModules() {
    navigate('spec/index')
  }

  function newFile() {
    closeMenu()
    const name = prompt('New file name')
    if (!name) return
    const path = joinPath(treePath, name)
    api.post('mod-set-fs/put', { path, body: '', encode: 'text' }, { modname }).then(() => {
      loadTree()
      openFile(path, name)
    })
  }

  function newFolder() {
    closeMenu()
    const name = prompt('New folder name')
    if (!name) return
    const path = joinPath(treePath, name)
    api.post('mod-set-fs/put', { path, isdir: true }, { modname }).then(() => loadTree())
  }

  // "Upload Single File" — open the OS picker and upload straight into the
  // current tree directory (no modal), one file at a time.
  function pickSingleFile() {
    closeMenu()
    singleInput.click()
  }

  async function onSinglePicked() {
    if (!singleInput.files || !singleInput.files.length) return
    for (const file of Array.from(singleInput.files)) {
      const r = await uploadModFsFile(modname, joinFsPath(treePath, file.name), file)
      if (!r.ok) alert(r.msg)
    }
    singleInput.value = ''
    loadTree()
  }

  // "Drag & Drop Batch Upload" — dedicated drop-zone modal (recursive folder
  // traversal), uploads into the current tree directory.
  function uploadBatch() {
    closeMenu()
    openModal({
      title: 'Drag & Drop Batch Upload',
      width: 640,
      height: 'auto',
      body: FsUpload,
      props: { modname, path: treePath, onDone: () => loadTree() },
      buttons: [{ title: 'Close', click: () => {} }],
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

  // Load the tree once the module name is known (and again if it ever changes).
  $effect(() => {
    if (modname) loadTree('/')
  })
  onMount(() => loadTree('/'))
</script>

<svelte:window onclick={onWindowClick} />

<div class="lcide-wrap">
  <div class="lcide-header">
    <button class="lcide-back" type="button" onclick={backToModules}>
      <i class="bi bi-arrow-left-short"></i> Modules
    </button>
    <span class="lcide-header-mod">{modname}</span>
  </div>

  <div class="lcide">
    <input type="file" class="lcide-hidden" bind:this={singleInput} onchange={onSinglePicked} />

  {#if ctxMenu}
    <div class="lcide-ctxmenu" style={`top:${ctxMenu.y}px;left:${ctxMenu.x}px`}>
      <button onclick={ctxRename}><i class="bi bi-pencil"></i> Rename</button>
      <button class="text-danger" onclick={ctxDelete}><i class="bi bi-trash"></i> Delete</button>
    </div>
  {/if}

  <div class="lcide-fsnav lynkui-scroll">
    <div class="lcide-fsbar">
      <span class="lcide-fsbar-title">Files</span>
      <div class="lcide-fsbar-actions">
        <div class="lcide-menu-wrap" bind:this={menuWrap}>
          <button class="lcide-fsbtn" title="New" onclick={() => (menuOpen = !menuOpen)}>
            <i class="bi bi-plus-lg"></i>
          </button>
          {#if menuOpen}
            <ul class="lcide-menu">
              <li>
                <button onclick={newFile}><i class="bi bi-file-earmark"></i> New File</button>
              </li>
              <li>
                <button onclick={newFolder}><i class="bi bi-folder-plus"></i> New Folder</button>
              </li>
              <li class="lcide-menu-sep"></li>
              <li>
                <button onclick={pickSingleFile}><i class="bi bi-upload"></i> Upload Single File</button>
              </li>
              <li>
                <button onclick={uploadBatch}
                  ><i class="bi bi-cloud-arrow-up"></i> Drag &amp; Drop Batch Upload</button
                >
              </li>
            </ul>
          {/if}
        </div>
        <button class="lcide-fsbtn" title="Refresh" onclick={rootRefresh}>
          <i class="bi bi-arrow-repeat"></i>
        </button>
      </div>
    </div>
    <div class="lcide-breadcrumb">
      <button class="lcide-bc-root" title="Go to root" onclick={() => loadTree('/')}>
        <i class="bi bi-house-door"></i>
      </button>
      {#each breadcrumbs as crumb, i (crumb + i)}
        <span class="lcide-bc-sep">/</span>
        <button type="button" class="hp-link-btn" onclick={() => navTo(breadcrumbs.slice(0, i + 1).join('/'))}>{crumb}</button>
      {/each}
    </div>
    <ul class="lcide-tree">
      {#each treeItems as it (it._abspath)}
        <li>
          <button
            type="button"
            class="lcide-tree-link hp-link-btn"
            title={it.name || ''}
            onclick={() => (it.isdir ? loadTree(it._abspath) : openFile(it._abspath, it.name || ''))}
            oncontextmenu={(e) => { e.preventDefault(); openCtx(e, it._abspath, it.name || '') }}
          >
            <i class={(it.isdir ? 'bi bi-folder' : 'bi bi-file-earmark') + ' lcide-tree-icon'}></i>
            <span class="lcide-tree-name">{it.name}</span>
          </button>
        </li>
      {/each}
    </ul>
  </div>

  <div class="lcide-main">
    <div class="lcide-tabs">
      {#each openFiles as f (f.path)}
        <div class={'lcide-tab' + (f.path === activePath ? ' active' : '')}>
          <button type="button" class="hp-link-btn" onclick={() => switchTab(f.path)}>
            {f.name}{#if dirty(f)}<span class="lcide-dirty">*</span>{/if}
          </button>
          <button class="lcide-tab-close" onclick={() => closeTab(f.path)}>×</button>
        </div>
      {/each}
    </div>

    <div class="lcide-editor">
      {#if active}
        <LcEditor bind:value={active.content} path={active.path} onSave={() => save(active.path)} />
        <div class="lcide-savebar">
          <button class="btn btn-sm btn-primary" onclick={() => save(active.path)}>Save</button>
          {#if dirty(active)}<span class="text-warning">unsaved</span>{/if}
        </div>
      {:else}
        <div class="text-muted p-4">Select a file to edit.</div>
      {/if}
    </div>
  </div>
  </div>
</div>

<style>
  .lcide-wrap {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 120px);
  }
  .lcide-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 0 2px 8px 2px;
    flex-shrink: 0;
  }
  .lcide-back {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 10px;
    background: #fff;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 13px;
    color: #333;
    cursor: pointer;
  }
  .lcide-back:hover {
    background: #f1f1f1;
    border-color: #aaa;
  }
  .lcide-back i {
    font-size: 16px;
    line-height: 1;
  }
  .lcide-header-mod {
    font-size: 13px;
    font-weight: 600;
    color: #666;
  }
  .lcide {
    display: flex;
    flex: 1;
    min-height: 0;
    border: 1px solid #ddd;
  }
  .lcide-fsnav {
    width: 240px;
    border-right: 1px solid #ddd;
    overflow: auto;
    background: #fafafa;
    display: flex;
    flex-direction: column;
  }
  .lcide-hidden {
    display: none;
  }
  /* top bar — dark, matches the legacy /hp/mgr/ .lcx-fsnav */
  .lcide-fsbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 28px;
    flex-shrink: 0;
    background-color: rgba(0, 0, 0, 0.6);
    color: #f5f5f5;
  }
  .lcide-fsbar-title {
    padding-left: 10px;
    font-size: 14px;
    font-weight: bold;
    line-height: 28px;
  }
  .lcide-fsbar-actions {
    display: flex;
    align-items: center;
    height: 28px;
  }
  .lcide-fsbtn {
    width: 28px;
    height: 28px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    color: #f5f5f5;
    cursor: pointer;
  }
  .lcide-fsbtn:hover {
    background-color: rgba(0, 0, 0, 0.9);
  }
  .lcide-menu-wrap {
    position: relative;
    height: 28px;
  }
  .lcide-menu {
    position: absolute;
    top: 28px;
    right: 0;
    z-index: 100;
    min-width: 190px;
    margin: 0;
    padding: 3px;
    list-style: none;
    background-color: rgba(0, 0, 0, 0.9);
    text-align: left;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }
  .lcide-menu li button {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 5px 8px;
    background: transparent;
    border: none;
    color: #fff;
    font-size: 12px;
    text-align: left;
    cursor: pointer;
  }
  .lcide-menu li button:hover {
    background-color: rgba(255, 255, 255, 0.3);
  }
  .lcide-menu-sep {
    height: 1px;
    margin: 4px 2px;
    background-color: rgba(255, 255, 255, 0.2);
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
    padding: 0 8px 8px 8px;
    margin: 0;
    font-size: 13px;
  }
  .lcide-tree li {
    padding: 0;
  }
  /* tree item: icon column + wrappable name column, so a wrapped line indents
     past the icon instead of flowing back under it. No underline. */
  .lcide-tree-link {
    display: flex;
    align-items: flex-start;
    gap: 6px;
    padding: 3px 6px;
    border-radius: 3px;
    text-decoration: none;
    color: #1a1a1a;
  }
  .lcide-tree-link:hover {
    background-color: rgba(0, 0, 0, 0.06);
  }
  .lcide-tree-icon {
    flex: 0 0 auto;
    line-height: 19px;
  }
  .lcide-tree-name {
    min-width: 0;
    word-break: break-all;
    line-height: 19px;
  }
  /* right-click context menu (Rename / Delete) */
  .lcide-ctxmenu {
    position: fixed;
    z-index: 200;
    min-width: 150px;
    padding: 4px;
    background: #fff;
    border: 1px solid #ccc;
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }
  .lcide-ctxmenu button {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 5px 10px;
    background: transparent;
    border: none;
    border-radius: 3px;
    font-size: 13px;
    text-align: left;
    cursor: pointer;
    color: #1a1a1a;
  }
  .lcide-ctxmenu button:hover {
    background-color: #f1f1f1;
  }
  .lcide-breadcrumb {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 2px;
    padding: 8px 8px 4px 8px;
    font-size: 12px;
    margin-bottom: 4px;
  }
  .lcide-breadcrumb .hp-link-btn {
    color: #1a1a1a;
    text-decoration: none;
  }
  .lcide-breadcrumb .hp-link-btn:hover {
    color: #0d6efd;
    text-decoration: underline;
  }
  .lcide-bc-root {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    padding: 0;
    background: transparent;
    border: none;
    border-radius: 3px;
    color: #555;
    font-size: 13px;
    cursor: pointer;
  }
  .lcide-bc-root:hover {
    background-color: rgba(0, 0, 0, 0.06);
    color: #1a1a1a;
  }
  .lcide-bc-sep {
    color: #999;
    margin: 0 1px;
  }
</style>
