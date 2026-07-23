// mod-set-fs file upload helper — base64 data-URL PUT to mod-set-fs/put under
// a module (ports l9rPodFs.Post encode=base64 in spec-editor.js). 10 MiB/file
// cap (server limit on mod-set-fs/put). Shared by the spec-editor single-file
// picker and the drag-drop batch upload modal.
import { api, ApiError } from './api'
import { readFileAsDataURL } from './s2/upload'

export interface FsUploadResult {
  ok: boolean
  msg: string
  name: string
}

export async function uploadModFsFile(
  modname: string,
  path: string,
  file: File,
): Promise<FsUploadResult> {
  if (file.size > 10 * 1024 * 1024) {
    return { ok: false, name: file.name, msg: file.name + ' Failed: File is too large to upload' }
  }
  try {
    const body = await readFileAsDataURL(file)
    const rsp = await api.post<{ kind?: string }>(
      'mod-set-fs/put',
      { path, body, encode: 'base64' },
      { modname },
    )
    if (rsp && rsp.kind === 'FsFile') return { ok: true, name: file.name, msg: file.name + ' OK' }
    return { ok: false, name: file.name, msg: file.name + ' Failed' }
  } catch (e) {
    const detail = e instanceof ApiError ? e.message : String(e)
    return { ok: false, name: file.name, msg: file.name + ' Failed: ' + detail }
  }
}

export function joinFsPath(dir: string, name: string): string {
  return (dir.replace(/\/+$/, '') + '/' + name).replace(/\/+/g, '/')
}
