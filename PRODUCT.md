# Product

<!-- impeccable:product-schema 1 -->

## Platform

adaptive

## Stack

Go core with a Linux-native GTK4 interface. The core has no GTK dependency so future Windows and macOS clients can reuse it.

## Users

Linux desktop users who need a local folder to stay synchronized with cloud storage without managing shell commands, timers, or separate synchronization scripts.

## Product Purpose

Nuvem provides continuous, understandable cloud-folder synchronization on Linux. Initial support is Google Drive; future integrations may include Dropbox and OneDrive.

## Positioning

Nuvem combines a native Linux control surface with a single managed background service and an embeddable synchronization core, rather than exposing cloud-provider command-line configuration to the user.

## Operating Context

Users choose a local folder and a remote folder once, then use the desktop app to see status, recent activity, and conflicts. The background service performs synchronization after the app window is closed.

## Capabilities and Constraints

- English and Brazilian Portuguese ship in the first version.
- Linux UI is native first; core code remains portable for later platforms.
- Google Drive is the first provider; Dropbox and OneDrive are future providers.
- A conflict automatically selects the most recently modified version and records the decision.
- The current beta embeds selected MIT-licensed rclone components while a Nuvem-native engine is developed.

## Brand Commitments

The name is Nuvem. The product should feel calm, trustworthy, and simple rather than terminal-oriented.

## Evidence on Hand

The repository includes a working command-line beta, a systemd user service, and bilingual CLI messages. No visual assets or established graphical interface exist yet.

## Product Principles

1. One setup, then stay out of the way.
2. Never run concurrent synchronization jobs.
3. Explain state and failures in plain language.
4. Resolve routine conflicts automatically, while leaving an audit trail.
5. Keep provider-specific mechanics behind a stable product interface.

## Accessibility & Inclusion

The interface must support keyboard navigation, readable system typography, and translated user-facing text.
