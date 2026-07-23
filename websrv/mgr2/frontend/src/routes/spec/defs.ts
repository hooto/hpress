// spec module type/option definitions — direct ports of hpSpec.*def in spec.js.
export { objectClone, withStableScroll } from '../../lib/util'

export const statuses = [
  { name: 'Enable', value: 1 },
  { name: 'Disable', value: 0 },
]

export const termTypedef = [
  { type: 'taxonomy', name: 'Categories' },
  { type: 'tag', name: 'Tags' },
]

export const dataxTypedef = [
  { type: 'list', name: 'List' },
  { type: 'entry', name: 'Entry' },
]

export const fieldTypedef = [
  { type: 'bool', name: 'Bool' },
  { type: 'string', name: 'Varchar' },
  { type: 'text', name: 'Text' },
  { type: 'date', name: 'Date' },
  { type: 'datetime', name: 'Datetime' },
  { type: 'int8', name: 'int8' },
  { type: 'uint8', name: 'uint8' },
  { type: 'int16', name: 'int16' },
  { type: 'uint16', name: 'uint16' },
  { type: 'int32', name: 'int32' },
  { type: 'uint32', name: 'uint32' },
  { type: 'int64', name: 'int64' },
  { type: 'uint64', name: 'uint64' },
  { type: 'float', name: 'Float' },
  { type: 'decimal', name: 'Decimal Float' },
]

export const fieldIdxTypedef = [
  { type: 0, name: 'No Index' },
  { type: 1, name: 'General Index' },
  { type: 2, name: 'Unique Index' },
]

export const generalOnoff = [
  { type: 'true', name: 'ON' },
  { type: 'false', name: 'OFF' },
]

export const permalinkDef = [
  { type: '', name: 'OFF' },
  { type: 'name', name: 'Name' },
]

export const namereg = /^[a-z][a-z0-9_]+$/

export const specdef = {
  kind: 'Spec',
  meta: { id: '', name: '' },
  srvname: '',
  title: '',
  status: 1,
  theme_config: '',
}

export const termdef = {
  kind: 'TermModel',
  meta: { name: '' },
  type: 'taxonomy',
  title: '',
  modname: '',
}

export const nodedef = {
  kind: 'NodeModel',
  meta: { name: '' },
  title: '',
  modname: '',
  fields: [],
  terms: [],
  extensions: {
    access_counter: false,
    comment_enable: false,
    comment_perentry: false,
    node_refer: '',
    text_search: false,
  },
}

export const actiondef = {
  kind: 'SpecAction',
  name: '',
  datax: [],
}

export const routedef = {
  kind: 'SpecRoute',
  path: '',
  dataAction: '',
  template: '',
  modname: '',
  params: {},
  default: false,
}
