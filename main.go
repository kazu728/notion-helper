package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/dstotijn/go-notion"
)

type NotionConfig struct {
	ApiKey               string `env:"API_KEY"`
	Databaseid           string `env:"DATABASE_ID"`
	MovingProperty       string `env:"MOVING_PROPERTY"`
	MovingColumnBefore   string `env:"MOVING_COLUMN_BEFORE"`
	MovingColumnAfter    string `env:"MOVING_COLUMN_AFTER"`
	DaysBeforeTaskMoving int    `env:"DAYS_BEFORE_TASK_MOVING"`
}

type Result struct {
	Id  string
	Err error
}

func main() {

	if err := Run(); err != nil {
		panic(err)
	}

	os.Exit(0)
}

func Run() error {
	config := NotionConfig{}

	if err := env.Parse(&config, env.Options{RequiredIfNoDef: true}); err != nil {
		panic(err)
	}

	n := NewNotion(config)
	tasks, err := n.getTasks()

	if err != nil {
		return err
	}

	canMovingTasks := tasks.canMovingTasks(n.config.DaysBeforeTaskMoving)
	results := make(chan Result, len(canMovingTasks))

	for _, task := range canMovingTasks {
		go func(id string) {
			if err := n.updatePage(context.Background(), id); err != nil {
				panic(err)
			}
			results <- Result{Id: id, Err: err}
		}(task.Id)
	}

	for range canMovingTasks {
		<-results
	}

	return nil
}

type Task struct {
	Id             string    `json:"id"`
	LastEditedTime time.Time `json:"last_edited_time"`
}

func (t *Task) canMoving(days int) bool {
	return time.Now().After(t.LastEditedTime.AddDate(0, 0, days))
}

type Tasks []Task

func (t *Tasks) canMovingTasks(days int) Tasks {
	resp := make(Tasks, 0, len(*t))
	for _, v := range *t {
		if v.canMoving(days) {
			resp = append(resp, v)
		}
	}
	return resp
}

type Notion struct {
	client *notion.Client
	config NotionConfig
}

func NewNotion(c NotionConfig) Notion {

	return Notion{client: notion.NewClient(c.ApiKey), config: c}
}

func (n *Notion) getTasks() (Tasks, error) {
	queryResponse, err := n.client.QueryDatabase(context.Background(), n.config.Databaseid, &notion.DatabaseQuery{
		Filter: &notion.DatabaseQueryFilter{
			Property: n.config.MovingProperty,
			DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
				Select: &notion.SelectDatabaseQueryFilter{
					Equals: n.config.MovingColumnBefore,
				},
			},
		},
	})

	if err != nil {
		return nil, err
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
	updatedProps[n.config.MovingProperty] = notion.DatabasePageProperty{
		Select: &notion.SelectOptions{
			Name: n.config.MovingColumnAfter,
		},
	}
	_, err := n.client.UpdatePage(ctx, pageID, notion.UpdatePageParams{
		DatabasePageProperties: updatedProps,
	})
	if err != nil {
		return err
	}

	return nil
}
