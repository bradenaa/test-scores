package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

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
	Exams []Test
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

// ExamData struct for all the exams found in stream
// using map, for O(1) lookup
type ExamData struct {
	v   map[int]Exam
	mux sync.Mutex
}

// Exam struct for holding information regarding a specific exam
type Exam struct {
	Scores []float64
}

// GetExam is a method to return an Exam stuct from ExamData map
func (m *ExamData) GetExam(key int) (Exam, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	exam, ok := m.v[key]
	return exam, ok
}

// SetExam is a method to set a new Exam in the ExamData map
func (m *ExamData) SetExam(key int, val Exam) {
	m.mux.Lock()
	m.v[key] = val
	m.mux.Unlock()
	return
}

// AddExamScore is a method to add a new score to a given exam
func (m *ExamData) AddExamScore(key int, score float64) {
	m.mux.Lock()
	exam := m.v[key]
	newScoreList := append(exam.Scores, score)
	exam.Scores = newScoreList
	m.v[key] = exam
	m.mux.Unlock()
	return
}

func handleEvents(events chan *sse.Event, s StudentData, e ExamData) {
	// convert event messages to Test struct
	for {
		msg := <-events
		var t TestEventData
		if err := json.Unmarshal(msg.Data, &t); err != nil {
			fmt.Println("error:", err)
		}

		// handle the exam
		_, ok1 := e.GetExam(t.ExamID)
		if ok1 {
			e.AddExamScore(t.ExamID, t.Score)
		} else {
			e.SetExam(t.ExamID, Exam{[]float64{t.Score}})
		}

		// handle the student
		student, ok2 := s.GetStudent(t.StudentID)
		fmt.Println(t.StudentID, student)
		if ok2 {
			s.AddStudentTest(t.StudentID, Test{t.ExamID, t.Score})
		} else {
			s.SetStudent(t.StudentID, Student{[]Test{Test{t.ExamID, t.Score}}})
		}
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	w.Write([]byte("hello!"))
}

func main() {
	// create Student data and Exam data for lookups
	s := StudentData{v: make(map[string]Student)}
	e := ExamData{v: make(map[int]Exam)}
	// handling event stream in a new channel
	events := make(chan *sse.Event)
	client := sse.NewClient("http://live-test-scores.herokuapp.com/scores")
	client.SubscribeChan("messages", events)
	go handleEvents(events, s, e)

	// establishing a server
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8080", nil)
}
