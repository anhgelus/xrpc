# Contributing to XRPC

Before opening a pull request (PR) or an issue, check the existing ones to avoid creating a duplicate.

If you find a bug or if you have a suggestion to improve the library, you may open an issue.

## Pull Requests

Fork the repo, create a branch, commits and then open a PR.
We encourage you to open an issue first (if yours does not resolve one).
When you are writing the description of your PR, don't forget to link it.

Please, [test your code](#testing-your-code)!

When you have finished your PR, one of our maintainers will review your work.
If everything is fine, it will be merged (yay :D).
If your work is mostly good, the reviewer will ask you to fix issues.
If your PR has conceptual issues, it will be closed and the reviewer will explain you why.

**Read the rest of this document to avoid losing your (and our) times.**

We encourage you to watch
[this conference at FOSDEM 2026](https://fosdem.org/2026/schedule/event/L7ERNP-prs-maintainers-will-love/) to understand
how to write good PR.

## Use of AI

The maintainers of XRPC do not use LLMs.
We are against this technology for many reasons, but we can't stop you from using these tools.

**When you are interacting with us, do not use an LLM.**
If you are, we will instantly close your issue/your PR.

**If you have used an LLM to help you, you must inform us.**
We will not reject your work for this reason.
If you hide this, we will instantly close your issue/your PR.

We will not accept poorly written code with useless comments.
We will not accept code that does not follow our code style.
We will not accept PR that does not follow our contributing guide.
You have to modify the LLM's output.
You have to test the code that you want to be merged.
If you don't understand something created by an LLM, avoid creating a PR/an issue, thanks you.

**Remember that XRPC is made for humans and by humans.**
If your contribution does not follow this vision, avoid contributing.
Every PR is reviewed by at least one human, every issue is triaged by at least one human, every line of code was written
for humans.

## Style

To standardize and make things less messy, we have a certain code style that is persistent throughout the codebase.

### Commits

A commit is an atomic modification.
It cannot be divided into smaller ones.
It must update tests and it must work without additional commits.

We follow this simple schema for their name:
```
kind(scope): description
```
`kind` is the kind of modification:
- `feat` for an addition
- `refactor` for a refactor
- `fix` for fixing an issue
- `style` for the code style
- `docs` for the documentations
- `build` for building tools
- `ci` for CI/CD

`scope` indicates the part touched of your modification.
We commonly use `ws` for websocket, `guild` for `guild` and `guild/guildapi` packages...

`description` is a *short* description of your modification.
If you want to explain more things, include them in other lines (not the first one).

If your history is messy, you must modify it and force push the updated version.
In futur versions of git, you will be able to use the new `git-history(1)` to edit this easily.

```
fix(ws): not panicing if bad setup during connection
```
is fixing an issue related to websocket.
Before this commit, the bot was not panicing if there is a bad setup during the connection.
After this commit, this issue is fixed.

```
feat(logger): option to trim version in caller
```
is adding something to the logger.
Now, the developer can use an option to trim versions.

### Organization

The package `atproto` contains ATProto primitives.
These must be lexicon-agnostic and must work anywhere without requiring any lexicons.

The package `xrpc` (root) contains essential stuff required to work with a standard XRPC: definitions and 
`com.atproto.repo.*` implementations.

The package `server` has the `com.atproto.server.*` implementations.

Every lexicon under the NSID `com.atproto.*` could be implemented in this project if it follows our goals.
The package containg `com.atproto.foo.bar` should be `foo/bar`.

## Testing your code

Before submitting a PR, you must test your changes.

First, you can simply run tests with
```bash
go test -race ./...
```

If everything looks fine, we encourage you to create a simple project in another directory to test your changes:
```bash
go work init . # init a new go.work file for this module
go work use path/to/xrpc # override the XRPC in the go.mod by the one present in this folder
```

