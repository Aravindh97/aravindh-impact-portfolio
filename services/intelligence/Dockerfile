# syntax=docker/dockerfile:1
FROM python:3.12.15-slim
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
WORKDIR /app
COPY services/__init__.py services/__init__.py
COPY services/intelligence/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt
COPY services/intelligence/ services/intelligence/
RUN useradd --system --uid 10001 appuser
USER appuser
# 0.0.0.0 is internal to the private container network; no host port is published.
EXPOSE 8091
CMD ["python", "-m", "uvicorn", "services.intelligence.app.main:app", "--host", "0.0.0.0", "--port", "8091", "--no-access-log"]
