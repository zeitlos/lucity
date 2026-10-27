---
paths:
  - "services/dashboard/src/**"
---

# Dashboard design language

Design guidelines for every new or changed view, one per area.

- **Buttons and icons.** All controls share one height, so a field and its button line up. The compact size is only for dense rows inside panels. A button sets the size and spacing of its own icon. Every icon-only control has a label and the shared dark tooltip. Each icon means one thing: a cross only closes, and a bin deletes or removes.
- **Surfaces.** Drawers and resource panels share one shell: a prominent title, a line under the header, and a close control. Only resource panels carry an icon, to tell resource kinds apart. Headers hold no helper text. If guidance is needed, keep it short and put it next to what it explains. Escape closes the topmost surface, but never while a menu, dialog or focused field is still holding it.
- **Dialogs and confirmations.** Every dialog opens with an icon, sits in an inset frame that fades in from the top, and keeps its actions below a full-width line. Destructive actions always ask first, through one shared confirmation of a fixed height:
  - The title names the action and the thing.
  - The button repeats the verb.
  - The text says when it can't be undone.

  Deleting something that holds data asks you to type its name. Say Delete when something is destroyed, Remove when it's detached, and Revoke for access. Quote names in copy instead of setting them in monospace.
- **Sections.** A section heading is short, quiet and in sentence case, and any explanation under it is smaller than the heading. Separate groups with space, never with capitals or divider lines. Delete sits last, as a plain card set apart only by its red-tinted border, with no "danger zone" banner.
