import { Client, isFullPage } from "npm:@notionhq/client";
import { QueryDatabaseResponse } from "npm:@notionhq/client/build/src/api-endpoints";

const oneMonthAgo = new Date(new Date().getTime() - 30 * 24 * 60 * 60 * 1000);

type Task = {
  title: string;
  url: string;
  lastEditedTime: string;
};

export const createClient = (): Client =>
  new Client({
    auth: Deno.env.get("NOTION_TOKEN"),
  });

export const fetchTasks = (client: Client): Promise<Task[]> => {
  return queryDatabase(client).then(filter);
};

const queryDatabase = (client: Client): Promise<QueryDatabaseResponse> => {
  return client.databases.query({
    database_id: Deno.env.get("NOTION_DATABASE_ID")!,
    filter: {
      and: [
        {
          property: "Status",
          select: {
            equals: "Done",
          },
        },
        {
          property: "Last edited time",
          date: {
            after: oneMonthAgo.toISOString(),
          },
        },
      ],
    },
    sorts: [
      {
        property: "Last edited time",
        direction: "descending",
      },
    ],
  });
};

const filter = (QueryDatabaseResponse: QueryDatabaseResponse): Task[] => {
  return QueryDatabaseResponse.results.map((response) => {
    if (!isFullPage(response)) {
      throw new Error("Unexpected response");
    }

    const title = response.properties.Name.type === "title"
      ? response.properties.Name.title[0].plain_text
      : "";

    return {
      title,
      url: response.url,
      lastEditedTime: response.last_edited_time,
    };
  });
};
