package main

import "time"

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func weekStartEnd() (string, string) {
	ti := time.Now()
	y, m, d := ti.Date()
	da := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	to := da.Format("2006-01-02")
	for da.Weekday() != time.Monday {
		da = da.AddDate(0, 0, -1)
	}
	from := da.Format("2006-01-02")
	return from, to
}

func getEntry(taskId int, entries []TimeEntry) TimeEntry {
	for _, e := range entries {
		if e.Task.Id == taskId {
			return e
		}
	}
	return TimeEntry{}
}
