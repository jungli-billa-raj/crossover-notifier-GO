package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Crossover Notifier")

	w.SetContent(widget.NewLabel("Hello, welcome to Crossover Notifier"))
	w.ShowAndRun()
}

type Config struct {
	UserID    string `json:"user_id"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Token     string `json:"token"`
}
