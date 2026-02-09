# ghloc — GitHub Lines of Code

![Lines of Code](.ghloc/badge.svg)

![LOC History](.ghloc/chart.svg)

A GitHub Action that automatically counts lines of code in your repository and generates LOC badges and history charts.

## Features

- **LOC counting** — Uses [scc](https://github.com/boyter/scc) for fast, accurate line counting
- **SVG badge** — Rendered badge committed to your repo
- **LOC history chart** — Star-history-style line chart showing LOC over time
- **Zero infrastructure** — Runs as a GitHub Action, no server needed

## Usage

Add this workflow to your repo at `.github/workflows/loc.yml`:

```yaml
name: Lines of Code
on:
  push:
    branches: [main]
jobs:
  loc:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
      - uses: rjwalters/ghloc@main
```

Then add the badge and chart to your README:

```markdown
![Lines of Code](.ghloc/badge.svg)
![LOC History](.ghloc/chart.svg)
```

The images won't render until the first push to main triggers the action.

## Inputs

| Input | Description | Default |
|---|---|---|
| `directory` | Directory to count | `.` |

## How It Works

On every push to main, the action:

1. Counts lines of code using [scc](https://github.com/boyter/scc)
2. Appends a snapshot to `.ghloc/history.json`
3. Generates `.ghloc/badge.svg` and `.ghloc/chart.svg`
4. Commits the changes back to the repo with `[skip ci]`

## Quick Install (copy-paste for AI agents)

If you're using an AI coding agent, paste this prompt to add LOC tracking to your repo:

```
Add a lines-of-code badge and history chart to this repo using the
ghloc GitHub Action (https://github.com/rjwalters/ghloc).

Steps:
1. Create .github/workflows/loc.yml with this content:

   name: Lines of Code
   on:
     push:
       branches: [main]
   jobs:
     loc:
       runs-on: ubuntu-latest
       permissions:
         contents: write
       steps:
         - uses: actions/checkout@v4
         - uses: rjwalters/ghloc@main

2. Add these two lines near the top of README.md (after the title):

   ![Lines of Code](.ghloc/badge.svg)
   ![LOC History](.ghloc/chart.svg)

3. Commit and push to main.

The action will run on the push, count LOC, and commit badge.svg and
chart.svg to the .ghloc/ directory. The images will render after that
first run completes.
```

## Privacy

Your code never leaves GitHub's infrastructure. The action runs inside GitHub Actions on your own repository — no third-party servers, no external APIs. Only LOC statistics are stored in `.ghloc/history.json`.

## License

MIT
