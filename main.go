package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/thatisuday/commando"
)

var once sync.Once

type App struct {
	c  *HarvestClient
	me *User
}

var app App

func main() {
	commando.SetExecutableName("hcli").
		SetVersion("v1.0.0").
		SetDescription(`Run Harvest from the CLI

Expects env vars HARVEST_API_TOKEN & HARVEST_ACCOUNT_ID
Retrieve them at https://id.getharvest.com/`)

	commando.Register("status").
		SetDescription("see the time entries for a chosen date - defaults to today").
		SetShortDescription("timer status").
		AddFlag("date,d", "date parameter in format YYYY-MM-DD", commando.String, time.Now().Format("2006-01-02")).
		AddFlag("week,w", "get week totals", commando.Bool, nil).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			date, err := flags["date"].GetString()
			handle(err)
			week, err := flags["week"].GetBool()
			handle(err)
			if week {
				statusWeek()
			} else {
				status(date)
			}
		})

	commando.Register("start").
		SetDescription("start a new time").
		SetShortDescription("start a new timer").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			start()
		})

	commando.Register("stop").
		SetDescription("stop the running timer").
		SetShortDescription("stop the running timer").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			stop()
		})

	commando.Parse(nil)
}

func getApp() *App {
	once.Do(func() {
		app = App{
			c: NewHarvestClient(os.Getenv("HARVEST_API_TOKEN"), os.Getenv("HARVEST_ACCOUNT_ID")),
		}
		app.me = app.c.GetMe()
	})
	return &app
}

func status(date string) {
	app := getApp()
	entries := app.c.GetEntries(date, date, app.me.Id)
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Client", "Project", "Task", "Start", "End", "Duration", "Running"})
	var total float64
	for _, e := range entries {
		var duration time.Duration
		var start time.Time
		var end time.Time
		var err error
		running := "*"
		if !e.IsRunning {
			running = ""
			end, err = time.Parse("2006-01-02 15:04", date+" "+e.EndedTime)
			handle(err)
			start, err = time.Parse("2006-01-02 15:04", date+" "+e.StartedTime)
			handle(err)
		} else {
			start, err = time.Parse("2006-01-02 15:04", date+" "+e.StartedTime)
			handle(err)
			end, _ = time.Parse("2006-01-02 15:04", time.Now().Format("2006-01-02 15:04"))
		}
		duration = end.Sub(start)
		total += duration.Seconds()
		t.AppendRows([]table.Row{
			{e.Client.Name, e.Project.Name, e.Task.Name, e.StartedTime, e.EndedTime, duration, running},
		})
	}
	t.AppendSeparator()
	totalTime, err := time.ParseDuration(fmt.Sprintf("%f", total) + "s")
	handle(err)
	t.AppendRows([]table.Row{
		{"Total", "", "", "", "", totalTime, ""},
	})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}

func statusWeek() {
	app := getApp()
	ti := time.Now()
	y, m, d := ti.Date()
	da := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	to := da.Format("2006-01-02")
	for da.Weekday() != time.Monday {
		da = da.AddDate(0, 0, -1)
	}
	from := da.Format("2006-01-02")
	entries := app.c.GetEntries(from, to, app.me.Id)
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Client", "Project", "Task", "Date", "Start", "End", "Duration", "Running"})
	var total float64
	for i, e := range entries {
		var duration time.Duration
		var start time.Time
		var end time.Time
		var err error
		running := "*"
		if i > 0 && e.SpentDate != entries[i-1].SpentDate {
			t.AppendSeparator()
		}
		if !e.IsRunning {
			running = ""
			end, err = time.Parse("2006-01-02 15:04", e.SpentDate+" "+e.EndedTime)
			handle(err)
			start, err = time.Parse("2006-01-02 15:04", e.SpentDate+" "+e.StartedTime)
			handle(err)
		} else {
			start, err = time.Parse("2006-01-02 15:04", e.SpentDate+" "+e.StartedTime)
			handle(err)
			end, _ = time.Parse("2006-01-02 15:04", time.Now().Format("2006-01-02 15:04"))
		}
		duration = end.Sub(start)
		total += duration.Seconds()
		t.AppendRows([]table.Row{
			{e.Client.Name, e.Project.Name, e.Task.Name, e.SpentDate, e.StartedTime, e.EndedTime, duration, running},
		})
	}
	t.AppendSeparator()
	totalTime, err := time.ParseDuration(fmt.Sprintf("%f", total) + "s")
	handle(err)
	t.AppendRows([]table.Row{
		{"Total", "", "", "", "", "", totalTime, ""},
	})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}

func start() {
	app := getApp()
	as := app.c.GetAssignments(app.me.Id)
	idx, err := fuzzyfinder.FindMulti(
		as.Assignments,
		func(i int) string {
			return as.Assignments[i].Client.Name + " " + as.Assignments[i].Project.Name
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Client: %s\nProject: %s",
				as.Assignments[i].Client.Name,
				as.Assignments[i].Project.Name)
		}))
	if err != nil {
		log.Fatal(err)
	}
	assignment := as.Assignments[idx[0]]
	tasks := assignment.TaskAssignments
	idx, err = fuzzyfinder.FindMulti(
		tasks,
		func(i int) string {
			return tasks[i].Task.Name
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Task: %s",
				tasks[i].Task.Name)
		}))
	if err != nil {
		log.Fatal(err)
	}
	task := tasks[idx[0]].Task
	app.c.StartTimer(assignment.Project, task)
	status(time.Now().Format("2006-01-02"))
}

func stop() {
	app := getApp()
	entries := app.c.GetRunningEntries(app.me.Id)
	if len(entries) > 0 {
		app.c.StopTimer(entries[0])
	}
	status(time.Now().Format("2006-01-02"))
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
