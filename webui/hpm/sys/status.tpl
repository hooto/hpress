<style>
  .page-header {
    margin: 10px 0;
  }
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

<div class="card">
  <div class="card-header">System Monitor Status</div>
  <div class="card-body">
    <table width="100%" class="hp-sys-table">
      <tr>
        <td width="30%">App Instance ID</td>
        <td class="">{[=it.instance_id]}</td>
      </tr>
      <tr>
        <td>App Version - Release</td>
        <td class="">{[=it.app_version]} - {[=it.app_release]}</td>
      </tr>
      <tr>
        <td>Runtime Version</td>
        <td class="">{[=it.runtime_version]}</td>
      </tr>
      <tr>
        <td>Uptime</td>
        <td>{[=lynkui.utilx.timeParseFormat(it.uptime, "Y-m-d H:i:s")]}</td>
      </tr>

      <tr class="line">
        <td>Current Coroutine Number</td>
        <td>{[=it.coroutine_number]}</td>
      </tr>
      <tr>
        <td>Current Memory Allocated</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.memstats.alloc)]}</td>
      </tr>
      <tr>
        <td>Total Memory Allocated</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.memstats.total_alloc)]}</td>
      </tr>
      <tr>
        <td>Memory obtained from system</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.memstats.sys)]}</td>
      </tr>

      <tr class="line">
        <td>Next GC Recycle</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.memstats.next_gc)]}</td>
      </tr>
      <tr>
        <td>Since Last GC Time</td>
        <td>
          {[=hpSys.UtilDurationFormat((new Date().getTime()) -
          (it.memstats.last_gc / 1000000))]}
        </td>
      </tr>
      <tr>
        <td>Total GC Pause</td>
        <td>
          {[=hpSys.UtilDurationFormat(it.memstats.pause_total_ns, 1000000)]}
        </td>
      </tr>
      <tr>
        <td>Total GC Times</td>
        <td>{[=it.memstats.num_gc]}</td>
      </tr>
      <tr>
        <td>Average GC Pause</td>
        <td>
          {[=hpSys.UtilDurationFormat((it.memstats.pause_total_ns /
          it.memstats.num_gc), 1000000)]}
        </td>
      </tr>

      <!-- <tr class="line">
        <td>CpuNum</td>
        <td>{[=it.info.cpu_num]}</td>
      </tr>
      <tr>
        <td>Uptime</td>
        <td>{[=hpSys.UtilDurationFormat(it.info.uptime * 1000)]}</td>
      </tr>
      <tr>
        <td>Loads</td>
        <td>{[=it.info.loads[0]]}</td>
      </tr>
      <tr>
        <td>MemTotal</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.info.mem_total)]}</td>
      </tr>
      <tr>
        <td>MemFree</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.info.mem_free)]}</td>
      </tr>
      <tr>
        <td>MemShared</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.info.mem_shared)]}</td>
      </tr>
      <tr>
        <td>MemBuffer</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.info.mem_buffer)]}</td>
      </tr>
      <tr>
        <td>MemUsed</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.info.mem_used)]}</td>
      </tr>
      <tr>
        <td>SwapTotal</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.info.swap_total)]}</td>
      </tr>
      <tr>
        <td>SwapFree</td>
        <td>{[=hpSys.UtilResourceSizeFormat(it.info.swap_free)]}</td>
      </tr>
      <tr>
        <td>Procs</td>
        <td>{[=it.info.procs]}</td>
      </tr>   -->
    </table>
  </div>
</div>
