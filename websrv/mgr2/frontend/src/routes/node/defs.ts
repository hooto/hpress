// node module shared option/typedef tables. Centralizes the inline constants
// that node/NodeListView and node/NodeSetView each duplicated (statusDef), plus
// the NodeSetView-local onoff/textFormats so all node constants live together.
export const statusDef = [
  { type: 1, name: 'Publish' },
  { type: 2, name: 'Draft' },
  { type: 3, name: 'Private' },
]

export const onoff = [
  { type: 'true', name: 'ON' },
  { type: 'false', name: 'OFF' },
]

export const textFormats = [
  { name: 'text', value: 'Text' },
  { name: 'html', value: 'Html' },
  { name: 'shtml', value: 'Script Html' },
  { name: 'md', value: 'Makedown' },
]
