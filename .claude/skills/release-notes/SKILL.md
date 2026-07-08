---
name: release-notes
description: >-
  Write Lucity GitHub release notes / changelog in the house style. Use this
  whenever the user asks to draft release notes, write a changelog, prepare a
  release announcement, summarize what has shipped since the last release or
  tag, or "write the notes for vX.Y.Z", even if they don't say "skill". Turns
  merged PRs and commits since the previous release into an intro placeholder
  (the maintainer writes the opening in their own voice) followed by grouped,
  user-facing, use-case-oriented bullets, with the social-preview image header,
  then writes the result straight into a draft GitHub release for the next
  CalVer tag. Internal-only changes are omitted unless they change day-to-day
  platform usage.
---

# Lucity release notes

Draft GitHub release notes that read like the maintainer wrote them personally, not
like an auto-generated changelog. The audience is developers using the platform: they
care about what changed *for them*, not about the internals. Your job is to translate
a pile of merged PRs into that.

**Read [references/examples.md](references/examples.md) first.** It contains the two
most recent real releases verbatim plus a "what to notice" breakdown. It is the ground
truth for tone, structure, and altitude. Everything below is the procedure; the examples
are the style.

## Workflow

### 1. Figure out the range

Find the last published release and collect everything merged since. From the repo root:

```sh
gh release list --limit 8                      # find the last non-draft release tag
git log <lastTag>..HEAD --oneline              # commits since then
```

PRs are squash-merged with a `(#NN)` suffix in the commit subject, so `git log` already
carries the PR numbers. Some changes land as direct commits with no PR number (e.g.
`fix deploy polling`) — that's fine, include them and just omit the `(#NN)`.

**Compute the next tag.** Tags use a CalVer scheme `vYY.M.PATCH`: two-digit year, month
with no leading zero, then a patch counter that resets each month. The patch always starts
at **1, never 0** (the first release of July 2026 is `v26.7.1`, its second is `v26.7.2`). So
take the current year and month; if releases already exist for that `vYY.M`, the next tag is
the highest patch plus one, otherwise it's `.1`. Ignore `-rc` pre-release and draft tags when
scanning for the highest patch, and confirm with the maintainer if the target month is
ambiguous (e.g. you're cutting a late release for the previous month).

```sh
YY=$(date +%y); M=$(( 10#$(date +%m) ))          # e.g. 26 and 7
gh release list --limit 40                        # find the highest existing patch for vYY.M
```

When a commit subject is too terse to know the user impact, read the PR:

```sh
gh pr view <NN>                                 # title + body explain what and why
```

Group related PRs by feature area, not one bullet per commit. Several PRs (`#31`, `#34`)
often add up to one story.

### 2. Decide what's user-facing

This is the core judgment. Ask of each change: *would a developer using Lucity notice or
benefit from this?*

**Include** — anything that changes what users can do or experience:
- New capabilities in the dashboard, CLI, or API (new resource types, new actions)
- Changed behavior of deploys, builds, variables, databases, volumes, domains, billing
- Security and isolation improvements that protect user workloads, data, or credentials
- Reliability, performance, and uptime improvements users will actually feel
- Bug fixes for problems a user could hit
- Landing page, docs, and onboarding changes users interact with

**Omit** — internal work with no day-to-day effect:
- Pure refactors, code cleanup, lint fixes, renames with no behavior change
- Test-only changes and dev-tooling that only the team touches
- CI / build-pipeline tweaks (unless they change release availability or cadence users feel)
- Dependency bumps with no user-visible effect

**Judgment call** — internal work that *does* have day-2 impact: include it, but frame it
by what the user feels, never by the implementation. The v26.6.1 control-plane-rewrite
paragraph in the examples is the model: a big internal change described purely as "you
might feel it in the stability of the platform and the cadence of new features." If you
can't state a user-visible upside in one honest sentence, leave it out.

When in doubt, show the maintainer a short "included / omitted / unsure" list before writing
the full draft. It's cheaper to confirm than to over- or under-report.

### 3. Write it

Structure, top to bottom:

1. **The image line, literally first**, then one blank line:
   ```
   <img width="1280" height="640" alt="lucity-social-preview" src="PASTE_IMAGE_URL" />
   ```
   See "The social-preview image" below for how the maintainer gets that URL. Leave the
   `PASTE_IMAGE_URL` placeholder in your draft; you can't mint the `user-attachments` URL.

2. **An intro placeholder listing the highlights, not finished prose.** The opening
   paragraph is what makes it a Lucity release and not a dependabot digest, and the
   maintainer writes it in their own voice. So don't ghost-write it. Just leave a clearly
   marked placeholder with the 3-6 headline items the intro should cover, nothing else: no
   theme, no personal-hook suggestions, no emoji. Use this shape:

   ```
   > Write the intro here (one to three short paragraphs, your voice, no em dashes).
   > Highlights to cover:
   > - <headline item 1>
   > - <headline item 2>
   > - <...>
   ```

   The example intros in [references/examples.md](references/examples.md) show the target
   the maintainer is aiming for. Your job is to hand off the highlights, not to write it.

3. **Grouped bullets.** Plain-text group headings (`Variables`, `Persistent Volumes`,
   `Security & isolation`, `Reliability & operations`, …), not markdown `##` headers.
   Order groups by how much users care: new capabilities first, polish and ops last.
   Each bullet says what the user can now do and why it helps, in plain language. Append
   `(#NN)` where a PR exists.

### 4. Voice

Follow the repo's marketing rules (`.claude/rules/marketing.md`), which these releases already embody:

- Confident and friendly, like explaining something cool to a smart colleague. An
  occasional quip lands better than three forced ones.
- **Never use em dashes.** Use periods, commas, colons, semicolons, or parentheses.
- Lead with what's genuinely new and useful. Be specific about what the platform now does;
  avoid buzzword soup, empty superlatives, and walls of text.
- **Use product-facing names, and credit the open source we build on.** Call it "Redis"
  (the name users see) and credit the engine in a link: "backed by [Valkey](https://valkey.io/)".
  Same for [PostgreSQL], [Zot], [CloudNativePG], [Hubble], Gateway API, [mise]. This builds
  trust and honors the ejectability story.
- **Never name the backing cloud or infra vendor** (the host provider, block-storage vendor,
  object-storage vendor, etc.). Engine and standard names are fine ("S3-compatible",
  "PostgreSQL"); the physical vendor is not. This mirrors how the dashboard talks.
- Light playful shade at other PaaS providers is fine; punch with substance, not sneers.

### 5. Deliver into the draft release

Write the notes straight into the draft GitHub release for the computed tag, so the
maintainer opens GitHub to a ready draft: drop in the social image, replace the intro
placeholder with their own words, publish.

Save the notes to `tmp/release-notes-<tag>.md`, then:

- If a release for the tag already exists (`gh release view <tag>`), update it in place:
  ```sh
  gh release edit <tag> --draft=true --title "<tag>" --notes-file tmp/release-notes-<tag>.md
  ```
- Otherwise create it as a draft:
  ```sh
  gh release create <tag> --draft --title "<tag>" --notes-file tmp/release-notes-<tag>.md
  ```

Keep it a **draft**. Never publish and never push the git tag: publishing is the maintainer's
call, and their moment to write the intro. Drafts are visible only to collaborators, so
creating or editing one is safe and reversible.

If you find a near-match draft whose tag is a typo (wrong year, wrong patch) but is clearly
meant to be this release, reuse it rather than leaving a duplicate: correct its tag while
writing the notes, and say so in your handoff.

```sh
gh release edit <typo-tag> --tag <correct-tag> --draft=true --title "<correct-tag>" \
  --notes-file tmp/release-notes-<correct-tag>.md
```

After writing, print the release URL (`gh release view <tag> --json url -q .url`) and remind
the maintainer of the two manual steps left: the social image and the intro paragraph.

## The social-preview image

Every release opens with a 1280×640 social card. It is generated in the dashboard, not by
this skill:

1. Open the dashboard **/brand** page ([services/dashboard/src/pages/BrandPage.vue](services/dashboard/src/pages/BrandPage.vue)),
   scroll to the **Social Preview** section.
2. Pick a background wallpaper, and optionally tailor the **Headline** / **Tagline** to the
   release theme (the default is "The PaaS you can leave." / "Open-source · Kubernetes-native
   · Ejectable"). A headline that nods at the release's theme is a nice touch but optional.
3. Click **PNG** (or **JPG**) to download `lucity-social-preview.png` (1280×640).
4. In the GitHub release editor, drag the downloaded file into the description box. GitHub
   uploads it and inserts the `<img ... src="https://github.com/user-attachments/assets/…" />`
   tag automatically — that replaces the `PASTE_IMAGE_URL` placeholder in the draft.

Remind the maintainer of these steps when you hand off the draft, since the URL only exists
after the upload.
