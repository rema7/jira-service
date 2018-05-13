package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/BurntSushi/toml"
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
	"encoding/json"
	"log"
)

type tomlConfig struct {
	Jira jiraConfig
}

type jiraConfig struct {
	Url string
	User  string
	Password  string
}

func readConfig() (tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile("settings.toml", &config); err != nil {
		fmt.Println(err)
		return tomlConfig{}, err
	}
	return config, nil
}

func jiraConnect(url string, user string, password string) (*jira.Client, error) {
	tp := jira.CookieAuthTransport{
		Username: user,
		Password: password,
		AuthURL:  fmt.Sprintf("%s/rest/auth/1/session", url),
	}
	return jira.NewClient(tp.Client(), url)
}

type JiraService interface {
	GetIssue(context.Context, string) (string, error)
}

type jiraService struct{}

func (jiraService) GetIssue(_ context.Context, issue string) (string, error) {
	if issue == "" {
		return "", ErrEmpty
	}

	return "XXXX---XXX", nil
}

func (jiraService) Count(_ context.Context, s string) int {
	return len(s)
}

var ErrEmpty = errors.New("empty string")

type issueRequest struct {
	Issue string `json:"issue"`
}

type issueResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

func getIssueEndpoint(js JiraService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(issueRequest)
		v, err := js.GetIssue(ctx, req.Issue)
		if err != nil {
			return issueResponse{v, err.Error()}, nil
		}
		return issueResponse{v, ""}, nil
	}
}

func main() {
	config, err := readConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	jiraConfig := config.Jira
	jiraClient, err := jiraConnect(jiraConfig.Url, jiraConfig.Password, jiraConfig.Password)

	if err != nil {
		fmt.Println(err)
		return
	}

	issue, resp, err := jiraClient.Issue.Get("RM-1", nil)
	if err != nil {
		fmt.Println(err)
		fmt.Println(resp)
		return
	}

	fmt.Printf("%s: %+v\n", issue.Key, issue.Fields.Summary)
	fmt.Printf("Type: %s\n", issue.Fields.Type.Name)
	fmt.Printf("Priority: %s\n", issue.Fields.Priority.Name)

	js := jiraService{}

	issueHandler := httptransport.NewServer(
		getIssueEndpoint(js),
		decodeIssueRequest,
		encodeResponse,
	)

	http.Handle("/uppercase", issueHandler)

	log.Fatal(http.ListenAndServe(":9090", nil))

}

func decodeIssueRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request issueRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
