// Ambient declaration for spark-md5 (its @types package triggers an interactive
// version prompt under pnpm in this environment, so we declare the surface we
// use instead). mirroring lynkui.utilx.cryptoMd5 -> SparkMD5.hash.
declare module 'spark-md5' {
  export const SparkMD5: {
    hash(str: string, raw?: boolean): string
    hashBinary(content: string, raw?: boolean): string
    ArrayBuf: {
      create(): { append: (arr: ArrayBuffer) => void; end: (raw?: boolean) => string }
    }
  }
  export default SparkMD5
}

// CodeMirror 5 (npm) has no bundled types; declare loosely. Modes/addons are
// side-effect imports.
declare module 'codemirror' {
  const CodeMirror: any
  export default CodeMirror
}
declare module 'codemirror/mode/markdown/markdown.js'
declare module 'codemirror/mode/xml/xml.js'
declare module 'codemirror/addon/selection/active-line.js'
declare module 'codemirror/mode/markdown/markdown'
declare module 'codemirror/mode/xml/xml'
declare module 'codemirror/addon/selection/active-line'
// CodeMirror 5 modes used by the file IDE (lc-editor)
declare module 'codemirror/mode/javascript/javascript'
declare module 'codemirror/mode/css/css'
declare module 'codemirror/mode/htmlmixed/htmlmixed'
declare module 'codemirror/mode/clike/clike'
declare module 'codemirror/mode/go/go'
declare module 'codemirror/mode/php/php'
declare module 'codemirror/mode/python/python'
declare module 'codemirror/mode/yaml/yaml'
declare module 'codemirror/mode/sql/sql'
declare module 'codemirror/mode/lua/lua'
declare module 'codemirror/mode/ruby/ruby'
declare module 'codemirror/mode/shell/shell'

