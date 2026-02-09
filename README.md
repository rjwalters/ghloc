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
      - uses: rjwalters/ghloc@v1
```

Then add the badge and chart to your README:

```markdown
![Lines of Code](.ghloc/badge.svg)
![LOC History](.ghloc/chart.svg)
```

The images won't render until the first push to main triggers the action.

### Branch Protection

The action commits and pushes to `main`. If your branch has protection rules requiring pull requests, the default `GITHUB_TOKEN` won't be able to push. To fix this:

1. Use **classic** branch protection (not rulesets) with **"Do not restrict administrators"** enabled
2. Create a [fine-grained personal access token](https://github.com/settings/personal-access-tokens/new) with **Contents: Read and write** scoped to your repo
3. Add it as a repository secret (e.g., `GH_LOC_PAT`) under **Settings > Secrets > Actions**
4. Pass the token to the checkout step:

```yaml
    steps:
      - uses: actions/checkout@v4
        with:
          token: ${{ secrets.GH_LOC_PAT }}
      - uses: rjwalters/ghloc@v1
```

If your branch has no protection rules, no extra setup is needed.

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
1. Check if the repo has a secret named GH_LOC_PAT (look for branch
   protection rules that require pull requests). If branch protection
   exists, the workflow needs a PAT token to push.

2. Create .github/workflows/loc.yml with this content:

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
         - uses: rjwalters/ghloc@v1

   If the repo has branch protection, add a token to the checkout step:

         - uses: actions/checkout@v4
           with:
             token: ${{ secrets.GH_LOC_PAT }}

3. Add these two lines near the top of README.md (after the title):

   ![Lines of Code](.ghloc/badge.svg)
   ![LOC History](.ghloc/chart.svg)

4. Commit and push to main.

The action will run on the push, count LOC, and commit badge.svg and
chart.svg to the .ghloc/ directory. The images will render after that
first run completes.
```

## Privacy

Your code never leaves GitHub's infrastructure. The action runs inside GitHub Actions on your own repository — no third-party servers, no external APIs. Only LOC statistics are stored in `.ghloc/history.json`.

## License

MIT
