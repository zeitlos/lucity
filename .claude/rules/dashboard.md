---
paths:
  - "services/dashboard/src/**"
---

# Dashboard design language

The decisions behind every view. Most of the look already lives in the shared pieces, so reach for those first.

- **Scale.** Every value comes from the theme scale: space, text size, color, shadow and radius. To adjust something, move one step on that scale. The theme redefines parts of that scale, radii among them, so read its values before picking a step.
- **Controls.** Icon-only controls use the shared labelled icon button. A cross closes, and a bin deletes or removes. The compact size is for dense rows inside panels.
- **Surfaces.** Drawers and resource panels share one shell, and only resource panels carry an icon. Guidance is a short line next to what it explains.
- **Dialogs.** Every dialog opens with the shared icon. A destructive action goes through the shared confirmation, whose button repeats the verb, and deleting something that holds data asks for its name. Delete destroys, Remove detaches and Revoke ends access. Names in copy are quoted.
- **Actions.** A surface keeps its actions together at the top right. Green is for the one action the surface exists for, and surfaces people mainly browse keep theirs in outline. The canvas says New, everywhere else says New <thing>.
- **Forms.** Every form ends in the shared footer, and Save waits for a change. A switch on its own applies right away, and a switch inside a form waits for Save. One kind of data has one editor, wherever it appears.
- **Loading and pending.** A load shows a skeleton shaped like the content, and a refresh keeps what is on screen. A running action spins on its own button, reads "Saving…" or the like, and locks the fields it acts on.
- **Errors.** A failed load shows the shared error block where the data would be, and editing waits until the data is there. A failed action shows the shared failure toast.
- **Empty states.** The large state is for a page with nothing at all. Panels, drawers and cards use the compact one, in the ok tone for "all clear".
- **Toasts.** A toast is for an outcome the person can't see where they acted, like a restore that finishes later or a delete that takes them somewhere else. It names the thing and the past tense, like Workspace "acme" deleted. Results that show up in place speak for themselves.
- **Status.** Words and tones come from the shared status map. Canvas cards show a pill, with a release in flight or a failed release in a tab on top. Lists and settings rows show just the word, colored only when something needs attention. The Deployments tab keeps its own badges.
- **Tables and lists.** Tables and lists in panels and drawers run to the edge, and every table is built from the shared table. A row that opens a page is a real link, and row actions stay visible.
- **Switches.** Changing a view in place, like a time range or a layout, uses the shared segmented control.
- **Data.** Numbers, dates and relative times go through the shared formatting, and schedules show in the viewer's own time. Secrets, snippets and log streams use the shared credential row, code block and log viewer.
- **Copy.** Labels use sentence case and name engines by name, like PostgreSQL and Redis. New opens a create surface, Create submits it, Add adds a row and Save commits a form. Sentences break with periods and commas.
- **Sections.** A heading is short and quiet, and its explanation is the same size in a softer color. Delete comes last, as a plain card with a red-tinted border.
- **Nesting.** A surface holds its content directly, with at most one level of card inside it. Inside a panel, drawer or dialog, space and a heading group the content. A card holds a form, or a set of rows that belong together.
- **Lines.** Every edge gets one line. Inside a panel, content runs out to the sides without lines of its own. A grid of cards shares the line between neighbours. Where space or a tinted background already separates two things, the gap is enough on its own.
