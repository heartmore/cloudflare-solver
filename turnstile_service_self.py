"""
Turnstile验证服务类 - 纯自建版
自建 API 同时获取 cf_clearance + turnstile_token
"""
import os, time, requests
from dotenv import load_dotenv
load_dotenv()

class TurnstileService:
    def __init__(self, solver_url="http://127.0.0.1:5072"):
        self.api_url = os.getenv('SOLVER_API_URL', '').strip()
        self.api_key = os.getenv('SOLVER_API_KEY', '').strip()

    def _headers(self):
        return {"X-API-Key": self.api_key} if self.api_key else {}

    def create_task(self, siteurl, sitekey):
        if not self.api_url:
            raise Exception("SOLVER_API_URL not configured")
        r = requests.post(f"{self.api_url}/v1/solve", json={
            "url": siteurl,
            "sitekey": sitekey,
            "timeout": 60000,
        }, headers=self._headers(), timeout=30)
        r.raise_for_status()
        return {"task_id": r.json()["task_id"]}

    def get_response(self, task, max_retries=20, initial_delay=2, retry_delay=1.5):
        if not isinstance(task, dict): return None
        tid = task.get("task_id")
        if not tid: return None
        time.sleep(initial_delay)
        result = {"token": None, "cookies": {}, "userAgent": ""}
        for _ in range(max_retries):
            try:
                r = requests.get(f"{self.api_url}/v1/result/{tid}", headers=self._headers(), timeout=10)
                d = r.json()
                if d["status"] == "completed":
                    sol = d.get("solution", {})
                    result["token"] = sol.get("turnstile_token")
                    result["cookies"] = {c["name"]: c["value"] for c in sol.get("cookies", [])}
                    result["userAgent"] = sol.get("userAgent", "")
                    return result if result["token"] else None
                elif d["status"] == "failed": return None
                time.sleep(retry_delay)
            except: time.sleep(retry_delay)
        return None
