// Ambient declaration for spark-md5 (its @types package triggers an interactive
// version prompt under pnpm in this environment, so we declare the surface we
// use instead). mirroring lynkui.utilx.cryptoMd5 -> SparkMD5.hash.
// Ambient declaration for spark-md5 (its @types package triggers an interactive
// version prompt under pnpm in this environment, so we declare the surface we
// use instead). mirroring lynkui.utilx.cryptoMd5 -> SparkMD5.hash.
//
// spark-md5 ships a CommonJS UMD module whose `module.exports = <the SparkMD5
// class>` — there is NO named export. Importing it as `import { SparkMD5 }`
// silently binds undefined (the bundle reads `mod.SparkMD5`, which doesn't
// exist), so callers must use the DEFAULT import: `import SparkMD5 from
// 'spark-md5'`. Only the default export is declared here to make that the only
// option the type-checker will accept.
declare module 'spark-md5' {
  const SparkMD5: {
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

