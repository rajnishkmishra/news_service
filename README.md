News Service

This is a RESTful web service that stores and retrieves news data.

Features
- Retrieve articles from a specic category (e.g., "Technology", "Business", "Sports").
- Retrieve articles based on a relevance score.
- Retrieve articles based on a search query.
- Retrieve articles from specic sources (e.g., "New York Times", "Reuters").
- Retrieve articles published within a specied radius (e.g., 10km) of a given location (latitude and longitude).
- Retrieve trending news near me.

Tech Stack
Language: Golang
Framework: Gin (for HTTP handling)
Database: MySQL (Dockerized)
Search Engine: Elastic Search (Dockerized)
ORM: GORM
Containerization: Docker, Docker Compose

API Endpoints
1. Retrieve articles from a specic category
SAMPLE CURL:
<pre> curl --location 'localhost:8080/api/v1/news/catorgory/sports?p=1&l=5' </pre>

2. Retrieve articles based on a relevance score
SAMPLE CURL:
<pre> curl --location 'localhost:8080/api/v1/news/score/0.4?p=1&l=10' </pre>

3. Retrieve articles based on a search query
SAMPLE CURL:
<pre> curl --location 'localhost:8080/api/v1/news/search?q=news%20from%20News18&p=1&l=2' </pre>

4. Retrieve articles from specic sources
SAMPLE CURL:
<pre> curl --location 'localhost:8080/api/v1/news/source/ANI News?p=2&l=20' </pre>

5. Retrieve articles published within a specied radius
SAMPLE CURL:
<pre> curl --location 'localhost:8080/api/v1/news/nearby?lat=17.900636&long=77.465262&radius=10&p=1&l=5' </pre>

6. Retrieve trending news near me
SAMPLE CURL:
<pre> curl --location 'localhost:8080/api/v1/trending?lat=18.069141&long=76.621249' </pre>

Setup and Run
Prerequisites:
- Go 1.25+
- Docker

Installation:
- git clone https://github.com/rajnishkmishra/news_service.git (clone the repository in path: "~/go/src")
- cd news_service
- go mod tidy

Running the Application:
- For running the application run below command from your terminal
<pre> docker-compose up --build </pre>
- This will spin up MySQL database and Elastic search in docker container.
- Run this go application <pre> go run main.go </pre>
- Now you can hit APIs from postman (Import the provided postman collection: NewsService.postman_collection).
