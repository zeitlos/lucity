---
paths:
  - "services/dashboard/src/**"
---

# Dashboard design language

Design guidelines for every new or changed view, one per area.

- **Buttons and icons.** All controls share one height, so a field and its button line up. The compact size is only for dense rows inside panels. A button sets the size and spacing of its own icon. Every icon-only control has a label and the shared dark tooltip. Each icon means one thing: a cross only closes, and a bin deletes or removes.
- **Surfaces.** Drawers and resource panels share one shell: a prominent title, a line under the header, and a close control. Only resource panels carry an icon, to tell resource kinds apart. Headers hold no helper text. If guidance is needed, keep it short and put it next to what it explains. Escape closes the topmost surface, but never while a menu, dialog or focused field is still holding it.
