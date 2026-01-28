# Feature Specifications

This directory contains feature specifications using a three-document pattern for planning substantial features.

## Structure

Each feature occupies a directory labeled `f###-short-description/` containing:

- **requirements.md**: Defines WHAT to build (specifications, acceptance criteria)
- **design.md**: Defines HOW to build (architecture, implementation)
- **tasks.md**: Tracks execution (phases, progress, lessons learned)

## Clear WHAT vs HOW Separation

**requirements.md** = Specifications (NO implementation details):
- Business logic and user needs
- Acceptance criteria (GIVEN/WHEN/THEN)
- Performance targets (WHAT metrics)
- Out of scope decisions
- Success criteria

**design.md** = Implementation (ALL technical details):
- Architecture and component design
- Technical decisions and trade-offs
- Algorithms and data structures
- Database changes, API changes
- Testing approach

**tasks.md** = Execution tracking:
- Implementation phases (2-3 hour chunks)
- Progress updates
- Lessons learned retrospective

## Naming Convention

Features use sequential numbering with three digits: `f001`, `f002`, `f003`, etc.

Never reuse abandoned feature numbers.

Examples:
- `f001-sqlite-migration/`
- `f002-pronunciation-guides/`
- `f003-analytics-dashboard/`

## When to Create Specs

### Warrant Full Specifications

- Language support additions
- Major architectural changes
- New storage backends
- API surface modifications
- Performance-critical work
- Integration features
- Database schema changes

### Don't Require Specs

- Bug fixes
- Documentation updates
- Refactoring without behavior changes
- Dependency updates
- Minor tweaks

## Philosophy

**Templates serve you, not vice versa.**

Simple features might skip sections entirely, while complex ones use all three documents. If the requirement is obvious from the feature name, skip the formal spec.

## Workflow

1. **Requirements Phase**:
   - Reference source GitHub Issue for traceability
   - Define explicit "out of scope" items
   - Use GIVEN/WHEN/THEN format for acceptance criteria
   - Identify open questions

2. **Design Phase**:
   - Explore existing code patterns
   - Answer architectural questions
   - Plan testing approach
   - Document technical trade-offs

3. **Tasks Phase**:
   - Break implementation into 2-3 hour phases
   - Track progress with status updates
   - Document surprises and lessons learned
   - Update retrospectively during work

4. **Implementation**:
   - Follow the design
   - Update tasks.md with progress
   - Link PRs to the feature spec

5. **Archive**:
   - Mark feature as complete in tasks.md
   - Keep spec for historical reference
   - Enables future contributors to understand design rationale

## GitHub Traceability

Always reference the source GitHub Issue:
```markdown
**GitHub Issue**: [#11 - Change migration to preserve words](https://github.com/wizact/te-reo-bot/issues/11)
```

Link PRs back to the feature spec in PR descriptions:
```markdown
Part of feature spec: docs/features/f001-migration-preserve-words/
```

## Templates

See `TEMPLATES/` directory for:
- `requirements.md.template` - WHAT to build
- `design.md.template` - HOW to build
- `tasks.md.template` - Execution tracking

Copy templates to start a new feature spec:
```bash
mkdir docs/features/f###-feature-name
cp docs/features/TEMPLATES/*.template docs/features/f###-feature-name/
# Rename .template to .md
# Fill in the content
```
