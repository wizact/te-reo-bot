package curator

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jroimartin/gocui"
	wotd "github.com/wizact/te-reo-bot/pkg/wotd"
)

const (
	listViewName       = "words"
	detailsViewName    = "details"
	statusViewName     = "status"
	modalFrameViewName = "modal_frame"
	messageViewName    = "message"
	messageTitle       = "Validation"
	helpTitle          = "Shortcuts"
	defaultStatus      = "Ready"
)

type modalMode string

const (
	modalNone    modalMode = ""
	modalForm    modalMode = "form"
	modalMessage modalMode = "message"
)

type formMode string

const (
	formFilter formMode = "filter"
	formAdd    formMode = "add"
	formEdit   formMode = "edit"
	formAssign formMode = "assign"
)

type formField struct {
	Name        string
	Label       string
	Value       string
	Placeholder string
}

type modalState struct {
	Mode      modalMode
	FormMode  formMode
	Title     string
	Message   []string
	Fields    []formField
	FieldView []string
	Active    int
}

// TUI is the interactive word curator interface.
type TUI struct {
	service        *Service
	options        ListOptions
	words          []wotd.Word
	selected       int
	status         string
	lastValidation *ValidationReport
	modal          modalState
	selectedWordID int
}

// NewTUI creates a keyboard-first word curator UI.
func NewTUI(service *Service) *TUI {
	return &TUI{
		service: service,
		options: ListOptions{
			SortColumn: SortByDayIndex,
		},
		status: defaultStatus,
	}
}

// Run starts the interactive curator UI.
func (t *TUI) Run() error {
	if err := t.reload(); err != nil {
		return err
	}

	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer gui.Close()

	gui.Cursor = true
	gui.SelFgColor = gocui.ColorBlack
	gui.SelBgColor = gocui.ColorCyan
	gui.BgColor = gocui.ColorBlack
	gui.FgColor = gocui.ColorWhite
	gui.SetManagerFunc(t.layout)

	if err := t.setKeybindings(gui); err != nil {
		return err
	}

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}

	return nil
}

func (t *TUI) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if maxX < 100 || maxY < 28 {
		v, err := g.SetView("too_small", 0, 0, maxX-1, maxY-1)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		v.Clear()
		v.Frame = false
		fmt.Fprintln(v, "Terminal too small for the curator TUI. Resize to at least 100x28.")
		return nil
	}
	g.DeleteView("too_small")

	listWidth := (maxX * 2) / 3
	statusTop := maxY - 5

	if err := t.layoutList(g, 0, 0, listWidth, statusTop); err != nil {
		return err
	}
	if err := t.layoutDetails(g, listWidth, 0, maxX-1, statusTop); err != nil {
		return err
	}
	if err := t.layoutStatus(g, 0, statusTop, maxX-1, maxY-1); err != nil {
		return err
	}
	if err := t.layoutModal(g, maxX, maxY); err != nil {
		return err
	}
	return nil
}

func (t *TUI) layoutList(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(listViewName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = fmt.Sprintf("Words • sort=%s • filter=%q", t.options.SortColumn, t.options.Filter)
	v.Frame = true
	v.Wrap = false
	v.Highlight = true
	v.SelFgColor = gocui.ColorBlack
	v.SelBgColor = gocui.ColorCyan
	v.BgColor = gocui.ColorBlack
	v.FgColor = gocui.ColorWhite
	v.Clear()

	header := formatListRow("ID", "Day", "Word", "Meaning")
	fmt.Fprintln(v, header)
	fmt.Fprintln(v, strings.Repeat("-", visibleWidth(header)))
	for _, word := range t.words {
		fmt.Fprintln(v, formatWordRow(word))
	}

	if len(t.words) == 0 {
		t.selected = 0
		t.selectedWordID = 0
		v.SetCursor(0, 0)
		v.SetOrigin(0, 0)
	} else {
		if t.selected >= len(t.words) {
			t.selected = len(t.words) - 1
		}
		if t.selected < 0 {
			t.selected = 0
		}
		t.selectedWordID = t.words[t.selected].ID
		if err := t.positionListCursor(v); err != nil {
			return err
		}
	}

	if t.modal.Mode == modalNone {
		g.SetCurrentView(listViewName)
	}

	return nil
}

func (t *TUI) layoutDetails(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(detailsViewName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = "Details"
	v.Frame = true
	v.Wrap = true
	v.BgColor = gocui.ColorBlack
	v.FgColor = gocui.ColorWhite
	v.Clear()

	word := t.currentWord()
	if word == nil {
		fmt.Fprintln(v, "No word selected.")
		return nil
	}

	fmt.Fprintf(v, "ID: %d\n", word.ID)
	fmt.Fprintf(v, "Day: %s\n", formatDay(word.DayIndex))
	fmt.Fprintf(v, "Word: %s\n", word.Word)
	fmt.Fprintf(v, "Meaning: %s\n", word.Meaning)
	fmt.Fprintf(v, "Link: %s\n", emptyPlaceholder(word.Link))
	fmt.Fprintf(v, "Photo: %s\n", emptyPlaceholder(word.Photo))
	fmt.Fprintf(v, "Attribution: %s\n", emptyPlaceholder(word.PhotoAttribution))
	fmt.Fprintf(v, "Created: %s\n", word.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(v, "Updated: %s\n", word.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(v, "Active: %t\n", word.IsActive)

	if t.lastValidation != nil {
		fmt.Fprintln(v)
		fmt.Fprintln(v, "Validation")
		fmt.Fprintln(v, "----------")
		fmt.Fprintf(v, "Assigned: %d\n", t.lastValidation.AssignedWords)
		fmt.Fprintf(v, "Unassigned: %d\n", t.lastValidation.UnassignedWords)
		fmt.Fprintf(v, "Missing days: %d\n", len(t.lastValidation.MissingDays))
		fmt.Fprintf(v, "Duplicate days: %d\n", len(t.lastValidation.DuplicateDays))
		fmt.Fprintf(v, "Out of range: %d\n", len(t.lastValidation.OutOfRangeDays))
	}

	return nil
}

func (t *TUI) layoutStatus(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView(statusViewName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = "Status"
	v.Frame = true
	v.Wrap = true
	v.BgColor = gocui.ColorBlack
	v.FgColor = gocui.ColorWhite
	v.Clear()

	fmt.Fprintln(v, t.status)
	fmt.Fprintln(v, "Arrows/jk move • / filter • c clear filter • s sort • g reverse • a add • e edit • d assign • u clear day • n next free • v validate • r reload • ? help • q quit")

	return nil
}

func (t *TUI) layoutModal(g *gocui.Gui, maxX, maxY int) error {
	switch t.modal.Mode {
	case modalForm:
		return t.layoutFormModal(g, maxX, maxY)
	case modalMessage:
		return t.layoutMessageModal(g, maxX, maxY)
	default:
		t.clearModalViews(g)
		return nil
	}
}

func (t *TUI) layoutFormModal(g *gocui.Gui, maxX, maxY int) error {
	width := maxX / 2
	if width < 50 {
		width = 50
	}
	height := len(t.modal.Fields)*3 + 6
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	frame, err := g.SetView(modalFrameViewName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	frame.Title = t.modal.Title
	frame.Frame = true
	frame.Wrap = true
	frame.BgColor = gocui.ColorBlack
	frame.FgColor = gocui.ColorYellow
	frame.Clear()
	fmt.Fprintln(frame, "Tab/↓ next • ↑ previous • Ctrl+S save • Esc cancel")

	t.modal.FieldView = t.modal.FieldView[:0]
	for i, field := range t.modal.Fields {
		fieldName := fmt.Sprintf("modal_field_%d", i)
		t.modal.FieldView = append(t.modal.FieldView, fieldName)
		fy0 := y0 + 2 + i*3
		fy1 := fy0 + 2
		input, viewErr := g.SetView(fieldName, x0+2, fy0, x1-2, fy1)
		if viewErr != nil && viewErr != gocui.ErrUnknownView {
			return viewErr
		}
		input.Title = field.Label
		input.Editable = true
		input.Wrap = false
		input.Frame = true
		input.BgColor = gocui.ColorBlack
		input.FgColor = gocui.ColorWhite
		if viewErr == gocui.ErrUnknownView {
			fmt.Fprint(input, field.Value)
			input.SetCursor(utf8.RuneCountInString(field.Value), 0)
		}
	}

	if len(t.modal.FieldView) > 0 {
		return t.focusModalField(g, t.modal.Active)
	}
	return nil
}

func (t *TUI) layoutMessageModal(g *gocui.Gui, maxX, maxY int) error {
	width := (maxX * 3) / 5
	height := len(t.modal.Message) + 4
	if height < 8 {
		height = 8
	}
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	t.clearFormViews(g)

	v, err := g.SetView(messageViewName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = t.modal.Title
	v.Frame = true
	v.Wrap = true
	v.BgColor = gocui.ColorBlack
	v.FgColor = gocui.ColorWhite
	v.Clear()
	for _, line := range t.modal.Message {
		fmt.Fprintln(v, line)
	}
	fmt.Fprintln(v)
	fmt.Fprintln(v, "Press Enter or Esc to close.")
	g.SetCurrentView(messageViewName)
	return nil
}

func (t *TUI) setKeybindings(g *gocui.Gui) error {
	bindings := []struct {
		view string
		key  interface{}
		fn   func(*gocui.Gui, *gocui.View) error
	}{
		{"", gocui.KeyCtrlC, t.quit},
		{"", 'q', t.quit},
		{"", gocui.KeyArrowDown, t.moveDown},
		{"", 'j', t.moveDown},
		{"", gocui.KeyArrowUp, t.moveUp},
		{"", 'k', t.moveUp},
		{"", '/', t.openFilterForm},
		{"", 'c', t.clearFilter},
		{"", 's', t.cycleSort},
		{"", 'g', t.toggleSortDirection},
		{"", 'a', t.openAddForm},
		{"", 'e', t.openEditForm},
		{"", 'd', t.openAssignForm},
		{"", 'u', t.clearDayIndex},
		{"", 'n', t.autoAssign},
		{"", 'v', t.validateAssignments},
		{"", 'r', t.reloadHandler},
		{"", '?', t.showHelp},
		{messageViewName, gocui.KeyEnter, t.closeModal},
		{messageViewName, gocui.KeyEsc, t.closeModal},
	}

	for _, binding := range bindings {
		if err := g.SetKeybinding(binding.view, binding.key, gocui.ModNone, binding.fn); err != nil {
			return err
		}
	}

	for i := 0; i < 8; i++ {
		viewName := fmt.Sprintf("modal_field_%d", i)
		if err := g.SetKeybinding(viewName, gocui.KeyTab, gocui.ModNone, t.nextModalField); err != nil {
			return err
		}
		if err := g.SetKeybinding(viewName, gocui.KeyArrowDown, gocui.ModNone, t.nextModalField); err != nil {
			return err
		}
		if err := g.SetKeybinding(viewName, gocui.KeyArrowUp, gocui.ModNone, t.previousModalField); err != nil {
			return err
		}
		if err := g.SetKeybinding(viewName, gocui.KeyCtrlS, gocui.ModNone, t.submitModal); err != nil {
			return err
		}
		if err := g.SetKeybinding(viewName, gocui.KeyEsc, gocui.ModNone, t.closeModal); err != nil {
			return err
		}
	}

	return nil
}

func (t *TUI) moveDown(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone || len(t.words) == 0 {
		return nil
	}
	if t.selected < len(t.words)-1 {
		t.selected++
		t.status = fmt.Sprintf("Selected %q", t.words[t.selected].Word)
	}
	return t.refreshGUI(g)
}

func (t *TUI) moveUp(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone || len(t.words) == 0 {
		return nil
	}
	if t.selected > 0 {
		t.selected--
		t.status = fmt.Sprintf("Selected %q", t.words[t.selected].Word)
	}
	return t.refreshGUI(g)
}

func (t *TUI) cycleSort(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	order := []SortColumn{SortByDayIndex, SortByWord, SortByMeaning, SortByID, SortByUpdated}
	for i, column := range order {
		if column == t.options.SortColumn {
			t.options.SortColumn = order[(i+1)%len(order)]
			break
		}
	}
	t.status = fmt.Sprintf("Sort column changed to %s", t.options.SortColumn)
	if err := t.reload(); err != nil {
		return t.handleError(err)
	}
	return t.refreshGUI(g)
}

func (t *TUI) toggleSortDirection(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	t.options.SortDescending = !t.options.SortDescending
	if t.options.SortDescending {
		t.status = "Sort direction: descending"
	} else {
		t.status = "Sort direction: ascending"
	}
	if err := t.reload(); err != nil {
		return t.handleError(err)
	}
	return t.refreshGUI(g)
}

func (t *TUI) openFilterForm(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	t.modal = modalState{
		Mode:     modalForm,
		FormMode: formFilter,
		Title:    "Filter words",
		Fields: []formField{
			{Name: "filter", Label: "Filter text", Value: t.options.Filter},
		},
	}
	return t.refreshGUI(g)
}

func (t *TUI) clearFilter(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	t.options.Filter = ""
	t.status = "Filter cleared"
	if err := t.reload(); err != nil {
		return t.handleError(err)
	}
	return t.refreshGUI(g)
}

func (t *TUI) openAddForm(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	t.modal = modalState{
		Mode:     modalForm,
		FormMode: formAdd,
		Title:    "Add word",
		Fields: []formField{
			{Name: "word", Label: "Word", Placeholder: "e.g. aroha"},
			{Name: "meaning", Label: "Meaning"},
			{Name: "link", Label: "Link"},
			{Name: "photo", Label: "Photo"},
			{Name: "photo_attribution", Label: "Photo attribution"},
			{Name: "day_index", Label: "Day index (blank or auto)"},
		},
	}
	return t.refreshGUI(g)
}

func (t *TUI) openEditForm(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	word := t.currentWord()
	if word == nil {
		return nil
	}
	t.modal = modalState{
		Mode:     modalForm,
		FormMode: formEdit,
		Title:    fmt.Sprintf("Edit word #%d", word.ID),
		Fields: []formField{
			{Name: "word", Label: "Word", Value: word.Word},
			{Name: "meaning", Label: "Meaning", Value: word.Meaning},
			{Name: "link", Label: "Link", Value: word.Link},
			{Name: "photo", Label: "Photo", Value: word.Photo},
			{Name: "photo_attribution", Label: "Photo attribution", Value: word.PhotoAttribution},
		},
	}
	return t.refreshGUI(g)
}

func (t *TUI) openAssignForm(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	word := t.currentWord()
	if word == nil {
		return nil
	}
	t.modal = modalState{
		Mode:     modalForm,
		FormMode: formAssign,
		Title:    fmt.Sprintf("Assign day for %s", word.Word),
		Fields: []formField{
			{Name: "day_index", Label: "Day index (blank or auto)", Value: formatDayInput(word.DayIndex)},
		},
	}
	return t.refreshGUI(g)
}

func (t *TUI) clearDayIndex(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	word := t.currentWord()
	if word == nil {
		return nil
	}
	if err := t.service.AssignDayIndex(word.ID, nil); err != nil {
		return t.handleError(err)
	}
	t.status = fmt.Sprintf("Cleared day index for %q", word.Word)
	if err := t.reload(); err != nil {
		return t.handleError(err)
	}
	return t.refreshGUI(g)
}

func (t *TUI) autoAssign(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	word := t.currentWord()
	if word == nil {
		return nil
	}
	day, err := t.service.AutoAssignNextDay(word.ID)
	if err != nil {
		return t.handleError(err)
	}
	t.status = fmt.Sprintf("Assigned day %d to %q", day, word.Word)
	if err := t.reload(); err != nil {
		return t.handleError(err)
	}
	return t.refreshGUI(g)
}

func (t *TUI) validateAssignments(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	report, err := t.service.Validate()
	if err != nil {
		return t.handleError(err)
	}
	t.lastValidation = report
	t.status = fmt.Sprintf("Validation completed with %d missing day indexes", len(report.MissingDays))
	t.modal = modalState{
		Mode:    modalMessage,
		Title:   messageTitle,
		Message: strings.Split(FormatValidationReport(report), "\n"),
	}
	return t.refreshGUI(g)
}

func (t *TUI) reloadHandler(g *gocui.Gui, v *gocui.View) error {
	if err := t.reload(); err != nil {
		return t.handleError(err)
	}
	t.status = "Reloaded words from SQLite"
	return t.refreshGUI(g)
}

func (t *TUI) showHelp(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalNone {
		return nil
	}
	t.modal = modalState{
		Mode:  modalMessage,
		Title: helpTitle,
		Message: []string{
			"Navigation: Arrow keys or j/k",
			"Filter: / to edit filter, c to clear",
			"Sort: s cycles day/word/meaning/id/updated, g toggles direction",
			"Word actions: a add, e edit, d assign day, u clear day, n auto-assign next free day",
			"Validation: v runs curator lint checks, r reloads data, q quits",
		},
	}
	return t.refreshGUI(g)
}

func (t *TUI) nextModalField(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalForm || len(t.modal.FieldView) == 0 {
		return nil
	}
	t.modal.Active = (t.modal.Active + 1) % len(t.modal.FieldView)
	return t.focusModalField(g, t.modal.Active)
}

func (t *TUI) previousModalField(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalForm || len(t.modal.FieldView) == 0 {
		return nil
	}
	t.modal.Active--
	if t.modal.Active < 0 {
		t.modal.Active = len(t.modal.FieldView) - 1
	}
	return t.focusModalField(g, t.modal.Active)
}

func (t *TUI) submitModal(g *gocui.Gui, v *gocui.View) error {
	if t.modal.Mode != modalForm {
		return nil
	}

	values := make(map[string]string, len(t.modal.FieldView))
	for i, viewName := range t.modal.FieldView {
		fieldView, err := g.View(viewName)
		if err != nil {
			return err
		}
		values[t.modal.Fields[i].Name] = strings.TrimSpace(fieldView.Buffer())
	}

	switch t.modal.FormMode {
	case formFilter:
		t.options.Filter = values["filter"]
		t.status = fmt.Sprintf("Filter updated to %q", t.options.Filter)
		if err := t.reload(); err != nil {
			return t.handleError(err)
		}
	case formAdd:
		input, err := wordInputFromForm(values)
		if err != nil {
			return t.handleError(err)
		}
		word, err := t.service.AddWord(input)
		if err != nil {
			return t.handleError(err)
		}
		t.status = fmt.Sprintf("Added %q", word.Word)
		if err := t.reload(); err != nil {
			return t.handleError(err)
		}
		t.selectedWordID = word.ID
	case formEdit:
		word := t.currentWord()
		if word == nil {
			return nil
		}
		input, err := wordInputFromForm(values)
		if err != nil {
			return t.handleError(err)
		}
		updated, err := t.service.UpdateWord(word.ID, input)
		if err != nil {
			return t.handleError(err)
		}
		t.status = fmt.Sprintf("Updated %q", updated.Word)
		if err := t.reload(); err != nil {
			return t.handleError(err)
		}
		t.selectedWordID = updated.ID
	case formAssign:
		word := t.currentWord()
		if word == nil {
			return nil
		}
		value := strings.TrimSpace(values["day_index"])
		if strings.EqualFold(value, "auto") {
			day, err := t.service.AutoAssignNextDay(word.ID)
			if err != nil {
				return t.handleError(err)
			}
			t.status = fmt.Sprintf("Assigned day %d to %q", day, word.Word)
		} else if value == "" {
			if err := t.service.AssignDayIndex(word.ID, nil); err != nil {
				return t.handleError(err)
			}
			t.status = fmt.Sprintf("Cleared day index for %q", word.Word)
		} else {
			day, err := strconv.Atoi(value)
			if err != nil {
				return t.handleError(fmt.Errorf("day index must be an integer or auto"))
			}
			if err := t.service.AssignDayIndex(word.ID, &day); err != nil {
				return t.handleError(err)
			}
			t.status = fmt.Sprintf("Assigned day %d to %q", day, word.Word)
		}
		if err := t.reload(); err != nil {
			return t.handleError(err)
		}
	}

	t.modal = modalState{}
	return t.refreshGUI(g)
}

func (t *TUI) closeModal(g *gocui.Gui, v *gocui.View) error {
	t.modal = modalState{}
	return t.refreshGUI(g)
}

func (t *TUI) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (t *TUI) focusModalField(g *gocui.Gui, index int) error {
	if len(t.modal.FieldView) == 0 {
		return nil
	}
	if index < 0 {
		index = 0
	}
	if index >= len(t.modal.FieldView) {
		index = len(t.modal.FieldView) - 1
	}
	t.modal.Active = index
	_, err := g.SetCurrentView(t.modal.FieldView[index])
	return err
}

func (t *TUI) clearModalViews(g *gocui.Gui) {
	t.clearFormViews(g)
	g.DeleteView(messageViewName)
}

func (t *TUI) clearFormViews(g *gocui.Gui) {
	g.DeleteView(modalFrameViewName)
	for i := 0; i < 8; i++ {
		g.DeleteView(fmt.Sprintf("modal_field_%d", i))
	}
}

func (t *TUI) refreshGUI(g *gocui.Gui) error {
	g.Update(func(*gocui.Gui) error { return nil })
	return nil
}

func (t *TUI) reload() error {
	words, err := t.service.ListWords(t.options)
	if err != nil {
		return err
	}
	t.words = words
	if len(words) == 0 {
		t.selected = 0
		t.selectedWordID = 0
		return nil
	}

	if t.selectedWordID != 0 {
		for idx, word := range words {
			if word.ID == t.selectedWordID {
				t.selected = idx
				return nil
			}
		}
	}

	if t.selected >= len(words) {
		t.selected = len(words) - 1
	}
	if t.selected < 0 {
		t.selected = 0
	}
	t.selectedWordID = words[t.selected].ID
	return nil
}

func (t *TUI) positionListCursor(v *gocui.View) error {
	if len(t.words) == 0 {
		return nil
	}
	maxX, maxY := v.Size()
	_ = maxX
	row := t.selected + 2
	visibleRows := maxY - 1
	originY := 0
	if row > visibleRows {
		originY = row - visibleRows
	}
	if err := v.SetOrigin(0, originY); err != nil {
		return err
	}
	return v.SetCursor(0, row-originY)
}

func (t *TUI) currentWord() *wotd.Word {
	if len(t.words) == 0 || t.selected < 0 || t.selected >= len(t.words) {
		return nil
	}
	return &t.words[t.selected]
}

func (t *TUI) handleError(err error) error {
	t.status = err.Error()
	return nil
}

func wordInputFromForm(values map[string]string) (WordInput, error) {
	input := WordInput{
		Word:             values["word"],
		Meaning:          values["meaning"],
		Link:             values["link"],
		Photo:            values["photo"],
		PhotoAttribution: values["photo_attribution"],
	}

	rawDay := strings.TrimSpace(values["day_index"])
	if rawDay == "" {
		return input, nil
	}
	if strings.EqualFold(rawDay, "auto") {
		input.AutoAssign = true
		return input, nil
	}

	day, err := strconv.Atoi(rawDay)
	if err != nil {
		return input, fmt.Errorf("day index must be an integer or auto")
	}
	input.DayIndex = &day
	return input, nil
}

func formatWordRow(word wotd.Word) string {
	return formatListRow(
		strconv.Itoa(word.ID),
		formatDay(word.DayIndex),
		word.Word,
		word.Meaning,
	)
}

func formatListRow(id, day, word, meaning string) string {
	return fmt.Sprintf("%-4s %-5s %-22s %-36s",
		truncate(id, 4),
		truncate(day, 5),
		truncate(word, 22),
		truncate(meaning, 36),
	)
}

func truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width == 1 {
		return string(runes[:1])
	}
	return string(runes[:width-1]) + "…"
}

func visibleWidth(value string) int {
	return utf8.RuneCountInString(value)
}

func formatDay(dayIndex *int) string {
	if dayIndex == nil {
		return "—"
	}
	return strconv.Itoa(*dayIndex)
}

func formatDayInput(dayIndex *int) string {
	if dayIndex == nil {
		return ""
	}
	return strconv.Itoa(*dayIndex)
}

func emptyPlaceholder(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}
