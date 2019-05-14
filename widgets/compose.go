package widgets

import (
	"io"
	"io/ioutil"
	gomail "net/mail"
	"os"
	"os/exec"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"

	"git.sr.ht/~sircmpwn/aerc2/config"
	"git.sr.ht/~sircmpwn/aerc2/lib/ui"
)

type Composer struct {
	headers struct {
		from    *headerEditor
		subject *headerEditor
		to      *headerEditor
	}

	config *config.AccountConfig

	editor *Terminal
	email  *os.File
	grid   *ui.Grid
	review *reviewMessage

	focusable []ui.DrawableInteractive
	focused   int
}

// TODO: Let caller configure headers, initial body (for replies), etc
func NewComposer(conf *config.AccountConfig) *Composer {
	grid := ui.NewGrid().Rows([]ui.GridSpec{
		{ui.SIZE_EXACT, 3},
		{ui.SIZE_WEIGHT, 1},
	}).Columns([]ui.GridSpec{
		{ui.SIZE_WEIGHT, 1},
	})

	// TODO: let user specify extra headers to edit by default
	headers := ui.NewGrid().Rows([]ui.GridSpec{
		{ui.SIZE_EXACT, 1}, // To/From
		{ui.SIZE_EXACT, 1}, // Subject
		{ui.SIZE_EXACT, 1}, // [spacer]
	}).Columns([]ui.GridSpec{
		{ui.SIZE_WEIGHT, 1},
		{ui.SIZE_WEIGHT, 1},
	})

	to := newHeaderEditor("To", "")
	from := newHeaderEditor("From", conf.From)
	subject := newHeaderEditor("Subject", "")
	headers.AddChild(to).At(0, 0)
	headers.AddChild(from).At(0, 1)
	headers.AddChild(subject).At(1, 0).Span(1, 2)
	headers.AddChild(ui.NewFill(' ')).At(2, 0).Span(1, 2)

	email, err := ioutil.TempFile("", "aerc-compose-*.eml")
	if err != nil {
		// TODO: handle this better
		return nil
	}

	// TODO: built-in config option, $EDITOR, then vi, in that order
	editor := exec.Command("vim", email.Name())
	term, _ := NewTerminal(editor)

	grid.AddChild(headers).At(0, 0)
	grid.AddChild(term).At(1, 0)

	c := &Composer{
		config: conf,
		editor: term,
		email:  email,
		grid:   grid,
		// You have to backtab to get to "From", since you usually don't edit it
		focused:   1,
		focusable: []ui.DrawableInteractive{from, to, subject, term},
	}
	c.headers.to = to
	c.headers.from = from
	c.headers.subject = subject

	term.OnClose = c.termClosed

	return c
}

func (c *Composer) Draw(ctx *ui.Context) {
	c.grid.Draw(ctx)
}

func (c *Composer) Invalidate() {
	c.grid.Invalidate()
}

func (c *Composer) OnInvalidate(fn func(d ui.Drawable)) {
	c.grid.OnInvalidate(func(_ ui.Drawable) {
		fn(c)
	})
}

func (c *Composer) Close() {
	if c.email != nil {
		path := c.email.Name()
		c.email.Close()
		os.Remove(path)
		c.email = nil
	}
	if c.editor != nil {
		c.editor.Destroy()
		c.editor = nil
	}
}

func (c *Composer) Bindings() string {
	if c.editor == nil {
		return "compose::review"
	} else if c.editor == c.focusable[c.focused] {
		return "compose::editor"
	} else {
		return "compose"
	}
}

func (c *Composer) Event(event tcell.Event) bool {
	return c.focusable[c.focused].Event(event)
}

func (c *Composer) Focus(focus bool) {
	c.focusable[c.focused].Focus(focus)
}

func (c *Composer) Config() *config.AccountConfig {
	return c.config
}

func (c *Composer) Header() (*mail.Header, []string, error) {
	// Extract headers from the email, if present
	c.email.Seek(0, os.SEEK_SET)
	var (
		rcpts  []string
		header mail.Header
	)
	reader, err := mail.CreateReader(c.email)
	if err == nil {
		header = reader.Header
		defer reader.Close()
	} else {
		c.email.Seek(0, os.SEEK_SET)
	}
	// Update headers
	// TODO: Custom header fields
	mhdr := (*message.Header)(&header.Header)
	mhdr.SetContentType("text/plain", map[string]string{"charset": "UTF-8"})
	if subject, _ := header.Subject(); subject == "" {
		header.SetSubject(c.headers.subject.input.String())
	}
	if date, err := header.Date(); err != nil && date != (time.Time{}) {
		header.SetDate(time.Now())
	}
	if from, _ := mhdr.Text("From"); from == "" {
		mhdr.SetText("From", c.headers.from.input.String())
	}
	if to := c.headers.to.input.String(); to != "" {
		// Dammit Simon, this branch is 3x as long as it ought to be because
		// your types aren't compatible enough with each other
		to_rcpts, err := gomail.ParseAddressList(to)
		if err != nil {
			return nil, nil, err
		}
		ed_rcpts, err := header.AddressList("To")
		if err != nil {
			return nil, nil, err
		}
		for _, addr := range to_rcpts {
			ed_rcpts = append(ed_rcpts, (*mail.Address)(addr))
		}
		header.SetAddressList("To", ed_rcpts)
		for _, addr := range ed_rcpts {
			rcpts = append(rcpts, addr.Address)
		}
	}
	// TODO: Add cc, bcc to rcpts
	return &header, rcpts, nil
}

func (c *Composer) WriteMessage(header *mail.Header, writer io.Writer) error {
	c.email.Seek(0, os.SEEK_SET)
	var body io.Reader
	reader, err := mail.CreateReader(c.email)
	if err == nil {
		// TODO: Do we want to let users write a full blown multipart email
		// into the editor? If so this needs to change
		part, err := reader.NextPart()
		if err != nil {
			return err
		}
		body = part.Body
		defer reader.Close()
	} else {
		c.email.Seek(0, os.SEEK_SET)
		body = c.email
	}
	// TODO: attachments
	w, err := mail.CreateSingleInlineWriter(writer, *header)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, body)
	return err
}

func (c *Composer) termClosed(err error) {
	// TODO: do we care about that error (note: yes, we do)
	c.grid.RemoveChild(c.editor)
	c.grid.AddChild(newReviewMessage(c)).At(1, 0)
	c.editor.Destroy()
	c.editor = nil
}

func (c *Composer) PrevField() {
	c.focusable[c.focused].Focus(false)
	c.focused--
	if c.focused == -1 {
		c.focused = len(c.focusable) - 1
	}
	c.focusable[c.focused].Focus(true)
}

func (c *Composer) NextField() {
	c.focusable[c.focused].Focus(false)
	c.focused = (c.focused + 1) % len(c.focusable)
	c.focusable[c.focused].Focus(true)
}

type headerEditor struct {
	name  string
	input *ui.TextInput
}

func newHeaderEditor(name string, value string) *headerEditor {
	return &headerEditor{
		input: ui.NewTextInput(value),
		name:  name,
	}
}

func (he *headerEditor) Draw(ctx *ui.Context) {
	name := he.name + " "
	size := runewidth.StringWidth(name)
	ctx.Fill(0, 0, size, ctx.Height(), ' ', tcell.StyleDefault)
	ctx.Printf(0, 0, tcell.StyleDefault.Bold(true), "%s", name)
	he.input.Draw(ctx.Subcontext(size, 0, ctx.Width()-size, 1))
}

func (he *headerEditor) Invalidate() {
	he.input.Invalidate()
}

func (he *headerEditor) OnInvalidate(fn func(ui.Drawable)) {
	he.input.OnInvalidate(func(_ ui.Drawable) {
		fn(he)
	})
}

func (he *headerEditor) Focus(focused bool) {
	he.input.Focus(focused)
}

func (he *headerEditor) Event(event tcell.Event) bool {
	return he.input.Event(event)
}

type reviewMessage struct {
	composer *Composer
	grid     *ui.Grid
}

func newReviewMessage(composer *Composer) *reviewMessage {
	grid := ui.NewGrid().Rows([]ui.GridSpec{
		{ui.SIZE_EXACT, 2},
		{ui.SIZE_EXACT, 1},
		{ui.SIZE_WEIGHT, 1},
	}).Columns([]ui.GridSpec{
		{ui.SIZE_WEIGHT, 1},
	})
	grid.AddChild(ui.NewText(
		"Send this email? [y]es/[n]o/[e]dit/[a]ttach file")).At(0, 0)
	grid.AddChild(ui.NewText("Attachments:").
		Reverse(true)).At(1, 0)
	// TODO: Attachments
	grid.AddChild(ui.NewText("(none)")).At(2, 0)

	return &reviewMessage{
		composer: composer,
		grid:     grid,
	}
}

func (rm *reviewMessage) Invalidate() {
	rm.grid.Invalidate()
}

func (rm *reviewMessage) OnInvalidate(fn func(ui.Drawable)) {
	rm.grid.OnInvalidate(func(_ ui.Drawable) {
		fn(rm)
	})
}

func (rm *reviewMessage) Draw(ctx *ui.Context) {
	rm.grid.Draw(ctx)
}
