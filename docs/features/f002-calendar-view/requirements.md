# Feature Requirements: Calendar View for Word Organization

**Related**: [Design](./design.md) | [Tasks](./tasks.md)
**Type**: Feature
**Priority**: Medium
**Created**: 2026-02-01

## Overview
Interactive calendar interface for organizing M\u0101ori words across 366 days of the year. Enables drag-and-drop assignment of words to calendar days, with word bank for unassigned words. Single-user local tool for content curation.

## Requirements

### R1: Calendar Display

**User Story**: As a content curator, I want to view the current year in a calendar format with daily view, so that I can see all 366 days at a glance.

**Acceptance Criteria**:
- The system shall display a 12-month calendar view for the current year
- The system shall show all 366 days including leap year day (Feb 29)
- WHEN a day has an assigned word, the system shall display the word text on that calendar day
- WHEN a day has no assigned word, the system shall display an empty state indicator
- The system shall highlight the current day in the calendar

### R2: Word Data Loading

**User Story**: As a content curator, I want words loaded from words.db based on day_index, so that I see the current word assignments.

**Acceptance Criteria**:
- WHEN the calendar view initializes, the system shall read all words from words.db
- The system shall map words to calendar days using the day_index field (1-366)
- WHERE day_index IS NULL, the system shall place words in the unassigned word bank
- The system shall load word metadata: id, word text, meaning, day_index

### R3: Drag and Drop Functionality

**User Story**: As a content curator, I want to drag words between calendar days, so that I can reorganize word assignments easily.

**Acceptance Criteria**:
- WHEN I drag a word from one day to another day, the system shall update the source day to empty and the target day to show the moved word
- IF target day already has a word, THEN the system shall swap the two words (source word to target, target word to source)
- The system shall support dragging words from calendar days to the word bank (sets day_index to NULL)
- The system shall support dragging words from word bank to calendar days (sets day_index to target day number)
- WHEN drag operation completes, the system shall update the UI immediately to reflect the change

### R4: Unassigned Word Bank

**User Story**: As a content curator, I want to see all words without day assignments in a sidebar, so that I can allocate them to specific days.

**Acceptance Criteria**:
- The system shall display a sidebar showing all words where day_index IS NULL
- The system shall show word text and meaning for each unassigned word
- The system shall support scrolling when unassigned words exceed viewport height
- WHEN a word is assigned to a day from the word bank, the system shall remove it from the sidebar
- WHEN a word is unassigned from a day to the word bank, the system shall add it to the sidebar

### R5: Database Persistence

**User Story**: As a content curator, I want changes saved to words.db, so that word assignments persist across sessions.

**Acceptance Criteria**:
- WHEN a word is moved to a different day, the system shall update the day_index field in words.db
- WHEN a word is moved to the word bank, the system shall set day_index to NULL in words.db
- WHEN a word is assigned from word bank to a day, the system shall set day_index to the target day number (1-366) in words.db
- IF database update fails, THEN the system shall revert the UI to the previous state and display an error message
- The system shall use existing Repository pattern methods (UpdateWordDayIndex)

### R6: User Interface Technology (⚠️ PENDING USER APPROVAL)

**Options**:

1. **Web-based UI (Go templates + HTMX + Vanilla JS)**
   - ✅ Pros: Minimal dependencies, integrates with existing Go server, browser-native drag-and-drop, lightweight
   - ❌ Cons: Requires web server running, less polished UX than desktop apps

2. **Web-based UI (Go + Tailwind CSS + Alpine.js)**
   - ✅ Pros: Modern styling, reactive UI, still lightweight, good developer experience
   - ❌ Cons: Additional build step for CSS, new dependency (Tailwind)

3. **Desktop App (Electron + React/Vue)**
   - ✅ Pros: Rich desktop UX, extensive drag-and-drop libraries, familiar to web developers
   - ❌ Cons: Heavy dependencies, separate from Go ecosystem, larger binary size

4. **Desktop App (Wails - Go + Web Frontend)**
   - ✅ Pros: Go backend integration, native OS experience, single binary deployment
   - ❌ Cons: New framework to learn, Go 1.18+ required (project uses 1.13)

**Decision Required**: User must select preferred technology stack before design phase.

### R7: Single-User Local Operation

**User Story**: As a content curator, I want the tool to run on my local machine without server deployment, so that I can manage words offline.

**Acceptance Criteria**:
- The system shall run entirely on the local machine
- The system shall access words.db directly from data/ directory
- The system shall not require internet connectivity for core functionality
- The system shall not require authentication or multi-user support

### R8: Manual Validation

**User Story**: As a content curator, I want to validate word assignments separately, so that I can verify data integrity when I choose.

**Acceptance Criteria**:
- The system shall NOT perform automatic validation during drag-and-drop operations
- The system shall allow me to run dict-gen validate command manually after organizing words
- The system shall save changes immediately without validation checks

## Out of Scope

### Permanently Out of Scope
- **Multi-user support**: Single curator only
- **Real-time collaboration**: No concurrent editing
- **Cloud synchronization**: Local database only
- **Mobile interface**: Desktop/laptop only
- **Word editing**: Only day assignment changes (use dict-gen for word CRUD)
- **Automatic validation**: User runs validator manually
- **Image preview**: Text-only display for words
- **Undo/redo**: No history tracking (direct database updates)
- **Search/filter**: All words visible at all times

### Out of Scope (v1.0)
- **Year selection**: Current year only (2026)
- **Multiple calendar views**: Daily view only (no weekly/monthly variations)
- **Keyboard shortcuts**: Mouse-only interaction
- **Accessibility features**: Basic browser accessibility only
- **Performance optimization**: < 1000 words only
- **Batch operations**: One word at a time

## Success Criteria

1. ✅ All requirements (R1-R8) pass acceptance tests
2. ✅ Drag-and-drop works for all 366 days + word bank
3. ✅ Database updates persist correctly (verified with dict-gen validate)
4. ✅ UI loads in < 2 seconds for 366 words
5. ✅ Single curator can reorganize entire year in < 30 minutes
6. ✅ Works offline on macOS (primary platform)

## References
- [tech.md](../../constitution/tech.md) - Repository pattern
- [conventions.md](../../conventions.md) - Go coding standards
- [pkg/repository/interface.go](../../../pkg/repository/interface.go) - UpdateWordDayIndex method
- [f001-preserve-migration-words](../f001-preserve-migration-words/) - Word bank pattern
