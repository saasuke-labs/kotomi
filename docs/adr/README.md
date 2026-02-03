# Architecture Decision Records (ADR)

This directory contains Architecture Decision Records (ADRs) for Kotomi. ADRs are documents that capture important architectural decisions made along with their context and consequences.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences. ADRs help:

- Document the reasoning behind architectural choices
- Provide historical context for future developers
- Enable informed decision-making when revisiting designs
- Share knowledge across the team

## Format

Each ADR follows this structure:

1. **Status**: Proposed, Accepted, Deprecated, Superseded
2. **Context**: What is the issue we're seeing that is motivating this decision?
3. **Decision**: What is the change that we're proposing?
4. **Consequences**: What becomes easier or more difficult as a result of this change?

## ADR List

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [001](001-user-authentication-for-comments-and-reactions.md) | User Authentication for Comments and Reactions | Proposed | 2026-01-31 |
| [002](002-code-structure-and-go-1.25-improvements.md) | Code Structure and Go 1.25 Improvements | Proposed | 2026-02-03 |

## Creating a New ADR

To create a new ADR:

1. Copy an existing ADR as a template
2. Increment the number (e.g., `002-your-decision-title.md`)
3. Fill in the sections with your decision details
4. Update this README with a link to the new ADR
5. Submit as a pull request for review

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Michael Nygard's original article](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions)
