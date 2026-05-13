import pytest
from app import app


@pytest.fixture
def client():
    app.config["TESTING"] = True
    with app.test_client() as client:
        yield client


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "healthy"
    assert data["service"] == "api-gateway"
    assert "timestamp" in data


def test_create_workflow_success(client):
    resp = client.post("/api/workflows", json={"name": "Test Workflow", "description": "A test"})
    assert resp.status_code == 201
    data = resp.get_json()
    assert data["name"] == "Test Workflow"
    assert data["status"] == "created"
    assert data["id"].startswith("wf-")


def test_create_workflow_missing_name(client):
    resp = client.post("/api/workflows", json={"description": "No name"})
    assert resp.status_code == 400
    data = resp.get_json()
    assert "error" in data


def test_create_workflow_no_body(client):
    resp = client.post("/api/workflows", content_type="application/json")
    assert resp.status_code == 400


def test_list_workflows(client):
    resp = client.get("/api/workflows")
    assert resp.status_code == 200
    data = resp.get_json()
    assert "workflows" in data
    assert "total" in data


def test_get_workflow_not_found(client):
    resp = client.get("/api/workflows/wf-nonexistent")
    assert resp.status_code == 404


def test_service_status(client):
    resp = client.get("/api/status")
    assert resp.status_code == 200
    data = resp.get_json()
    assert "api-gateway" in data
    assert data["api-gateway"] == "healthy"
