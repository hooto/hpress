<script lang="ts">
  // sys/status — read-only runtime/GC dashboard. Ports sys/status.tpl + the
  // Status handler in sys.js.
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { alertError } from '../../lib/alert'
  import { timeParseFormat, fmtResourceSize, fmtDuration } from '../../lib/util'
  import type { SysStatus } from '../../lib/types'

  let data: SysStatus | null = null
  let now = Date.now()

  onMount(async () => {
    try {
      data = await api.get<SysStatus>('sys/status')
      now = Date.now()
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        alertError('Error: Please try again later')
      }
    }
  })

  $: mem = data?.memstats || ({} as any)
  $: sinceGc = mem.last_gc ? fmtDuration(now - mem.last_gc / 1000000) : '0'
  $: avgGc =
    mem.pause_total_ns && mem.num_gc
      ? fmtDuration(mem.pause_total_ns / mem.num_gc, 1000000)
      : '0'
  $: totalGcPause = mem.pause_total_ns ? fmtDuration(mem.pause_total_ns, 1000000) : '0'
</script>

{#if data}
  <div class="card">
    <div class="card-header">System Monitor Status</div>
    <div class="card-body">
      <table width="100%" class="hp-sys-table">
        <tbody>
        <tr>
          <td width="30%">App Instance ID</td>
          <td>{data.instance_id}</td>
        </tr>
        <tr>
          <td>App Version - Release</td>
          <td>{data.app_version} - {data.app_release}</td>
        </tr>
        <tr>
          <td>Runtime Version</td>
          <td>{data.runtime_version}</td>
        </tr>
        <tr>
          <td>Uptime</td>
          <td>{timeParseFormat(data.uptime, 'Y-m-d H:i:s')}</td>
        </tr>

        <tr class="line">
          <td>Current Coroutine Number</td>
          <td>{data.coroutine_number}</td>
        </tr>
        <tr>
          <td>Current Memory Allocated</td>
          <td>{fmtResourceSize(mem.alloc)}</td>
        </tr>
        <tr>
          <td>Total Memory Allocated</td>
          <td>{fmtResourceSize(mem.total_alloc)}</td>
        </tr>
        <tr>
          <td>Memory obtained from system</td>
          <td>{fmtResourceSize(mem.sys)}</td>
        </tr>

        <tr class="line">
          <td>Next GC Recycle</td>
          <td>{fmtResourceSize(mem.next_gc)}</td>
        </tr>
        <tr>
          <td>Since Last GC Time</td>
          <td>{sinceGc}</td>
        </tr>
        <tr>
          <td>Total GC Pause</td>
          <td>{totalGcPause}</td>
        </tr>
        <tr>
          <td>Total GC Times</td>
          <td>{mem.num_gc}</td>
        </tr>
        <tr>
          <td>Average GC Pause</td>
          <td>{avgGc}</td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
{:else}
  <div class="text-muted p-3">loading</div>
{/if}

<style>
  .hp-sys-table {
    font-size: 10pt;
  }
  .hp-sys-table td {
    padding: 5px !important;
  }
  .hp-sys-table tr.line {
    border-top: 1px solid #ccc;
  }
</style>
