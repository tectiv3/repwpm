package main

import (
	"time"

	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
)

func main() {
	app.Import(&Menu{})

	app.Run(&mac.Driver{
		Bundle: mac.Bundle{
			AppName:          "repwpm", //r/earthporn wallpaper menu
			Version:          "0.0.3",
			Background:       true,
			Icon:             "icon.png",
			DeploymentTarget: "10.11",
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
	TimerEnabled bool
	ticker       *time.Ticker
}

// Render returns the HTML describing the status menu.
func (m *Menu) Render() string {
	return `
<menu>
	<menuitem label="Get new wallpapers" onclick="OnGetNew"></menuitem>
	<menuitem label="Next wallpaper" onclick="OnNext"></menuitem>
	<menuitem separator></menuitem>
	<menuitem label="Enable timer (1H)" onclick="OnEnableTimer" {{if .TimerEnabled}}disabled{{end}}></menuitem>
	<menuitem label="Disable timer" onclick="OnDisableTimer" {{if not .TimerEnabled}}disabled{{end}}></menuitem>
	<menuitem separator></menuitem>
	<menuitem label="Quit" selector="terminate:"></menuitem>
</menu>
	`
}

// OnGetNew is the function called when the get new wallpapers button is clicked.
func (m *Menu) OnGetNew() {
	getWallpapers(true)
}

// OnNext called when next wallpapers button clicked
func (m *Menu) OnNext() {
	if err := nextWallpaper(); err != nil {
		app.Log(err.Error())
	}
}

// OnEnableTimer called when enable timer button clicked
func (m *Menu) OnEnableTimer() {
	if m.TimerEnabled {
		return
	}
	app.Log("Enabling timer")
	m.TimerEnabled = true
	m.ticker = time.NewTicker(time.Hour * 1)
	go func() {
		for {
			select {
			case <-m.ticker.C:
				getWallpapers(true)
			}
		}
	}()
	app.Render(m)
}

// OnDisableTimer called when disable time button clicked
func (m *Menu) OnDisableTimer() {
	if !m.TimerEnabled {
		return
	}
	m.ticker.Stop()
	m.TimerEnabled = false
	app.Render(m)
}
