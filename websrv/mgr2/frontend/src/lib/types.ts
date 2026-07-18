// Core API types mirroring api/*.go and the /hp/v1 response envelope. Every
// response inlines TypeMeta: { kind, apiVersion?, error?: {code,message}, ... }.
// Lists add { meta: TypeListMeta, items: [...] }.

export interface ApiError {
  code: string
  message: string
}

export interface TypeListMeta {
  totalResults?: number
  startIndex?: number
  itemsPerList?: number
  selfLink?: string
  resourceVersion?: string
}

export interface TypeMeta {
  kind?: string
  apiVersion?: string
  error?: ApiError
}

// ---- Node ----
export interface NodeField {
  name: string
  title?: string
  type?: string
  value: string
  attrs?: { key: string; value: string }[]
  // multi-language: { items: {lang, value}[] }
  langs?: { items: { lang: string; value: string }[] }
}

export interface NodeTerm {
  name: string
  value: string
}

export interface NodeModelField {
  name: string
  title?: string
  type?: string
  length?: number
  indexType?: number
  attrs?: { key: string; value: string }[]
}

export interface NodeModelTerm {
  meta?: { name: string }
  title?: string
  type?: string // taxonomy | tag
}

export interface NodeModelExtensions {
  access_counter?: boolean
  comment_enable?: boolean
  comment_perentry?: boolean
  node_refer?: string
  node_sub_refer?: boolean
  text_search?: boolean
  permalink?: boolean
  [k: string]: any
}

export interface NodeModel extends TypeMeta {
  meta?: { name: string }
  title?: string
  modname?: string
  fields?: NodeModelField[]
  terms?: NodeModelTerm[]
  extensions?: NodeModelExtensions
}

export interface Node extends TypeMeta {
  id: string
  status: number
  title: string
  created?: number
  updated?: number
  fields?: NodeField[]
  terms?: NodeTerm[]
  model?: any
  ext_comment_perentry?: boolean
  ext_permalink_name?: string
  ext_node_refer?: string
  [k: string]: any
}

export interface NodeList extends TypeMeta {
  meta?: TypeListMeta
  model?: any
  items?: Node[]
}

// ---- Term ----
export interface Term extends TypeMeta {
  id: string
  pid?: string
  title: string
  status?: number
  weight?: number
  created?: number
  updated?: number
  type?: string
}

export interface TermList extends TypeMeta {
  meta?: TypeListMeta
  model?: { title?: string; type?: string }
  items?: Term[]
}

export interface TermModel extends TypeMeta {
  meta?: { name: string }
  type?: string
  title?: string
  modname?: string
}

// ---- Spec / ModSet ----
export interface SpecMeta {
  id?: string
  name: string
  version?: string
  created?: number
  updated?: number
}

export interface Spec extends TypeMeta {
  meta?: SpecMeta
  srvname?: string
  title?: string
  status?: number
  theme_config?: string
  nodeModels?: NodeModel[]
  termModels?: TermModel[]
  actions?: SpecAction[]
  router?: { routes?: SpecRoute[] }
  views?: { path?: string }[]
}

export interface SpecList extends TypeMeta {
  items?: Spec[]
}

export interface SpecActionDataxQuery {
  table?: string
  limit?: number
  order?: string
}

export interface SpecActionDatax {
  name: string
  type: string // node.list | node.entry | term.list | term.entry
  pager?: boolean
  query?: SpecActionDataxQuery
  cache_ttl?: number
}

export interface SpecAction extends TypeMeta {
  name: string
  modname?: string
  datax?: SpecActionDatax[]
}

export interface SpecRoute extends TypeMeta {
  path: string
  dataAction?: string
  template?: string
  modname?: string
  params?: { [k: string]: string }
  default?: boolean
}

export interface SpecTemplateList extends TypeMeta {
  items?: { path: string }[]
}

export interface SpecLangItem {
  id: string
  name: string
}

export interface SpecLangList extends TypeMeta {
  items?: SpecLangItem[]
}

// ---- S2 / FS ----
export interface FsFile extends TypeMeta {
  path?: string
  name?: string
  size?: number
  body?: string
  encode?: string
  sumcheck?: string
  modtime?: string
  isdir?: boolean
}

export interface FsFileList extends TypeMeta {
  path?: string
  items?: FsFile[]
}

// ---- Sys ----
export interface SysConfigItem {
  key: string
  value: string
  comment?: string
  type?: string
}

export interface SysConfigList extends TypeMeta {
  items?: SysConfigItem[]
}

export interface SysStatus extends TypeMeta {
  instance_id?: string
  app_version?: string
  app_release?: string
  runtime_version?: string
  uptime?: number
  coroutine_number?: number
  info?: any
  memstats?: any
  [k: string]: any
}

export interface IamPermission {
  summary?: string
  permission?: string
  roles?: string[]
}

export interface SysIamInstance {
  id?: string
  name?: string
  version?: string
  url?: string
  permissions?: IamPermission[]
  [k: string]: any
}

export interface SysIamStatus extends TypeMeta {
  base_url?: string
  app_id?: string
  secret_key?: string
  instance_self?: SysIamInstance
  instance_registered?: SysIamInstance
  [k: string]: any
}

// ---- Pager (mirrors hpMgr.Pager) ----
export interface Pager {
  ItemCount: number
  CountPerPage: number
  PageCount: number
  CurrentPageNumber: number
  FirstPageNumber: number
  PrevPageNumber: number
  NextPageNumber: number
  LastPageNumber: number
  RangeLen: number
  RangeStartNumber: number
  RangeEndNumber: number
  RangePages: number[]
}
