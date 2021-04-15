# Contributing

## Setup

Clone the repository from your personal fork on GitHub.

Then, setup the project referring to [README.md](./README.md).

## Commit Message

We are following the [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) message rules, though there are minor differences, to enable us to generate changelogs and follow the [semantic versioning](https://semver.org/).

Each commit message consists of a header, an optional body and an optional footer. It should be structured as follows:

```
<type>[(optional scope)]: <description>

[optional body]

[optional footer(s)]
```

Examples:

```
feat(event): implement event flush worker
```

```
refactor: drop support for Node 6

we don't support Node 6 any more because of XXX reason.

BREAKING CHANGE: refactor to use JavaScript features not available in Node 6
```

```
docs: add link to xxx doc
```

## Creating a pull request

After pushing your commits to the forked repository, create a pull request to the original repository on GitHub.

## Code review

After a pull request created, Bucketeer team members will review it.

Then, After the pull request approved, we will merge it.
