import express, { Request, Response } from "express";
import cors from "cors";

const app = express();
app.use(cors());
app.use(express.json());

const PORT = process.env.NOTIFICATION_PORT || 8082;
const LOG_LEVEL = process.env.LOG_LEVEL || "info";

interface Notification {
  id: string;
  type: string;
  message: string;
  channel: string;
  sentAt: number;
  status: string;
}

const notifications: Notification[] = [];

function log(level: string, message: string): void {
  if (level === "debug" && LOG_LEVEL !== "debug") return;
  const timestamp = new Date().toISOString();
  console.log(`${timestamp} [${level.toUpperCase()}] notification-service: ${message}`);
}

app.get("/health", (_req: Request, res: Response) => {
  res.json({
    status: "healthy",
    service: "notification-service",
    timestamp: Date.now(),
  });
});

app.post("/notifications/send", (req: Request, res: Response) => {
  const { type, message, channel } = req.body;

  if (!type || !message) {
    log("warn", "Invalid notification request: missing type or message");
    res.status(400).json({ error: "type and message are required" });
    return;
  }

  const notification: Notification = {
    id: `notif-${Date.now()}`,
    type,
    message,
    channel: channel || "default",
    sentAt: Date.now(),
    status: "sent",
  };

  notifications.push(notification);
  log("info", `Notification sent: ${notification.id} (type=${type}, channel=${notification.channel})`);

  res.status(201).json(notification);
});

app.get("/notifications", (_req: Request, res: Response) => {
  log("info", `Listing ${notifications.length} notifications`);
  res.json({ notifications, total: notifications.length });
});

app.get("/notifications/channels", (_req: Request, res: Response) => {
  const channels = [...new Set(notifications.map((n) => n.channel))];
  log("info", `Listing ${channels.length} channels`);
  res.json({ channels });
});

export { app };

if (require.main === module) {
  app.listen(PORT, () => {
    log("info", `Starting Notification Service on port ${PORT}`);
  });
}
