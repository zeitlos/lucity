---
paths:
  - "services/dashboard/src/**"
---

# Dashboard design language

Design guidelines for every new or changed view, one per area.

- **Scale.** Every text size, space and radius comes from the standard scale, so the same values repeat across views.
- **Buttons and icons.** All controls share one height, so a field and its button line up. The compact size is for dense rows inside panels. A button sets the size and spacing of its own icon. Every icon-only control has a label and the shared dark tooltip. Each icon has one meaning: a cross closes, and a bin deletes or removes.
- **Surfaces.** Drawers and resource panels share one shell: a prominent title, a line under the header, and a close control. Resource panels carry an icon to tell resource kinds apart. A header holds only the title and close. Any guidance is short and sits next to what it explains. Escape closes the topmost surface, once menus, dialogs and focused fields have handled it.
- **Dialogs and confirmations.** Every dialog opens with an icon, sits in an inset frame that fades in from the top, and keeps its actions below a full-width line. Destructive actions ask first, through one shared confirmation of a fixed height:
  - The title names the action and the thing.
  - The button repeats the verb.
  - The text says when it can't be undone.

  Deleting something that holds data asks for its name. Delete destroys, Remove detaches, and Revoke ends access. Names in copy are quoted.
- **Sections.** A section heading is short, quiet and in sentence case. Its explanation is the same size, in a softer color. Space separates groups. Delete comes last, as a plain card with a red-tinted border.
