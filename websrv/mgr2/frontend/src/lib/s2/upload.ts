// S2 object upload helpers — base64 data-URL POST (server cap 10 MiB/file), plus
// recursive drag-drop folder traversal via webkitGetAsEntry (ports
// hps2_fsUploadTraverseTree / hps2_fsUploadCommit in s2.js).
import { api, ApiError } from '../api'

export interface UploadResult {
  ok: boolean
  msg: string
  name: string
}

export function readFileAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const r = new FileReader()
    r.onload = () => resolve(r.result as string)
    r.onerror = () => reject(r.error)
    r.readAsDataURL(file)
  })
}

export async function uploadS2Object(ppath: string, file: File): Promise<UploadResult> {
  if (file.size > 10 * 1024 * 1024) {
    return { ok: false, name: file.name, msg: file.name + ' Failed: File is too large to upload' }
  }
  try {
    const body = await readFileAsDataURL(file)
    const rsp = await api.post<{ kind?: string }>('s2-obj/put', {
      path: ppath + '/' + file.name,
      size: file.size,
      body,
      encode: 'base64',
    })
    if (rsp && rsp.kind === 'FsFile') return { ok: true, name: file.name, msg: file.name + ' OK' }
    return { ok: false, name: file.name, msg: file.name + ' Failed' }
  } catch (e) {
    if (e instanceof ApiError) {
      return { ok: false, name: file.name, msg: file.name + ' Failed: ' + (e.message || '') }
    }
    return { ok: false, name: file.name, msg: file.name + ' Failed' }
  }
}

// Flatten dropped items (files + folders) into a flat file list with relative
// paths. Folder structure is collapsed (server path uses file.name only, like
// the legacy uploader). Returns a flat File[].
export async function collectDroppedFiles(items: DataTransferItemList): Promise<File[]> {
  const out: File[] = []
  const entries: any[] = []
  for (let i = 0; i < items.length; i++) {
    const item: any = (items[i] as any).webkitGetAsEntry?.()
    if (item) entries.push(item)
  }
  for (const entry of entries) {
    await walkEntry(entry, out)
  }
  return out
}

function walkEntry(entry: any, out: File[]): Promise<void> {
  return new Promise((resolve) => {
    if (entry.isFile) {
      entry.file((file: File) => {
        out.push(file)
        resolve()
      })
    } else if (entry.isDirectory) {
      const reader = entry.createReader()
      const readBatch = () =>
        reader.readEntries(async (batch: any[]) => {
          if (!batch.length) return resolve()
          for (const b of batch) await walkEntry(b, out)
          readBatch()
        })
      readBatch()
    } else {
      resolve()
    }
  })
}
