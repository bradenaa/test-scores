package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/r3labs/sse"
)

func main() {
	// create Student data and Exam data for lookups
	s := &StudentData{v: make(map[string]Student)}
	e := &ExamData{v: make(map[string]Exam)}

	// handling event stream in a new channel
	events := make(chan *sse.Event)
	client := sse.NewClient("http://live-test-scores.herokuapp.com/scores")
	client.SubscribeChan("messages", events)
	go handleEvents(events, s, e)

	// Creating routes
	r := mux.NewRouter()

	// Student Routes
	r.HandleFunc("/students", GetAllStudents(s))
	r.HandleFunc("/students/{studentID}", GetStudentTestsAndAverage(s))

	// Exam Routes
	r.HandleFunc("/exams", GetAllExams(e))
	r.HandleFunc("/exams/{examID}", GetExamResults(e))

	// establishing a server
	log.Fatal(http.ListenAndServe(":8080", r))
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// =================== SSE HANDLER ========================
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// TestEventData data structure for extracting the data from SSE stream
type TestEventData struct {
	StudentID string  `json:"studentId"`
	ExamID    int     `json:"exam"`
	Score     float64 `json:"score"`
}

/*
handleEvents will write all the SSE data into the StudentData and the
ExamData maps. Since the SSE data comes in batches of 20, we will
re-average the student test scores after every twenty items
*/
func handleEvents(events chan *sse.Event, s *StudentData, e *ExamData) {
	var count int
	var examStr string
	// convert event messages to Test struct
	for {
		msg := <-events
		var t TestEventData
		if err := json.Unmarshal(msg.Data, &t); err != nil {
			fmt.Println("error:", err)
		}

		examStr = strconv.Itoa(t.ExamID)
		// Try to get the current exam by its ID. If it is found, then add the score to
		// the that exam. Otherwise, create a new exam to the ExamData in memory
		_, examFound := GetExam(e, examStr)
		if examFound {
			AddExamScore(e, examStr, t.Score)
		} else {
			SetExam(e, examStr, Exam{[]float64{t.Score}})
		}

		// Try to find the current student. If we find the student, then add
		// a new test to that student in the Student Data in memory. Otherwise,
		// Create a new Student and set that student in the Student Data in memory.
		_, studentFound := GetStudent(s, t.StudentID)
		if studentFound {
			AddStudentTest(s, t.StudentID, Test{t.ExamID, t.Score})
		} else {
			SetStudent(s, t.StudentID, Student{[]Test{Test{t.ExamID, t.Score}}, t.Score})
		}

		// increment count
		count++
		// When the count is 20, we have reached the end of our stream batch, so
		// time to do some cleaning up in memory. Re-average all of the exams scores,
		// and also parse the Student Data map into []unit8 and store in memory, so we won't have
		// to perform this operation with every request to "/students". The event stream,
		// sends a batch about every 5 seconds.
		if count >= 20 {
			count = 0
			AverageAllStudentsExams(s)
			UpdateStudentDataJSON(s)
		}
	}
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// =================== STUDENTS ===========================
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// StudentData structure for all of the students found in stream
// using map, for O(1) lookup
type StudentData struct {
	v    map[string]Student
	json []uint8
	mux  sync.Mutex
}

// Student struct for reference to a students previous exam data
type Student struct {
	Exams   []Test
	Average float64
}

// Test struct for information about a specific test taken
type Test struct {
	ExamID int
	Score  float64
}

// GetStudent will return a student struct, on its key name
func GetStudent(m *StudentData, key string) (Student, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	student, ok := m.v[key]
	return student, ok
}

// SetStudent will set a new student into StudentData
func SetStudent(m *StudentData, key string, val Student) {
	m.mux.Lock()
	m.v[key] = val
	m.mux.Unlock()
	return
}

// AddStudentTest will add a test to a students list of scores
func AddStudentTest(m *StudentData, key string, test Test) {
	m.mux.Lock()
	student := m.v[key]
	newExams := append(student.Exams, test)
	student.Exams = newExams
	m.v[key] = student
	m.mux.Unlock()
	return
}

/*
AverageAllStudentsExams will be called after the 20 event data points
have accumulated. This will happen so that the averages get up to date
after every SSE stream batch
*/
func AverageAllStudentsExams(m *StudentData) {
	m.mux.Lock()
	var average float64
	var sum float64
	for studentID, student := range m.v {
		sum = 0
		for _, test := range student.Exams {
			sum += test.Score
		}
		average = (sum / float64(len(m.v[studentID].Exams)))
		student.Average = average
		m.v[studentID] = student
	}
	m.mux.Unlock()
	return
}

// UpdateStudentDataJSON will update the uint8 information so that we won't have
// to json.Marshal on every GetAllStudents Requests
func UpdateStudentDataJSON(m *StudentData) {
	m.mux.Lock()
	jsonString, err := json.Marshal(m.v)
	if err != nil {
		fmt.Println("error:", err)
	}
	m.json = jsonString
	m.mux.Unlock()
	return
}

// GetAllStudents will send the json []uint8 chars from student map
func GetAllStudents(m *StudentData) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		m.mux.Lock()
		w.Write(m.json)
		m.mux.Unlock()
		return
	}
}

// GetStudentTestsAndAverage will look up student by id
func GetStudentTestsAndAverage(m *StudentData) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		m.mux.Lock()
		vars := mux.Vars(req)

		jsonString, err := json.Marshal(m.v[vars["studentID"]])
		if err != nil {
			fmt.Println("error:", err)
		}
		w.Write(jsonString)
		m.mux.Unlock()
		return
	}
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// ====================  EXAMS  ===========================
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// ExamData struct for all the exams found in stream
// using map, for O(1) lookup
type ExamData struct {
	v   map[string]Exam
	mux sync.Mutex
}

// Exam struct for holding information regarding a specific exam
type Exam struct {
	Scores []float64
}

// GetExam is a method to return an Exam stuct from ExamData map
func GetExam(m *ExamData, key string) (Exam, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	exam, ok := m.v[key]
	return exam, ok
}

// SetExam is a method to set a new Exam in the ExamData map
func SetExam(m *ExamData, key string, val Exam) {
	m.mux.Lock()
	m.v[key] = val
	m.mux.Unlock()
	return
}

// AddExamScore is a method to add a new score to a given exam
func AddExamScore(m *ExamData, key string, score float64) {
	m.mux.Lock()
	exam := m.v[key]
	newScoreList := append(exam.Scores, score)
	exam.Scores = newScoreList
	m.v[key] = exam
	m.mux.Unlock()
	return
}

// GetAllExams will arrange all exams in a slice, then respond slice as json
func GetAllExams(m *ExamData) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		m.mux.Lock()
		var exams = make([]string, 0)
		for e := range m.v {
			exams = append(exams, e)
		}
		jsonString, err := json.Marshal(exams)
		if err != nil {
			fmt.Println("error", err)
		}
		w.Write(jsonString)
		m.mux.Unlock()
		return
	}
}

// GetExamResults will look up scores of an examID then return the slice of
// test scores as json
func GetExamResults(m *ExamData) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		m.mux.Lock()
		vars := mux.Vars(req)

		jsonString, err := json.Marshal(m.v[vars["examID"]].Scores)
		if err != nil {
			fmt.Println("error", err)
		}
		w.Write(jsonString)
		m.mux.Unlock()
		return
	}
}
