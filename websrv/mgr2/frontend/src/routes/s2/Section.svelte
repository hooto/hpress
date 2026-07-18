<script lang="ts">
  // s2 object/file storage browser (top-level section). Ports s2.js + s2/index.tpl.
  // Fixed bucket /deft; breadcrumb dirnav, folder navigation, image thumbnails,
  // single + drag-drop upload (S2Upload modal), delete with native confirm.
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { openModal } from '../../lib/modal'
  import { s2ObjPathActive } from '../../lib/store'
  import { trim, timeParseFormat, fmtResourceSize, md5 } from '../../lib/util'
  import S2Upload from '../../lib/s2/S2Upload.svelte'
  import type { FsFile } from '../../lib/types'

  // the s2 browser is always at s2/index; route is accepted for shell uniformity
  export let route = 's2/index'

  const bucket = '/deft'

  type Item = FsFile & { _id: string; _abspath: string; _isimg: boolean; self_link?: string }
  let path = ''
  let items: Item[] = []
  let dirnav: { path: string; name: string }[] = []

  $: trim // referenced

  function normPath(p: string): string {
    p = (p || '').replace(/\/+/g, '/')
    if (p.indexOf(bucket) !== 0) p = bucket
    return p
  }

  async function load(p?: string) {
    if (p) {
      p = normPath(p)
    } else {
      p = normPath($s2ObjPathActive || bucket)
    }
    path = p
    s2ObjPathActive.set(p)
    try {
      const data = await api.get<{ items?: FsFile[] }>('s2-obj/list', { path: p })
      if (!data || !data.items) {
        items = []
      } else {
        const list: Item[] = []
        for (const it of data.items) {
          const name = it.name || ''
          const abspath = p + '/' + name
          const ext = name.lastIndexOf('.') > 0 ? name.toLowerCase().slice(name.lastIndexOf('.') + 1) : ''
          const isimg = ['jpg', 'jpeg', 'png', 'gif', 'svg'].indexOf(ext) >= 0
          list.push({
            ...it,
            _id: md5(abspath),
            _abspath: abspath,
            _isimg: isimg,
          })
        }
        items = list
      }
      // breadcrumb
      const trimmed = trim(p.replace(/\/+/g, '/'), '/')
      const segs = trimmed ? trimmed.split('/') : []
      let acc = ''
      dirnav = segs.map((s, i) => {
        acc += '/' + s
        return { path: acc, name: i === 0 ? 'Bucket: ' + s : s }
      })
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        items = []
      }
    }
  }

  function upload() {
    openModal({
      title: 'New File',
      width: 850,
      height: 550,
      body: S2Upload,
      props: { path, onDone: () => load(path) },
      buttons: [
        { title: 'Upload', class: 'btn-primary', click: () => {} },
        { title: 'Close', click: () => {} },
      ],
    })
  }

  async function del(abspath: string, id: string) {
    if (!window.confirm('This file will be deleted, Confirm?')) return
    try {
      const data = await api.get<FsFile>('s2-obj/del', { path: abspath })
      if (data && data.kind === 'FsFile') {
        items = items.filter((it) => it._id !== id)
      }
    } catch {
      /* ignore */
    }
  }

  onMount(() => load())
</script>

<div class="hpm-block-gap-column">
  <div class="d-flex flex-row justify-content-between hpm-block-gap-row">
    <div class="align-self-center hpm-breadcrumb">
      <ul class="hpm-breadcrumb-list">
        {#each dirnav as d (d.path)}
          <li>
            <a href="javascript:void(0)" on:click={() => load(d.path)}>{d.name}</a>
          </li>
        {/each}
      </ul>
    </div>
    <div class="hpm-node-nav hpm-nav-right">
      <button class="btn btn-primary" on:click={upload}>Upload New File</button>
    </div>
  </div>

  <div class="hpm-table-std">
    <table class="table table-hover align-middle">
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
                <a href="javascript:void(0)" on:click={() => load(v._abspath)}>{v.name}</a>
              {:else}
                <a href={v.self_link} target="_blank">{v.name}</a>
              {/if}
            </td>
            <td align="right">
              {#if !v.isdir}{fmtResourceSize(v.size || 0)}{/if}
            </td>
            <td align="right">{timeParseFormat(v.modtime, 'Y-m-d H:i:s')}</td>
            <td align="right">
              {#if !v.isdir}
                <button class="btn btn-outline-dark btn-sm" on:click={() => del(v._abspath, v._id)}
                  >Delete</button
                >
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
