package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/r3labs/sse"
)

// TestEventData data structure for extracting the data from SSE stream
type TestEventData struct {
	StudentID string  `json:"studentId"`
	ExamID    int     `json:"exam"`
	Score     float64 `json:"score"`
}

// StudentData structure for all of the students found in stream
// using map, for O(1) lookup
type StudentData struct {
	v   map[string]Student
	mux sync.Mutex
}

// Student struct for reference to a students previous exam data
type Student struct {
	Exams   []Test
	Average float64
}

// Test struct for information about a specific test taking
type Test struct {
	ExamID int
	Score  float64
}

// GetStudent will return a student struct, on its key name
func (m *StudentData) GetStudent(key string) (Student, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	student, ok := m.v[key]
	return student, ok
}

// SetStudent will set a new student into StudentData
func (m *StudentData) SetStudent(key string, val Student) {
	m.mux.Lock()
	m.v[key] = val
	m.mux.Unlock()
	return
}

// AddStudentTest will add a test to a students list of scores
func (m *StudentData) AddStudentTest(key string, test Test) {
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
func (m *StudentData) AverageAllStudentsExams() {
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

// GetAllStudents will create json out of the StudentData map
func (m *StudentData) GetAllStudents(w http.ResponseWriter, req *http.Request) {
	// log.Printf("inc %v", req)
	m.mux.Lock()
	jsonString, err := json.Marshal(m.v)
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Write(jsonString)
	m.mux.Unlock()
	return
}

// GetStudentTestsAndAverage will look up student by id
func (m *StudentData) GetStudentTestsAndAverage(w http.ResponseWriter, req *http.Request) {
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
func (m *ExamData) GetExam(key string) (Exam, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	exam, ok := m.v[key]
	return exam, ok
}

// SetExam is a method to set a new Exam in the ExamData map
func (m *ExamData) SetExam(key string, val Exam) {
	m.mux.Lock()
	m.v[key] = val
	m.mux.Unlock()
	return
}

// AddExamScore is a method to add a new score to a given exam
func (m *ExamData) AddExamScore(key string, score float64) {
	m.mux.Lock()
	exam := m.v[key]
	newScoreList := append(exam.Scores, score)
	exam.Scores = newScoreList
	m.v[key] = exam
	m.mux.Unlock()
	return
}

// GetAllExams will arrange all exams in a slice, then respond slice as json
func (m *ExamData) GetAllExams(w http.ResponseWriter, req *http.Request) {
	m.mux.Lock()
	var exams = make([]string, 1)
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

// GetExamResults will look up scores of an examID then return the slice of
// test scores as json
func (m *ExamData) GetExamResults(w http.ResponseWriter, req *http.Request) {
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

/*
handleEvents will write all the SSE data into the StudentData and the
ExamData maps. Since the SSE data comes in batches of 20, we will
re-average the student test scores after every twenty items
*/
func handleEvents(events chan *sse.Event, s StudentData, e ExamData) {
	var count int = 0
	var examStr string
	// convert event messages to Test struct
	for {
		msg := <-events
		var t TestEventData
		if err := json.Unmarshal(msg.Data, &t); err != nil {
			fmt.Println("error:", err)
		}

		examStr = strconv.Itoa(t.ExamID)
		// handle the exam
		_, ok1 := e.GetExam(examStr)
		if ok1 {
			e.AddExamScore(examStr, t.Score)
		} else {
			e.SetExam(examStr, Exam{[]float64{t.Score}})
		}
		fmt.Println(t)
		// handle the student
		_, ok2 := s.GetStudent(t.StudentID)
		if ok2 {
			s.AddStudentTest(t.StudentID, Test{t.ExamID, t.Score})
		} else {
			s.SetStudent(t.StudentID, Student{[]Test{Test{t.ExamID, t.Score}}, t.Score})
		}

		// increment count
		count++
		// handle last stream item, by find all the new averages
		if count >= 20 {
			count = 0
			s.AverageAllStudentsExams()
		}
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
}

func main() {
	// create Student data and Exam data for lookups
	s := StudentData{v: make(map[string]Student)}
	e := ExamData{v: make(map[string]Exam)}

	// handling event stream in a new channel
	events := make(chan *sse.Event)
	client := sse.NewClient("http://live-test-scores.herokuapp.com/scores")
	client.SubscribeChan("messages", events)
	go handleEvents(events, s, e)

	// Creating routes
	r := mux.NewRouter()
	r.HandleFunc("/students", s.GetAllStudents)
	r.HandleFunc("/students/{studentID}", s.GetStudentTestsAndAverage)
	r.HandleFunc("/exams", e.GetAllExams)
	r.HandleFunc("/exams/{examID}", e.GetExamResults)

	// establishing a server
	http.ListenAndServe(":8080", r)
}
