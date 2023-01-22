# Staled Notion task actions

[![Testing with the private action](https://github.com/kazu728/staled-notion-task-actions/actions/workflows/master.yml/badge.svg)](https://github.com/kazu728/staled-notion-task-actions/actions/workflows/master.yml)

For moving staled tasks to other column such as `Backlog` column -> `Staled` column


## Requirements
- Notion API key
- Connecting with integration for Database


## example

```yml
jobs:
  run:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Notion task handle actions
        uses: kazu728/staled-notion-task-actions@v0.0.1
        env:
          API_KEY: ${{ secrets.API_KEY }}
          DATABASE_ID: ${{ secrets.DATABASE_ID }}
          MOVING_PROPERTY: ${{ secrets.MOVING_PROPERTY }}
          MOVING_COLUMN_BEFORE: ${{ secrets.MOVING_COLUMN_BEFORE }}
          MOVING_COLUMN_AFTER: ${{ secrets.MOVING_COLUMN_AFTER }}
          DAYS_BEFORE_TASK_MOVING: ${{ secrets.DAYS_BEFORE_TASK_MOVING }}
```