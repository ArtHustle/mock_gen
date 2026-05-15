# Project rules for Claude

## Git: do not commit or push without explicit request

Never run `git commit`, `git push`, `git tag`, or any other write-side git
operation unless the user has explicitly asked for it in the current turn.
"The work is done" or "tests pass" is not a request to commit. Wait for an
explicit instruction such as "commit", "push", "tag", or equivalent.

This applies to all branches and all tags, including bumping versions,
re-tagging, and amending. When in doubt, stop and ask.

## Tests are the source of truth

Tests outrank production code. They define the contract; production code is
the implementation that has to satisfy it. When a test and the code disagree,
the test is usually right — fix the code, not the test. Only change a test
when its assertion is genuinely wrong, and say so out loud when you do.

Tests are also the project's documentation. A reader should be able to learn
how a feature works by reading its tests. Keep entries short, name them so
the description alone explains the case, and let the expected output speak
for itself rather than wrapping it in narration.

## Test-driven by default

Prefer TDD: write the failing test first, then the code that makes it pass,
then refactor. When working from an existing test suite, the workflow is:
add or update the failing test that captures the new requirement, run it,
implement, run again. This catches design mistakes early and produces tests
that actually exercise the code path you changed.

Exceptions are fine — exploratory spikes, trivial renames, doc tweaks — but
state when you're skipping TDD and why.

## Keep tests clean and short

- One entry per behavior. If two entries assert the same thing in slightly
  different shapes, one of them is redundant.
- No overlap across files. If `foo_test.go` already asserts a shape, don't
  re-assert it in `bar_test.go`. Cross-reference instead.
- Prefer literal expected output for shape tests; prefer substring or
  ordering assertions for negation/invariant tests. Pick the form that
  expresses the assertion most directly.
- Strip incidental setup. If the test doesn't read a value, don't compute it.
- Comments in tests explain *why*, never *what*. The code already says what.
