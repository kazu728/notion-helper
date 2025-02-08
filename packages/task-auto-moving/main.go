package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/dstotijn/go-notion"
	"gopkg.in/yaml.v3"
)

type Result struct {
	ID  string
	Err error
}

type Task struct {
	Id             string    `json:"id"`
	LastEditedTime time.Time `json:"last_edited_time"`
}

type Tasks []Task

type NotionConfig struct {
	APIKey     string `env:"API_KEY"`
	DatabaseID string `env:"DATABASE_ID"`
}

type Setting struct {
	Action Action `yaml:"action"`
}

type Action struct {
	Move MoveAction `yaml:"move"`
}

type MoveAction struct {
	Property PropertyAction `yaml:"property"`
}

type PropertyAction struct {
	Name          string `yaml:"name"`
	From          string `yaml:"from"`
	To            string `yaml:"to"`
	ExpiresInDays int    `yaml:"expires_in_days"`
}

type Notion struct {
	client *notion.Client
	config NotionConfig
	action Action
}

var wg sync.WaitGroup

func main() {
	if err := run(); err != nil {
		panic(err)
	}

	os.Exit(0)
}

func run() error {
	n, err := newNotion()
	if err != nil {
		return err
	}

	tasks, err := n.getTasks()
	if err != nil {
		return err
	}

	canMovingTasks := tasks.canMovingTasks(n.action.Move.Property.ExpiresInDays)
	results := make(chan Result, len(canMovingTasks))

	for _, task := range canMovingTasks {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := n.updatePage(context.Background(), id); err != nil {
				panic(err)
			}
			results <- Result{ID: id, Err: err}
		}(task.Id)
	}

	wg.Wait()
	close(results)

	for result := range results {
		if result.Err != nil {
			log.Printf("Error processing task %s: %v", result.ID, result.Err)
		}
	}

	return nil
}

func (t *Task) canMoving(days int) bool {
	return time.Now().After(t.LastEditedTime.AddDate(0, 0, days))
}

func (t Tasks) canMovingTasks(days int) Tasks {
	resp := make(Tasks, 0, len(t))
	for _, v := range t {
		if v.canMoving(days) {
			resp = append(resp, v)
		}
	}
	return resp
}

func newNotion() (Notion, error) {
	config := NotionConfig{}

	if err := env.Parse(&config, env.Options{RequiredIfNoDef: true}); err != nil {
		return Notion{}, err
	}

	data, err := os.ReadFile("./setting.yml")
	if err != nil {
		return Notion{}, err
	}

	var setting Setting
	if err = yaml.Unmarshal(data, &setting); err != nil {
		return Notion{}, err
	}

	return Notion{client: notion.NewClient(config.APIKey), config: config, action: setting.Action}, nil
}

func (n *Notion) getTasks() (Tasks, error) {
	queryResponse, err := n.client.QueryDatabase(context.Background(), n.config.DatabaseID, &notion.DatabaseQuery{
		Filter: &notion.DatabaseQueryFilter{
			Property: n.action.Move.Property.Name,
			DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
				Select: &notion.SelectDatabaseQueryFilter{
					Equals: n.action.Move.Property.From,
				},
			},
		},
	})

	if err != nil {
		return Tasks{}, err
	}

	resp := make(Tasks, 0, len(queryResponse.Results))
	for _, v := range queryResponse.Results {
		resp = append(resp, Task{
			Id:             v.ID,
			LastEditedTime: v.LastEditedTime,
		})
	}

	return resp, nil
}

func (n *Notion) updatePage(ctx context.Context, pageID string) error {
	updatedProps := make(notion.DatabasePageProperties)

	log.Printf("Updating. ID:  %s \n", pageID)
	updatedProps[n.action.Move.Property.Name] = notion.DatabasePageProperty{
		Select: &notion.SelectOptions{
			Name: n.action.Move.Property.To,
		},
	}

	_, err := n.client.UpdatePage(ctx, pageID, notion.UpdatePageParams{
		DatabasePageProperties: updatedProps,
	})

	return err
}
