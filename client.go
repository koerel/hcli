package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type HarvestClient struct {
	token     string
	account   string
	baseUrl   string
	debugMode bool
	client    *http.Client
}

func NewHarvestClient(token string, account string) *HarvestClient {
	c := new(HarvestClient)
	c.token = token
	c.account = account
	c.baseUrl = "https://api.harvestapp.com/v2"
	c.client = &http.Client{}
	c.debugMode = false

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
	res, err := c.Post("/time_entries", string(data))
	handle(err)
	body, err := ioutil.ReadAll(res.Body)
	handle(err)
	c.debug(string(body))
}

func (c *HarvestClient) StopTimer(e TimeEntry) {
	_, err := c.Patch("/time_entries/" + strconv.Itoa(e.Id) + "/stop")
	handle(err)
}

func (c *HarvestClient) GetEntriesToday(userId int) []TimeEntry {
	time := time.Now()
	date := time.Format("2006-01-02")
	return c.GetEntries(date, date, userId)
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

func (c *HarvestClient) GetEntries(dateFrom string, dateTo string, userId int) []TimeEntry {
	q := url.Values{}
	q.Add("user_id", strconv.Itoa(userId))
	q.Add("from", dateFrom)
	q.Add("to", dateTo)
	res, err := c.Get("/time_entries?" + q.Encode())
	handle(err)
	er := new(TimeEntryResponse)
	body, err := ioutil.ReadAll(res.Body)
	handle(err)
	c.debug(string(body))
	json.Unmarshal(body, &er)

	return er.TimeEntries
}

func (c *HarvestClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseUrl+url, nil)
	handle(err)
	c.setAuthHeaders(req)
	res, err := c.client.Do(req)

	return res, err
}

func (c *HarvestClient) Patch(url string) (*http.Response, error) {
	req, err := http.NewRequest("PATCH", c.baseUrl+url, nil)
	handle(err)
	c.setAuthHeaders(req)
	res, err := c.client.Do(req)

	return res, err
}

func (c *HarvestClient) Post(url string, body string) (*http.Response, error) {
	r := strings.NewReader(body)
	req, err := http.NewRequest("POST", c.baseUrl+url, r)
	handle(err)
	c.setAuthHeaders(req)
	c.setContentType(req)
	res, err := c.client.Do(req)

	return res, err
}

func (c *HarvestClient) setAuthHeaders(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Harvest-Account-ID", c.account)
}

func (c *HarvestClient) setContentType(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
}

func (c *HarvestClient) debug(data interface{}) {
	if c.debugMode {
		fmt.Println(data)
	}
}
