News Service

This is a RESTful web service that stores and retrieves news data.

Features
- Retrieve articles from a specic category (e.g., "Technology", "Business", "Sports").
- Retrieve articles based on a relevance score.
- Retrieve articles based on a search query.
- Retrieve articles from specic sources (e.g., "New York Times", "Reuters").
- Retrieve articles published within a specied radius (e.g., 10km) of a given location (latitude and longitude).

Tech Stack
Language: Golang
Framework: Gin (for HTTP handling)
Database: MySQL (running on Docker), ElasticSearch
ORM: GORM

API Endpoints
1. Create a Transaction
POST /transactionservice/transaction
SAMPLE CURL:
<pre> curl --location --request POST 'localhost:8080/transactionservice/transaction' </pre>

2. Modify a Transaction
PUT /transactionservice/transaction/{transaction_id}
SAMPLE CURL:
<pre> curl --location --request PUT 'localhost:8080/transactionservice/transaction/2' \
--header 'Content-Type: application/json' \
--data '{
    "amount": 5000,
    "type": "shopping",
    "parent_id": 1
}' </pre>

3. Retrieve a Transaction
GET /transactionservice/transaction/{transaction_id}
SAMPLE CURL:
<pre> curl --location 'localhost:8080/transactionservice/transaction/2' </pre>

4. Get Transactions by Type
GET /transactionservice/types/{type}
SAMPLE CURL:
<pre> curl --location 'localhost:8080/transactionservice/types/shopping' </pre>

5. Compute Sum of Transactions Linked to a Given Transaction
GET /transactionservice/sum/{transaction_id}
SAMPLE CURL:
<pre> curl --location 'localhost:8080/transactionservice/sum/1' </pre>

Setup and Run
Prerequisites:
- Go 1.23
- Docker

Installation:
- git clone https://github.com/rajnishkmishra/transaction_service.git (clone the repository in path: "~/go/src/bitbucket.org")
- cd transaction_service
- go mod tidy

Running the Application:
- For running the application run below command from your terminal
<pre> docker-compose up --build </pre>
- This will spin up the API along with the MySQL database.
- Now you can hit APIs from postman (Import the provided postman collection: TrasactionService.postman_collection.json).
