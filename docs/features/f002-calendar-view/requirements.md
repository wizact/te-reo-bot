# Feature Requirements: Curator TUI

**Related**: [Design](./design.md) | [Tasks](./tasks.md)
**Type**: Feature
**Priority**: High
**Created**: 2026-02-01
**Updated**: 2026-06-21

## Overview

Local terminal-based curator application for maintaining the Māori word database in `data/words.db`. The tool is table-centric, keyboard-first, and optimized for single-user curation on a local machine.

## Requirements

### R1: Word Listing

**User Story**: As a curator, I want to list all words and their metadata in a terminal UI, so that I can review the current dataset quickly.

**Acceptance Criteria**:
- The system shall load all words from SQLite using the repository layer
- The system shall display word ID, day index, word text, and meaning in the main table
- The system shall show full metadata for the selected word in a details pane
- The system shall support unassigned words where `day_index IS NULL`

### R2: Sorting and Filtering

**User Story**: As a curator, I want to sort and filter words from the keyboard, so that I can find records quickly.

**Acceptance Criteria**:
- The system shall support sorting by multiple columns including day index, word text, meaning, ID, and updated time
- The system shall support ascending and descending sort order
- The system shall support text filtering across word metadata
- The system shall update the table without requiring a mouse

### R3: Unicode and Macron Support

**User Story**: As a curator, I want Māori macrons and Unicode text to work correctly, so that I can curate words accurately.

**Acceptance Criteria**:
- The system shall treat input and display text as UTF-8 end-to-end
- The system shall support filtering words containing macrons
- The system shall avoid ASCII-only assumptions in search and table rendering

### R4: Word Creation and Editing

**User Story**: As a curator, I want to add and edit words from the terminal, so that I can maintain the dataset locally.

**Acceptance Criteria**:
- The system shall provide modal forms for adding and editing words
- The system shall require non-empty word text and meaning
- The system shall allow editing optional metadata fields including link, photo, and attribution
- The system shall persist changes to SQLite immediately after save

### R5: Day Assignment Workflow

**User Story**: As a curator, I want to assign, clear, and reassign day indexes, so that I can organize the posting calendar from the terminal.

**Acceptance Criteria**:
- The system shall allow assigning a selected word to a specific day index from 1 to 366
- The system shall allow clearing a selected word back to an unassigned state
- The system shall auto-assign the next free day index when requested
- IF a target day is already assigned during reassignment, THEN the system shall swap the two assignments predictably
- IF a new word is created with explicit day assignment and that day is occupied, THEN the system shall reject the save with a curator-friendly error

### R6: Validation and Linting

**User Story**: As a curator, I want to validate the dataset, so that I can detect gaps before publishing.

**Acceptance Criteria**:
- The system shall report missing day indexes across 1..366
- The system shall report assigned and unassigned word counts
- The system shall report duplicate or invalid day assignments if encountered
- The system shall report empty word or meaning fields if encountered
- The system shall surface validation results inside the TUI and via a CLI validation mode

### R7: Keyboard-First Dark UI

**User Story**: As a curator, I want a dark terminal UI that works without the mouse, so that I can curate efficiently from the keyboard.

**Acceptance Criteria**:
- The system shall use a dark terminal theme by default
- The system shall be fully operable with keyboard shortcuts
- The system shall provide a persistent shortcut/status area
- The system shall not require mouse interaction for any core workflow

### R8: Local Operation

**User Story**: As a curator, I want the tool to run locally against SQLite, so that I can work offline.

**Acceptance Criteria**:
- The system shall run as a local binary from `cmd/curator`
- The system shall default to `./data/words.db` when no database path is provided
- The system shall reuse the repository layer instead of embedding raw SQL in the TUI
- The system shall provide a non-interactive `-validate` mode for scripting and CI checks

## Out of Scope

### Permanently Out of Scope
- Browser-based calendar UI
- Drag-and-drop interaction
- Mouse-only workflows
- Multi-user collaboration
- Cloud synchronization
- Mobile UI

### Out of Scope (v1.0)
- Undo/redo history
- Batch editing of multiple rows
- Image preview rendering in the terminal
- Monthly calendar visualization

## Success Criteria

1. ✅ Curator can browse all words and metadata from the terminal
2. ✅ Curator can sort, filter, add, edit, assign, and unassign without a mouse
3. ✅ Auto-allocation finds the next free day index reliably
4. ✅ Validation reports missing days and summary counts locally and in CLI mode
5. ✅ Māori macrons work correctly in filtering and display

## References
- [tech.md](../../constitution/tech.md) - repository and application architecture
- [conventions.md](../../conventions.md) - Go coding standards
- [pkg/repository/interface.go](../../../pkg/repository/interface.go) - repository contract used by the curator app
- [pkg/curator](../../../pkg/curator/) - curator service and TUI implementation
