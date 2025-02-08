import { createClient, fetchTasks } from "./notion.ts";
import { IncomingWebhook } from "npm:@slack/webhook";

const validateEnv = (): void => {
  const envs = [
    "NOTION_TOKEN",
    "NOTION_DATABASE_ID",
    "SLACK_WEBHOOK_URL",
  ] as const;

  for (const env of envs) {
    if (!Deno.env.get(env)) {
      throw new Error(`${env} is not provided in the environment variables`);
    }
  }
};

if (import.meta.main) {
  validateEnv();

  const client = createClient();
  const tasks = await fetchTasks(client);

  const webhook = new IncomingWebhook(Deno.env.get("SLACK_WEBHOOK_URL")!);
  await Promise.all([
    webhook.send({ text: JSON.stringify(tasks, null, 2) }),
    webhook.send({ text: tasks.map((task) => task.url).join("\n") }),
  ]);
}
