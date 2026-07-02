import type { NavItem } from '~/types/nav'

// Presentation only: which modules render as sidebar entries, and their
// sub-routes/icons. Whether a module actually appears is still decided
// server-side by /me/manifest — this just knows how to render the ones
// that come back enabled.
export const moduleNavConfig: Record<string, NavItem> = {
  sales: {
    title: 'Sales',
    icon: 'lucide:handshake',
    children: [
      { title: 'Accounts', link: '/sales/accounts' },
      { title: 'Contacts', link: '/sales/contacts' },
      { title: 'Leads', link: '/sales/leads' },
    ],
  },
  service: {
    title: 'Service',
    icon: 'lucide:life-buoy',
    children: [
      { title: 'Cases', link: '/service/cases' },
      { title: 'Queues', link: '/service/queues' },
    ],
  },
  collab: {
    title: 'Chat',
    icon: 'lucide:message-circle',
    link: '/chat',
  },
}

export function isNavGroup(item: NavItem): item is Extract<NavItem, { children: unknown[] }> {
  return 'children' in item
}
