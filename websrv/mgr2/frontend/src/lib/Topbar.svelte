<script lang="ts">
  // Top navigation bar, structurally identical to the legacy #hpm-topbar in
  // websrv/mgr/views/index.tpl. Branding comes from sys/config-list; the
  // dynamic per-module nav comes from mod-set/spec-list (status==1).
  import { siteName, siteLogo, specs } from './boot'
  import { hashRoute } from './router'
  import { paths } from './config'

  function nav(prefix: string) {
    return $hashRoute === prefix || $hashRoute.startsWith(prefix + '/')
  }
  function nodeNav(name: string) {
    return $hashRoute === 'node/index/' + name
  }
</script>

<div id="hpm-topbar" class="hpm-topbar">
  <div class="hpm-topbar-collapse">
    <ul class="hpm-nav" id="hpm-topbar-siteinfo">
      {#if $siteLogo}
        <li><img class="hpm-topbar-logo" src={$siteLogo} alt="" /></li>
      {/if}
      <li class="hpm-topbar-brand">{$siteName}</li>
    </ul>

    <ul class="hpm-nav hpm-topbar-nav" id="hpm-topbar-nav-node-specls">
      {#each $specs as s (s.meta?.name)}
        <li>
          <a
            class={'lynkui-nav-item' + (nodeNav(s.meta!.name) ? ' active' : '')}
            href={'#node/index/' + s.meta!.name}>{s.title || s.meta!.name}</a
          >
        </li>
      {/each}
    </ul>

    <ul class="hpm-nav hpm-nav-right" id="hpm-topbar-userbar">
      <li><a href={paths.signOut}>Sign Out</a></li>
    </ul>
    <ul class="hpm-nav hpm-nav-right">
      <li>
        <a class={'lynkui-nav-item' + (nav('s2/index') ? ' active' : '')} href="#s2/index">Storage</a>
      </li>
      <li>
        <a class={'lynkui-nav-item' + (nav('spec/index') ? ' active' : '')} href="#spec/index"
          >Modules</a
        >
      </li>
      <li>
        <a class={'lynkui-nav-item' + (nav('sys/index') || $hashRoute.startsWith('sys/') ? ' active' : '')}
          href="#sys/index">System</a
        >
      </li>
    </ul>
  </div>
</div>
