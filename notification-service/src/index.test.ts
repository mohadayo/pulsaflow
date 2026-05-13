import request from "supertest";
import { app } from "./index";

describe("Notification Service", () => {
  describe("GET /health", () => {
    it("should return healthy status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.status).toBe("healthy");
      expect(res.body.service).toBe("notification-service");
      expect(res.body.timestamp).toBeDefined();
    });
  });

  describe("POST /notifications/send", () => {
    it("should send a notification successfully", async () => {
      const res = await request(app)
        .post("/notifications/send")
        .send({ type: "workflow_created", message: "Test notification", channel: "email" });
      expect(res.status).toBe(201);
      expect(res.body.type).toBe("workflow_created");
      expect(res.body.message).toBe("Test notification");
      expect(res.body.channel).toBe("email");
      expect(res.body.status).toBe("sent");
      expect(res.body.id).toMatch(/^notif-/);
    });

    it("should use default channel when not specified", async () => {
      const res = await request(app)
        .post("/notifications/send")
        .send({ type: "test", message: "No channel" });
      expect(res.status).toBe(201);
      expect(res.body.channel).toBe("default");
    });

    it("should return 400 when type is missing", async () => {
      const res = await request(app)
        .post("/notifications/send")
        .send({ message: "No type" });
      expect(res.status).toBe(400);
      expect(res.body.error).toBeDefined();
    });

    it("should return 400 when message is missing", async () => {
      const res = await request(app)
        .post("/notifications/send")
        .send({ type: "test" });
      expect(res.status).toBe(400);
      expect(res.body.error).toBeDefined();
    });
  });

  describe("GET /notifications", () => {
    it("should list notifications", async () => {
      const res = await request(app).get("/notifications");
      expect(res.status).toBe(200);
      expect(res.body.notifications).toBeDefined();
      expect(Array.isArray(res.body.notifications)).toBe(true);
      expect(res.body.total).toBeGreaterThanOrEqual(0);
    });
  });

  describe("GET /notifications/channels", () => {
    it("should list channels", async () => {
      const res = await request(app).get("/notifications/channels");
      expect(res.status).toBe(200);
      expect(res.body.channels).toBeDefined();
      expect(Array.isArray(res.body.channels)).toBe(true);
    });
  });
});
