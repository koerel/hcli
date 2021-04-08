package main

import (
	"fmt"
	"log"
	"os"
	"sync"

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
		SetDescription("see the current time entries for today").
		SetShortDescription("timer status").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			status()
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

func status() {
	app := getApp()
	entries := app.c.GetEntriesToday(app.me.Id)
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Client", "Project", "Task", "Start", "End", "Running"})

	for _, e := range entries {
		t.AppendRows([]table.Row{
			{e.Client.Name, e.Project.Name, e.Task.Name, e.StartedTime, e.EndedTime, e.IsRunning},
		})
	}
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
	status()
}

func stop() {
	app := getApp()
	entries := app.c.GetRunningEntries(app.me.Id)
	if len(entries) > 0 {
		app.c.StopTimer(entries[0])
	}
	status()
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
