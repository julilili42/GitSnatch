# GitSnatch

GitSnatch is a fast and user-friendly CLI tool built in Go to copy the content of one or multiple files from a GitHub repository directly to your clipboard.

## Features

- Fetch file contents from public/private GitHub repositories.
- Interactive file selection.
- Clipboard integration for convenience.
- Support for specific commit SHAs.

## Installation

Install GitSnatch via Go:

```bash
go install github.com/julilili42/GitSnatch@latest
```

Set your GitHub token:

```bash
export GITHUB_TOKEN=your_token
```

## Usage

Fetch repository files:

```bash
gitsnatch fetch [repoOwner] [repoName] [commitSHA]
```

Interactive fetch:

```bash
gitsnatch fetch
```

## Examples

Interactive mode:

```bash
gitsnatch fetch
```

Explicit repository and commit:

```bash
gitsnatch fetch julilili42 GitSnatch f9c2f0a
```

## Configuration

Requires a GitHub Personal Access Token with repo access:

[Generate Token](https://github.com/settings/tokens)

## Dependencies

- [spf13/cobra](https://github.com/spf13/cobra)
- [AlecAivazis/survey](https://github.com/AlecAivazis/survey)
- [atotto/clipboard](https://github.com/atotto/clipboard)
