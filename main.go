package main

import (
	"path/filepath"

	"github.com/google/uuid"
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
)

func main() {
	app.Import(&Menu{})

	app.Run(&mac.Driver{
		Bundle: mac.Bundle{
			AppName:    "repwpm", //r/earthporn wallpaper menu
			Version:    "0.0.1",
			Background: true,
			Icon:       "icon.png",
		},

		OnRun: func() {
			app.NewStatusMenu(app.StatusMenuConfig{
				// Icon:       app.Resources("logo.png"),
				Text:       "🖥", //⛩
				DefaultURL: "/Menu",
			})
		},
	}, app.Logs())
}

// Menu is a component that describes a status able to change its text and icon.
type Menu struct {
	IconHidden bool
	TextHidden bool
}

// Render returns the HTML describing the status menu.
func (m *Menu) Render() string {
	return `
<menu>
	<menuitem label="Get new wallpapers" onclick="OnGetNew"></menuitem>
	<menuitem label="Next wallpaper" onclick="OnNext"></menuitem>
	<menuitem separator></menuitem>
	<menuitem label="Quit" selector="terminate:"></menuitem>
</menu>
	`
}

//	<menuitem label="Hide text" onclick="OnHideText" {{if .TextHidden}}disabled{{end}}></menuitem>

// OnShowIcon is the function called when the show icon button is clicked.
func (m *Menu) OnShowIcon() {
	statMenu, err := app.StatusMenuByComponent(m)
	if err != nil {
		return
	}

	statMenu.SetIcon(app.Resources("icon.png"))
	m.IconHidden = false
	app.Render(m)
}

// OnGetNew is the function called when the get new wallpapers button is clicked.
func (m *Menu) OnGetNew() {
	total, result := downloadNewWallpapers()
	if total > 0 {
		app.NewNotification(app.NotificationConfig{
			Title: "r/EarthPorn Wallpapers",
			Text:  result,
			Sound: false,
		})
	} else {
		app.NewNotification(app.NotificationConfig{
			Title: "r/EarthPorn Wallpapers",
			Text:  "No new wallpapers",
			Sound: false,
		})
	}
}

// OnHideIcon is the function called when the hide icon button is clicked.
func (m *Menu) OnHideIcon() {
	statMenu, err := app.StatusMenuByComponent(m)
	if err != nil {
		return
	}

	statMenu.SetIcon("")
	m.IconHidden = true
	app.Render(m)

	if m.TextHidden {
		m.OnShowText()
	}
}

// OnShowText is the function called when the show text button is clicked.
func (m *Menu) OnShowText() {
	app.NewNotification(app.NotificationConfig{
		Title:     "hello",
		Subtitle:  "world",
		Text:      uuid.New().String(),
		ImageName: filepath.Join(app.Resources(), "icon.png"),
		Sound:     false,
	})
}

// OnHideText is the function called when the hide text button is clicked.
func (m *Menu) OnHideText() {
	statMenu, err := app.StatusMenuByComponent(m)
	if err != nil {
		return
	}

	statMenu.SetText("")
	m.TextHidden = true
	app.Render(m)

	if m.IconHidden {
		m.OnShowIcon()
	}
}
