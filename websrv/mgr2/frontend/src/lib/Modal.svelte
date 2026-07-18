<script lang="ts">
  // Renders the modal stack (modals store). Top modal sits on top. Each layer
  // is a centered Bootstrap-styled panel sized per spec.width/height.
  import { modals, closeModal, prevModal, type ModalSpec } from './modal'
  import { alertClose } from './alert'

  export let scope = 'hpm' // css scope prefix

  function boxStyle(m: ModalSpec): string {
    let s = ''
    if (m.width === 'max' || m.height === 'max') {
      s = 'width:96vw;height:92vh;max-width:96vw;max-height:92vh;'
    } else {
      if (m.width) s += `width:${m.width}px;`
      if (m.height && m.height !== 'auto') s += `height:${m.height}px;`
      if (m.height === 'auto') s += 'height:auto;'
    }
    return s
  }

  function btnClick(m: ModalSpec, dismiss: boolean, click?: () => void) {
    if (click) click()
    if (dismiss !== false) closeModal()
  }
</script>

{#each $modals as m, i (m)}
  {@const Body = m.body}
  <div
    class={`${scope}-modal-overlay lynkui-modal-overlay`}
    style="z-index:{1300 + i}"
    role="dialog"
    aria-modal="true"
  >
    <div class={`${scope}-modal-box lynkui-modal`} style={boxStyle(m)}>
      {#if m.title}
        <div class={`${scope}-modal-head lynkui-modal-head`}>
          {#if m.backEnable !== false && i > 0}
            <button
              type="button"
              class="btn btn-dark btn-sm lynkui-modal-back"
              on:click={() => prevModal(m.onPrev)}>Back</button
            >
          {/if}
          <span class={`${scope}-modal-title`}>{m.title}</span>
          <button
            type="button"
            class="btn-close lynkui-modal-close"
            aria-label="Close"
            on:click={() => closeModal()}></button
          >
        </div>
      {/if}
      <div class={`${scope}-modal-body lynkui-modal-body lynkui-scroll`}>
        {#if Body}
          <Body {...m.props} />
        {:else if m.html}
          {@html m.html}
        {/if}
      </div>
      {#if m.buttons && m.buttons.length}
        <div class={`${scope}-modal-foot lynkui-modal-foot`}>
          {#each m.buttons as b}
            <button type="button" class={`btn ${b.class || 'btn-dark'}`} {...{}}
              on:click={() => btnClick(m, b.dismiss !== false, b.click)}>{b.title}</button
            >
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/each}

<style>
  .hpm-modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.45);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 4vh;
  }
  .hpm-modal-box {
    background: #fff;
    border-radius: 6px;
    box-shadow: 0 6px 30px rgba(0, 0, 0, 0.3);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 320px;
  }
  .hpm-modal-head {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 10px 14px;
    border-bottom: 1px solid #eee;
    background: #fafafa;
  }
  .hpm-modal-title {
    flex: 1;
    font-weight: 600;
    font-size: 0.95rem;
  }
  .hpm-modal-body {
    flex: 1;
    overflow: auto;
    padding: 14px;
  }
  .hpm-modal-foot {
    padding: 10px 14px;
    border-top: 1px solid #eee;
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
    background: #fafafa;
  }
  .lynkui-scroll {
    overflow: auto;
  }
</style>
