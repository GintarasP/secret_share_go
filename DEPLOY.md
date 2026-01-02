# Deployment Guide

This application is containerized and ready to be deployed to any platform including free tiers like Fly.io or Render.

## Option 1: Fly.io (Recommended)

Fly.io offers a free allowance that fits this application perfectly.

1.  **Install Fly CLI:**
    ```bash
    curl -L https://fly.io/install.sh | sh
    ```

2.  **Login:**
    ```bash
    fly auth login
    ```

3.  **Launch:**
    Run this in the project root. It will use the provided `Dockerfile` and `fly.toml`.
    ```bash
    fly launch
    ```
    *   Follow the prompts. You can name your app (e.g., `my-secret-share`).
    *   Say "Yes" to copying configuration.

4.  **Deploy:**
    ```bash
    fly deploy
    ```

Your app will be live at `https://<your-app-name>.fly.dev`.

## Option 2: Render.com

Render has a free "Web Service" tier.

1.  Push this code to a GitHub repository.
2.  Log in to [Render.com](https://render.com).
3.  Click "New +", select "Web Service".
4.  Connect your repository.
5.  Select "Docker" as the Runtime.
6.  Click "Create Web Service".

## Option 3: Docker Local

You can run it anywhere with Docker:

1.  **Build:**
    ```bash
    docker build -t secret-share .
    ```

2.  **Run:**
    ```bash
    docker run -p 8080:8080 secret-share
    ```
