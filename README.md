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

# Run

Now we can run this by navigating to the project

``` cd $GOPATH/src/github.com/bradenaa/test-scores ```

Build the binary

``` go build ```

Run the program

``` ./test-scores ```

No you can vist any of the endpoints listed above to find the student or exam information

# Test

Unit tests are setup for the student data map methods, the exam data map methods, and the route handlers. Run the following if you make changes and want to check if these are still working.

```go test```

or for more details

```go test -v```