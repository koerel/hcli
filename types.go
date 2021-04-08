package main

type User struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Project struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Assignment struct {
	Id              int              `json:"id"`
	Project         Project          `json:"project"`
	Client          Client           `json:"client"`
	TaskAssignments []TaskAssignment `json:"task_assignments"`
}

type AssignmentResponse struct {
	Assignments []Assignment `json:"project_assignments"`
}

type Task struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type TaskAssignment struct {
	Id   int  `json:"id"`
	Task Task `json:"task"`
}

type TimeEntryInput struct {
	ProjectId int    `json:"project_id"`
	TaskId    int    `json:"task_id"`
	SpentDate string `json:"spent_date"`
}

type TimeEntry struct {
	Id          int     `json:"id"`
	SpentDate   string  `json:"spent_date"`
	Client      Client  `json:"client"`
	Project     Project `json:"project"`
	Task        Task    `json:"task"`
	StartedTime string  `json:"started_time"`
	EndedTime   string  `json:"ended_time"`
	IsRunning   bool    `json:"is_running"`
}

type TimeEntryResponse struct {
	TimeEntries []TimeEntry `json:"time_entries"`
}
