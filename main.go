package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/BurntSushi/toml"
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

}
