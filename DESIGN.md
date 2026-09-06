# Design

## Direction

Nuvem follows the calm, direct grammar expected of a Linux desktop utility.
The first surface is a single GTK4 window with clear status, one configuration
area, and two primary actions: save the setup and synchronize now.

## Visual rules

- Use the system GTK theme and typography so Nuvem belongs beside GNOME apps.
- Put current synchronization state first; use plain language instead of
  transfer jargon.
- Keep folder setup and Google OAuth connection visible and editable rather
  than hiding essential setup behind a modal or terminal command.
- Use the system error style only for a problem with an actionable recovery.
- Preserve the same hierarchy in English and Brazilian Portuguese.

## Accessibility

All controls are native GTK controls with keyboard focus and accessible labels.
The local-folder selector uses the system file chooser.
