package widgets

import (
	"log"
	"sort"

	"github.com/gdamore/tcell"

	"git.sr.ht/~sircmpwn/aerc/config"
	"git.sr.ht/~sircmpwn/aerc/lib/ui"
	"git.sr.ht/~sircmpwn/aerc/worker/types"
)

type DirectoryList struct {
	ui.Invalidatable
	acctConf  *config.AccountConfig
	uiConf    *config.UIConfig
	dirs      []string
	logger    *log.Logger
	selecting string
	selected  string
	spinner   *Spinner
	worker    *types.Worker
}

func NewDirectoryList(acctConf *config.AccountConfig, uiConf *config.UIConfig,
	logger *log.Logger, worker *types.Worker) *DirectoryList {

	dirlist := &DirectoryList{
		acctConf: acctConf,
		uiConf:   uiConf,
		logger:   logger,
		spinner:  NewSpinner(),
		worker:   worker,
	}
	dirlist.spinner.OnInvalidate(func(_ ui.Drawable) {
		dirlist.Invalidate()
	})
	dirlist.spinner.Start()
	return dirlist
}

func (dirlist *DirectoryList) UpdateList(done func(dirs []string)) {
	var dirs []string
	dirlist.worker.PostAction(
		&types.ListDirectories{}, func(msg types.WorkerMessage) {

			switch msg := msg.(type) {
			case *types.Directory:
				dirs = append(dirs, msg.Name)
			case *types.Done:
				sort.Strings(dirs)
				dirlist.dirs = dirs
				dirlist.spinner.Stop()
				dirlist.Invalidate()
				if done != nil {
					done(dirs)
				}
			}
		})
}

func (dirlist *DirectoryList) Select(name string) {
	dirlist.selecting = name
	dirlist.worker.PostAction(&types.OpenDirectory{Directory: name},
		func(msg types.WorkerMessage) {
			switch msg.(type) {
			case *types.Error:
				dirlist.selecting = ""
			case *types.Done:
				dirlist.selected = dirlist.selecting
				dirlist.filterDirsByFoldersConfig()
				hasSelected := false
				for _, d := range dirlist.dirs {
					if d == dirlist.selected {
						hasSelected = true
						break
					}
				}
				if !hasSelected && dirlist.selected != "" {
					dirlist.dirs = append(dirlist.dirs, dirlist.selected)
				}
				sort.Strings(dirlist.dirs)
			}
			dirlist.Invalidate()
		})
	dirlist.Invalidate()
}

func (dirlist *DirectoryList) Selected() string {
	return dirlist.selected
}

func (dirlist *DirectoryList) Invalidate() {
	dirlist.DoInvalidate(dirlist)
}

func (dirlist *DirectoryList) Draw(ctx *ui.Context) {
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', tcell.StyleDefault)

	if dirlist.spinner.IsRunning() {
		dirlist.spinner.Draw(ctx)
		return
	}

	if len(dirlist.dirs) == 0 {
		style := tcell.StyleDefault
		ctx.Printf(0, 0, style, dirlist.uiConf.EmptyDirlist)
		return
	}

	row := 0
	for _, name := range dirlist.dirs {
		if row >= ctx.Height() {
			break
		}
		if len(dirlist.acctConf.Folders) > 1 && name != dirlist.selected {
			idx := sort.SearchStrings(dirlist.acctConf.Folders, name)
			if idx == len(dirlist.acctConf.Folders) ||
				dirlist.acctConf.Folders[idx] != name {
				continue
			}
		}
		style := tcell.StyleDefault
		if name == dirlist.selected {
			style = style.Reverse(true)
		}
		ctx.Fill(0, row, ctx.Width(), 1, ' ', style)
		ctx.Printf(0, row, style, "%s", name)
		row++
	}
}

func (dirlist *DirectoryList) nextPrev(delta int) {
	for i, dir := range dirlist.dirs {
		if dir == dirlist.selected {
			var j int
			ndirs := len(dirlist.dirs)
			for j = i + delta; j != i; j += delta {
				if j < 0 {
					j = ndirs - 1
				}
				if j >= ndirs {
					j = 0
				}
				name := dirlist.dirs[j]
				if len(dirlist.acctConf.Folders) > 1 && name != dirlist.selected {
					idx := sort.SearchStrings(dirlist.acctConf.Folders, name)
					if idx == len(dirlist.acctConf.Folders) ||
						dirlist.acctConf.Folders[idx] != name {

						continue
					}
				}
				break
			}
			dirlist.Select(dirlist.dirs[j])
			break
		}
	}
}

func (dirlist *DirectoryList) Next() {
	dirlist.nextPrev(1)
}

func (dirlist *DirectoryList) Prev() {
	dirlist.nextPrev(-1)
}

// filterDirsByFoldersConfig filters a folders list to only contain folders
// present in the account.folders config option
func (dirlist *DirectoryList) filterDirsByFoldersConfig() {
	// config option defaults to show all if unset
	if len(dirlist.acctConf.Folders) == 0 {
		return
	}
	var filtered []string
	for _, folder := range dirlist.dirs {
		for _, cfgfolder := range dirlist.acctConf.Folders {
			if folder == cfgfolder {
				filtered = append(filtered, folder)
				break
			}
		}
	}
	dirlist.dirs = filtered
}
