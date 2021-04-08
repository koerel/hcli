package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type HarvestClient struct {
	token   string
	account string
	baseUrl string
	client  *http.Client
}

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

func NewHarvestClient(token string, account string) *HarvestClient {
	c := new(HarvestClient)
	c.token = token
	c.account = account
	c.baseUrl = "https://api.harvestapp.com/v2"
	c.client = &http.Client{}

	return c
}

func (c *HarvestClient) GetMe() *User {
	res, err := c.Get("/users/me")
	handle(err)
	user := new(User)
	body, err := ioutil.ReadAll(res.Body)
	handle(err)
	json.Unmarshal(body, &user)

	return user
}

func (c *HarvestClient) GetAssignments(userId int) *AssignmentResponse {
	res, err := c.Get("/users/" + strconv.Itoa(userId) + "/project_assignments")
	handle(err)
	ar := new(AssignmentResponse)
	body, err := ioutil.ReadAll(res.Body)
	handle(err)
	json.Unmarshal(body, &ar)

	return ar
}

func (c *HarvestClient) StartTimer(p Project, t Task) {
	time := time.Now()
	date := time.Format("2006-01-02")
	entry := TimeEntryInput{
		ProjectId: p.Id,
		TaskId:    t.Id,
		SpentDate: date,
	}
	data, err := json.Marshal(entry)
	handle(err)
	_, err = c.Post("/time_entries", string(data))
	handle(err)
}

func (c *HarvestClient) StopTimer(e TimeEntry) {
	_, err := c.Patch("/time_entries/" + strconv.Itoa(e.Id) + "/stop")
	handle(err)
}

func (c *HarvestClient) GetEntriesToday(userId int) []TimeEntry {
	time := time.Now()
	date := time.Format("2006-01-02")
	return c.GetEntries(date, userId)
}

func (c *HarvestClient) GetRunningEntries(userId int) []TimeEntry {
	q := url.Values{}
	q.Add("user_id", strconv.Itoa(userId))
	q.Add("is_running", "true")
	res, err := c.Get("/time_entries?" + q.Encode())
	handle(err)
	er := new(TimeEntryResponse)
	body, err := ioutil.ReadAll(res.Body)
	handle(err)
	json.Unmarshal(body, &er)

	return er.TimeEntries
}

func (c *HarvestClient) GetEntries(date string, userId int) []TimeEntry {
	q := url.Values{}
	q.Add("user_id", strconv.Itoa(userId))
	q.Add("from", date)
	q.Add("to", date)
	res, err := c.Get("/time_entries?" + q.Encode())
	handle(err)
	er := new(TimeEntryResponse)
	body, err := ioutil.ReadAll(res.Body)
	handle(err)
	json.Unmarshal(body, &er)

	return er.TimeEntries
}

func (c *HarvestClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseUrl+url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Harvest-Account-ID", c.account)
	res, err := c.client.Do(req)

	return res, err
}

func (c *HarvestClient) Patch(url string) (*http.Response, error) {
	req, err := http.NewRequest("PATCH", c.baseUrl+url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Harvest-Account-ID", c.account)
	res, err := c.client.Do(req)

	return res, err
}

func (c *HarvestClient) Post(url string, body string) (*http.Response, error) {
	r := strings.NewReader(body)
	req, err := http.NewRequest("POST", c.baseUrl+url, r)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Harvest-Account-ID", c.account)
	req.Header.Add("Content-Type", "application/json")
	res, err := c.client.Do(req)

	return res, err
}
