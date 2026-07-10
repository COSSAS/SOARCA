<!--
Thanks for contributing to SOARCA!
Before opening this PR, please confirm:
  - There is at least one issue linked below. Every PR needs a corresponding issue.
  - Your branch name follows the allowed prefixes (see Contribution Guidelines).
  - This PR targets the `development` branch (not `master`).
Do NOT report security vulnerabilities here. Contact the maintainers via Slack instead.
-->

## Summary

<!-- What does this PR do and why? A few sentences is fine. -->

## Related issue

<!-- Every PR needs at least one issue. Use a closing keyword so it links automatically. -->

Closes #

## Type of change

<!-- Tick all that apply. -->

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that changes existing behaviour or API)
- [ ] Documentation
- [ ] Refactor / internal change (no functional change)
- [ ] CI / build / tooling

## What changed

<!-- List the concrete changes. Call out anything a reviewer should look at closely. -->

-

## How this was tested

<!-- Describe the tests you added or ran. Include commands and relevant output where useful. -->

-

---

## Contributor checklist

<!-- The author fills this in before requesting review. -->

- [ ] There is at least one issue linked above, and this PR is scoped to it
- [ ] I checked for existing logic first, and reused what was already there instead of duplicating it
- [ ] Branch name uses an allowed prefix: `feature/`, `feature/docs/`, `bugfix/`, `release/x.x`, or `hotfix/`
- [ ] PR targets the `development` branch
- [ ] Branch is rebased on the latest `development`, with a clean history (no merge commits)
- [ ] `make test` (unit tests) passes locally
- [ ] `make integration-test` passes where relevant
- [ ] `make lint` (golangci-lint) passes with no new issues
- [ ] New or changed logic is covered by tests
- [ ] Code follows the [Google Go style guide](https://google.github.io/styleguide/go/), with the project's exceptions: receiver names are descriptive (e.g. `info`, not `i`), and initialisms are CamelCase (e.g. `Xml`, not `XML`)
- [ ] Public API changed → Swagger regenerated (`make swagger-docs`) and committed
- [ ] User-facing behaviour changed → docs updated (`/documentation`)
- [ ] Config or environment variables changed → `.env.example` updated
- [ ] Self-reviewed the diff: no debug code, stray logging, or commented-out blocks
- [ ] I have read and followed the [Contribution Guidelines](https://cossas.github.io/SOARCA/docs/contribution-guidelines/) and agree to the [Code of Conduct](https://cossas.github.io/SOARCA/docs/contribution-guidelines/code_of_conduct/)
- [ ] This PR does not disclose a security vulnerability (those go to the maintainers via Slack, not here)

## Reviewer checklist

<!-- A core maintainer fills this in during review. At least one approval is required to merge. -->

- [ ] **No duplication: this doesn't reimplement logic, helpers, or patterns that already exist in the codebase. Where existing code could have solved the issue, it is reused rather than rewritten.**
- [ ] **The approach is the right one: the problem couldn't have been solved more simply by extending or reusing what's already there.**
- [ ] A valid issue is linked and the change matches its scope (no unrelated changes riding along)
- [ ] Branch prefix and target branch (`development`) are correct
- [ ] CI is green: build, lint, unit tests, and integration tests
- [ ] Changes are covered by tests, and the tests actually exercise the new behaviour (not just present)
- [ ] Code is readable and follows the project's Go style; naming, error handling, and structure fit the codebase
- [ ] No obvious security concerns: input validation, injection, secrets in code/logs, unsafe defaults
- [ ] Public API / Swagger and user-facing docs are updated and accurate where behaviour changed
- [ ] Breaking changes are called out, justified, and acceptable for the current release line
- [ ] Commit history is clean enough to merge (rebased, meaningful messages)
- [ ] I have pulled and run the branch locally where the change warrants it

## Notes for reviewers

<!-- Optional: anything that helps review — screenshots, trade-offs, open questions, follow-ups. -->