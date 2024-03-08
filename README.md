rate-limiter-go

Client will be sending requests to our end point server, but it will have to first go through the middleware rate limiter layer which will allow only limited number of requests to the end server.

There are two options to hit the enpoint server, either we can send the requests sequentially or we can send concurrently.

runSequentially() -> hits the rate limiter sequentially
runConcurrently() -> hits the rate limiter concurrently
How to Run?

Step 1 [Start Server] Server will start the endpoint server with rate limiter as middleware go to server directory -> go build -> ./server

Step 2 [Start CLient] Client will start the endpoint server with rate limiter as middleware go to client directory -> go build -> ./client
