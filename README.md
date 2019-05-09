# Test-scores
Welcome to the test-scores API. This is an http server that when running will connect to a SSE event stream from http://live-test-scores.herokuapp.com/scores. The data that is gathered from the stream is aggregated in two places: ExamData and StudentData. These are two structs that contain a map of the aggregated data that pertains to a specific student or to a specific exam. These maps are created when the program is run, and stored on local memory so all the information is lost when one closes the server.

This is also an API that will return some json data back from StudentData and ExamData on the following endpoints

```
http://localhost:8080/students
http://localhost:8080/students/{studentID}
http://localhost:8080/exams
http://localhost:8080/exams/{examID}
```

# Install

If you do not have GO installed, please follow the the official docs:

https://golang.org/doc/install

To install this repo on to your local machine:

``` go get github.com/bradenaa/test-scores ```

# Run

Now we can run this by navigating to the project

``` cd $GOPATH/src/github.com/bradenaa/test-scores ```

Building the binary

``` go build ```

Runing the executable

``` ./test-scores ```

Now you can make a get request to any of the endpoints listed above to find the student or exam information. This data is being updated as time progresses.

# Test

Unit tests are setup for the StudentData map methods, the ExamData map methods, and the route handlers. Run the following if you make changes and want to check if these are still working.

```go test```

or for more details

```go test -v```