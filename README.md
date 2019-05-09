# Test-scores
Welcome to the test-scores API. This is an http server that when running will connect to a SSE event stream from http://live-test-scores.herokuapp.com/scores. The data that is gathered from the stream is aggregated to two places, ExamData and StudentData. These are two structs that contain a map of the aggregated data that pertains to a specific student or to a specific exam. These maps are created when the program is run, and stored on local memory so all the information is lost when one closes the server.

This is also an API server that will return some json data back on the following routes

```
http://localhost:8080/students
http://localhost:8080/students/{studentID}
http://localhost:8080/exams
http://localhost:8080/exams/{examID}
```

# Install

To install this first go this repo on to your local machine

``` go get github.com/bradenaa/test-scores ```

Now grab the gorilla mux dependency (if you don't have it already). This is a router to better handle some of the dynamic routes that take a request parameter

``` go get github.com/gorilla/mux ```

Lastly grab the SSE depency, which will create a new client connect to the SSE server at the heroku link above.

``` go get github.com/r3labs/sse ```

# Run

Now we can run this by navigating to the project

``` cd $GOPATH/src/github.com/bradenaa/test-scores ```

Build the binary

``` go build ```

Run the program

``` ./test-scores ```

No you can vist any of the endpoints listed above to find the student or exam information