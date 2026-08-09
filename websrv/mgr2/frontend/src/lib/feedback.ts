// "flash then act" helper. Centralizes the repeated pattern of showing an
// inline success/info alert, then closing the modal / refreshing the list
// after a short delay so the user actually sees the confirmation. Previously
// every Set form and delete handler inlined innerShow(...) + setTimeout(...,N)
// with ad-hoc 500–1000ms magic numbers.
import { innerShow, type AlertType } from './alert'

// Show `msg` as `type` under `id`, then run `after` once after `ms` (so the
// confirmation stays visible briefly before the modal slides away / the list
// reloads). A null `after` just flashes the alert.
export function flashThen(
  id: string,
  type: AlertType,
  msg: string,
  after?: () => void,
  ms = 600,
): void {
  if (type) innerShow(id, type, msg)
  if (after) setTimeout(after, ms)
}
