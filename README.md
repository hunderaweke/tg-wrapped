# 🚀 TG-Wrapped

TG-Wrapped is a Go-based application that generates and displays analytics for a given Telegram channel. It provides insights into channel activity, such as message frequency and streaks, and serves this information through a web interface.

## 🛠️ Tech Stack

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white) ![Gin](https://img.shields.io/badge/Gin-0081cb?style=for-the-badge&logo=go&logoColor=white) ![Telegram](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white) ![Minio](https://img.shields.io/badge/Minio-C72E49?style=for-the-badge&logo=minio&logoColor=white) ![Redis](https://img.shields.io/badge/redis-%23DD0031.svg?style=for-the-badge&logo=redis&logoColor=white)

- **Go**: The primary programming language.
- **Gin**: A web framework for building the API.
- **gotd/td**: A library for interacting with the Telegram API.
- **Minio**: An object storage server used for storing channel profile pictures.
- **Redis**: An in-memory data store, likely used for caching or session management.
- **godotenv**: A library for managing environment variables.

## ⚙️ Setup and Run

### Prerequisites

- **Go** (version 1.24 or higher)
- **Redis** (running on `localhost:6379`)
- **Minio** (running and accessible)

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/hunderaweke/tg-wrapped.git
    cd tg-wrapped
    ```

2.  **Set up environment variables:**

    Create a `.env` file in the root directory with the following variables:

    ```env
    APP_ID=your_telegram_app_id
    APP_HASH=your_telegram_app_hash
    APP_SESSION_STORAGE=session.json
    SERVER_PORT=8080

    MINIO_ENDPOINT=localhost:9000
    MINIO_ACCESS_ID=your_minio_access_key
    MINIO_SECRET_ID=your_minio_secret_key
    MINIO_BUCKET=tg-wrapped-profiles
    # MINIO_TOKEN=optional_token
    ```

3.  **Run the application:**

    ```bash
    go run .
    ```

## 📡 API Reference

### 1. Health Check

- **Endpoint**: `GET /health`
- **Description**: Checks if the server is running.
- **Response**: `200 OK`

### 2. Generate Analytics

- **Endpoint**: `POST /analytics`
- **Description**: Generates analytics for a specific Telegram channel.
- **Request Body**:

  ```json
  {
    "username": "channel_username"
  }
  ```

- **Response**:

  ```json
  {
    "channel_profile": "profile.jpg",
    "channel_name": "Channel Title",
    "totals": {
      "total_views": 1000,
      "total_comments": 50,
      "total_reactions": 200,
      "total_posts": 10,
      "total_forwards": 5
    },
    "trends": {
      "views_by_month": {
        "2025-January": 500
      },
      "posts_by_day": {
        "2025-January": [1, 0, 2, ...]
      },
      "posts_by_month": {
        "2025-January": 10
      },
      "posts_by_hour": {
        "14": 5
      },
      "longest_posting_streak": 3
    },
    "highlights": {
      "most_viewed_id": 123,
      "most_viewed_count": 500,
      "most_commented_id": 124,
      "most_commented_count": 20,
      "forwards_by_post": {
        "123": 2
      },
      "reactions_by_type": {
        "❤️": 10,
        "👍": 5
      }
    }
  }
  ```

### 3. Get Profile Picture

- **Endpoint**: `GET /profiles/:objectName`
- **Description**: Redirects to a pre-signed URL for the channel's profile picture.
- **Parameters**:
  - `objectName`: The filename of the profile picture (returned in the analytics response).

## 🔍 How It Works

1.  **🔐 Authentication**: The application authenticates with the Telegram API using credentials provided through environment variables.
2.  **📊 Channel Analysis**: When a user requests analytics for a specific Telegram channel, the application:
    - Fetches the channel's metadata.
    - Downloads the channel's profile picture and stores it in a Minio bucket.
    - Retrieves the message history of the channel.
3.  **📈 Analytics Generation**: The message history is processed to generate various analytics, such as the longest messaging streak.
4.  **🌐 API**: The generated analytics are exposed through a RESTful API built with Gin.

## 🚧 Current Limitations

- **⏳ Synchronous Processing**: Analytics generation is a long-running task that currently blocks incoming requests. This can lead to timeouts for channels with a large number of messages.
- **⌨️ Interactive Authentication**: The current authentication method requires interactive input from the terminal, which is not ideal for a service that is intended to run in the background.
- **📅 Hardcoded Analysis Start Date**: The starting date for message analysis is currently hardcoded, limiting the flexibility of the analytics.

## 🔮 Future Plans

- **⚡ Asynchronous Request Processing**: To address the limitations of synchronous processing, we plan to implement a message queue (e.g., RabbitMQ or NATS). This will allow for the asynchronous processing of analytics requests, improving the responsiveness and reliability of the service.
- **🤖 Non-Interactive Authentication**: We will explore more suitable authentication methods, such as using a bot token or implementing a more robust session management system.
- **🗓️ Configurable Analysis Period**: The start date for the analysis will be made configurable, allowing users to specify the time range for which they want to generate analytics.
- **📊 Expanded Analytics**: We plan to add more types of analytics to provide more comprehensive insights into channel activity.
- **🖥️ Frontend Interface**: A user-friendly frontend will be developed to visualize the analytics and provide a more engaging user experience.
