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
- **Actions.** A surface keeps its actions in one place, at the top right: in the page toolbar, or in a row above the content with a short line of context on the left. The green button goes to the action the surface exists for. Surfaces that people mainly browse keep their actions in outline. Creating something is "New" on the canvas and "New <thing>" everywhere else, empty states included.
- **Forms.** Every form is a card that ends in the same footer: a short hint about what saving does, Discard once something has changed, and Save, which waits for a change. A switch that stands on its own applies right away; a switch inside a form waits for Save. Leaving unsaved changes asks first, and a background refresh keeps the edits in progress. One kind of data gets one editor, wherever it appears.
- **Pending.** While an action runs, its button shows a spinner in place of its icon and a label like "Saving…". The fields it acts on lock until it finishes, and a switch that applies right away shows a small spinner beside it.
- **Loading.** Show a skeleton shaped like the content that will appear, in the place where it will appear. Surfaces open right away and fill in. A refresh keeps the current content on screen.
- **Errors.** Every error tells the person which of three things happened. A lost connection asks them to check it and try again, a refusal from the server shows its reason as given, and a failure on our side apologizes and offers Report. A failed load shows the shared error block where the data would be, with "Couldn't load …", the reason and Try again. An error never looks like an empty list, and editing stays off until the data is there. Status messages use the shared alert in one of four tones.
- **Empty states.** An empty state takes the place of the content it stands in for. A page that has nothing at all shows the large serif state on the shared crosshatch. Panels, drawers and cards use the compact one, with an icon in the shared square, a short title, one sentence, and the action that fills it. "All clear" results use the compact state in the ok tone.
- **Toasts.** A toast reports an outcome the person can't see where they acted, like a restore that finishes later or an invite that went out by mail. Results that show up in place, like a new node on the canvas or a switch settling, speak for themselves. A failed action says "Failed to" plus its verb, with the reason below, and stays until it is dismissed, while a lost connection fades on its own because it usually clears up. One failure gets one toast, and loading failures and copying give feedback in place instead. Every toast sits in the dialogs' inset frame, tinted by its tone, with an icon mark and a dismiss control.
- **Sections.** A section heading is short, quiet and in sentence case. Its explanation is the same size, in a softer color. Space separates groups. Delete comes last, as a plain card with a red-tinted border.
