<script lang="ts">
  // S2 image picker modal body (s2/selector.tpl). Used by the node editor to
  // insert an image. Folders navigate; the Select button on an image invokes
  // onselect(abspath) and closes the modal. imageOnly filters to images.
  import { onMount } from 'svelte'
  import { api, ApiError } from '../api'
  import { closeModal } from '../modal'
  import { s2ObjPathActive } from '../store'
  import { trim, timeParseFormat, fmtResourceSize, md5 } from '../util'
  import type { FsFile } from '../types'

  let {
    onselect = () => {},
    imageOnly = true,
  }: { onselect?: (path: string) => void; imageOnly?: boolean } = $props()

  const bucket = '/deft'
  type Item = FsFile & { _id: string; _abspath: string; _isimg: boolean; self_link?: string }
  let path = ''
  let items: Item[] = $state([])
  let dirnav: { path: string; name: string }[] = $state([])

  function normPath(p: string): string {
    p = (p || '').replace(/\/+/g, '/')
    if (p.indexOf(bucket) !== 0) p = bucket
    return p
  }

  async function load(p?: string) {
    path = normPath(p || $s2ObjPathActive || bucket)
    s2ObjPathActive.set(path)
    try {
      const data = await api.get<{ items?: FsFile[] }>('s2-obj/list', { path })
      const list: Item[] = []
      for (const it of data?.items || []) {
        const name = it.name || ''
        const abspath = path + '/' + name
        const dot = name.lastIndexOf('.')
        const ext = dot > 0 ? name.toLowerCase().slice(dot + 1) : ''
        const isimg = ['jpg', 'jpeg', 'png', 'gif', 'svg'].indexOf(ext) >= 0
        if (imageOnly && !isimg && !it.isdir) continue
        list.push({ ...it, _id: md5(abspath), _abspath: abspath, _isimg: isimg })
      }
      items = list
      const trimmed = trim(path.replace(/\/+/g, '/'), '/')
      const segs = trimmed ? trimmed.split('/') : []
      let acc = ''
      dirnav = segs.map((s, i) => {
        acc += '/' + s
        return { path: acc, name: i === 0 ? 'Bucket: ' + s : s }
      })
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) items = []
    }
  }

  function select(abspath: string) {
    onselect(abspath)
    closeModal()
  }

  onMount(() => load())
</script>

<div id="hpm-s2-objls-navbar" class="hpm-breadcrumb">
  <ul class="hpm-breadcrumb-list">
    {#each dirnav as d (d.path)}
      <li><button type="button" class="hp-link-btn" onclick={() => load(d.path)}>{d.name}</button></li>
    {/each}
  </ul>
</div>

<table class="table table-hover">
  <thead>
    <tr>
      <th style="width:64px"></th>
      <th>Name</th>
      <th style="text-align:right">Size</th>
      <th></th>
      <th></th>
    </tr>
  </thead>
  <tbody>
    {#each items as v (v._id)}
      <tr id={'obj' + v._id}>
        <td>
          {#if v.isdir}
            <i class="bi bi-folder-fill" style="font-size:16px"></i>
          {:else if v._isimg}
            <a href={v.self_link} target="_blank"
              ><img src={(v.self_link || '') + '?ipl=w64,h64,c'} width="64" height="64" alt="" /></a
            >
          {/if}
        </td>
        <td class="ts3-fontmono">
          {#if v.isdir}
            <button type="button" class="hp-link-btn" onclick={() => load(v._abspath)}>{v.name}</button>
          {:else}
            <a href={v.self_link} target="_blank">{v.name}</a>
          {/if}
        </td>
        <td align="right">{#if !v.isdir}{fmtResourceSize(v.size || 0)}{/if}</td>
        <td align="right">{timeParseFormat(v.modtime, 'Y-m-d H:i:s')}</td>
        <td align="right">
          {#if !v.isdir}
            <button class="btn btn-outline-dark btn-sm" onclick={() => select(v._abspath)}
              >Select</button
            >
          {/if}
        </td>
      </tr>
    {/each}
  </tbody>
</table>
