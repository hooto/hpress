<script lang="ts">
  // sys/iam-status — Local vs IAM-registered app info + permissions. Ports
  // sys/iam-status.tpl + IamStatus/IamSync in sys.js. Roles are hardcoded
  // client-side (hpSys.roles). Sync POSTs iam-sync (no body) and refreshes.
  import { onMount } from 'svelte'
  import { api, ApiError } from '../../lib/api'
  import { innerShow } from '../../lib/alert'
  import { flashThen } from '../../lib/feedback'
  import Alert from '../../lib/Alert.svelte'
  import type { SysIamStatus, SysIamInstance } from '../../lib/types'

  const roles = [
    { idxid: 'sa', name: 'Sys Admin' },
    { idxid: 'user', name: 'User' },
    { idxid: 'guest', name: 'Guest' },
  ]

  let data = $state<SysIamStatus | null>(null)
  let syncing = $state(false)
  const alertId = 'hp-mgr-sys-iam-alert'

  onMount(load)

  async function load() {
    try {
      const d = await api.get<SysIamStatus>('sys/iam-status')
      d.instance_self = norm(d.instance_self)
      d.instance_registered = d.instance_registered ? norm(d.instance_registered) : (undefined as any)
      data = d
    } catch (e) {
      if (!(e instanceof ApiError && e.code === 'Unauthorized')) {
        alert('Error: Please try again later')
      }
    }
  }

  function norm(inst: SysIamInstance | undefined): SysIamInstance {
    if (!inst) return {}
    const permissions = (inst.permissions || []).map((p: any) => ({
      ...p,
      roles: p.roles || [],
    }))
    return { ...inst, permissions }
  }

  function roleNames(roleIds: string[] | undefined): string[] {
    if (!roleIds || roleIds.length === 0) return ['Owner']
    const out: string[] = []
    for (const rv of roleIds) {
      const drv = roles.find((r) => r.idxid === rv)
      if (drv) out.push(drv.name)
    }
    return out
  }

  async function sync() {
    if (syncing) return
    syncing = true
    try {
      await api.post('sys/iam-sync')
      flashThen(alertId, 'success', 'Successful registered', load, 1000)
    } catch (e) {
      if (e instanceof ApiError) {
        innerShow(alertId, 'danger', e.message || 'Network Connection Exception')
      }
    } finally {
      syncing = false
    }
  }
</script>

{#if data}
  <div class="card">
    <div class="card-header">IAM Service Status</div>
    <div class="card-body">
      <table width="100%" class="table hpm-table-cols">
        <tbody>
        <tr>
          <td width="20%">IAM Base URL</td>
          <td>{data.base_url}</td>
        </tr>
        <tr>
          <td width="20%">App ID</td>
          <td>{data.app_id}</td>
        </tr>
        <tr>
          <td width="20%">Secret Key</td>
          <td>{data.secret_key}</td>
        </tr>
        </tbody>
      </table>


      <form id="hp-mgr-sys-iam" onsubmit={(e) => e.preventDefault()}>
        <table width="100%" class="table hpm-table-cols hpm-table-middle">
          <thead>
            <tr>
              <th style="width:20%"></th>
              <th style="width:40%"><strong>Local App Info</strong></th>
              <th><strong>Registered in IAM Service</strong></th>
            </tr>
          </thead>
          <tbody>
          <tr>
            <td>App Version</td>
            <td>{data.instance_self?.version}</td>
            <td>{#if data.instance_registered}{data.instance_registered.version}{/if}</td>
          </tr>
          <tr>
            <td>App ID</td>
            <td>{data.instance_self?.id}</td>
            <td>{#if data.instance_registered}{data.instance_registered.id}{/if}</td>
          </tr>
          <tr>
            <td>App Name</td>
            <td>
              <input
                type="text"
                class="form-control input-sm"
                placeholder="Enter the App Name"
                value={data.instance_self?.name || ''}
              />
            </td>
            <td>{#if data.instance_registered}{data.instance_registered.name}{/if}</td>
          </tr>
          <tr>
            <td>Entry URL</td>
            <td>
              <input
                type="text"
                class="form-control input-sm"
                placeholder="Enter the Entry URL of App Instance"
                value={data.instance_self?.url || ''}
              />
            </td>
            <td>{#if data.instance_registered}{data.instance_registered.url}{/if}</td>
          </tr>
          <tr>
            <td>Permissions</td>
            <td>
              <table class="table hpm-table-cols hpm-table-middle hpm-table-clean">
                <thead>
                  <tr><th>Permission</th><th>Roles</th></tr>
                </thead>
                <tbody>
                  {#each data.instance_self?.permissions || [] as v (v.permission)}
                    <tr>
                      <td>
                        {v.permission}
                      </td>
                      <td>{roleNames(v.roles).join(', ')}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </td>
            <td>
              {#if data.instance_registered}
                <table class="table hpm-table-cols hpm-table-middle hpm-table-clean">
                  <thead>
                    <tr><th>Permission</th><th>Roles</th></tr>
                  </thead>
                  <tbody>
                    {#each data.instance_registered.permissions || [] as v (v.permission)}
                      <tr>
                        <td>
                          {v.permission}
                        </td>
                        <td>{roleNames(v.roles).join(', ')}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              {:else}
                <p>Not registered yet</p>
              {/if}
            </td>
          </tr>
          </tbody>
        </table>
      </form>


      <div class="text-center" style="margin-top: 1rem">
      <Alert id={alertId}/>
        <button type="submit" class="btn btn-primary" onclick={sync} disabled={syncing}>
          Sync to IAM Service
        </button>
      </div>
    </div>
  </div>
{:else}
  <div class="text-muted p-3">loading</div>
{/if}
