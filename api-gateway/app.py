import os
import logging
import time
from flask import Flask, jsonify, request
from flask_cors import CORS
import requests

app = Flask(__name__)
CORS(app)

logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("api-gateway")

ANALYTICS_URL = os.getenv("ANALYTICS_URL", "http://localhost:8081")
NOTIFICATION_URL = os.getenv("NOTIFICATION_URL", "http://localhost:8082")


@app.route("/health")
def health():
    return jsonify({"status": "healthy", "service": "api-gateway", "timestamp": time.time()})


@app.route("/api/workflows", methods=["POST"])
def create_workflow():
    data = request.get_json()
    if not data or "name" not in data:
        logger.warning("Invalid workflow creation request: missing name")
        return jsonify({"error": "Workflow name is required"}), 400

    workflow = {
        "id": f"wf-{int(time.time())}",
        "name": data["name"],
        "description": data.get("description", ""),
        "status": "created",
        "created_at": time.time(),
    }
    logger.info("Workflow created: %s", workflow["id"])

    try:
        requests.post(
            f"{ANALYTICS_URL}/analytics/event",
            json={"event": "workflow_created", "workflow_id": workflow["id"]},
            timeout=5,
        )
    except requests.RequestException as e:
        logger.error("Failed to send analytics event: %s", e)

    try:
        requests.post(
            f"{NOTIFICATION_URL}/notifications/send",
            json={
                "type": "workflow_created",
                "message": f"Workflow '{workflow['name']}' has been created",
                "channel": data.get("notify_channel", "default"),
            },
            timeout=5,
        )
    except requests.RequestException as e:
        logger.error("Failed to send notification: %s", e)

    return jsonify(workflow), 201


@app.route("/api/workflows", methods=["GET"])
def list_workflows():
    logger.info("Listing workflows")
    return jsonify({"workflows": [], "total": 0})


@app.route("/api/workflows/<workflow_id>", methods=["GET"])
def get_workflow(workflow_id):
    logger.info("Getting workflow: %s", workflow_id)
    return jsonify({"error": f"Workflow {workflow_id} not found"}), 404


@app.route("/api/status", methods=["GET"])
def service_status():
    statuses = {"api-gateway": "healthy"}
    for name, url in [("analytics-engine", ANALYTICS_URL), ("notification-service", NOTIFICATION_URL)]:
        try:
            resp = requests.get(f"{url}/health", timeout=3)
            statuses[name] = "healthy" if resp.status_code == 200 else "unhealthy"
        except requests.RequestException:
            statuses[name] = "unreachable"
    return jsonify(statuses)


if __name__ == "__main__":
    port = int(os.getenv("API_GATEWAY_PORT", "8080"))
    logger.info("Starting API Gateway on port %d", port)
    app.run(host="0.0.0.0", port=port, debug=os.getenv("FLASK_DEBUG", "false").lower() == "true")
