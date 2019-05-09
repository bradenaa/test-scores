package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// see the student data
var s = &StudentData{v: make(map[string]Student)}
var e = &ExamData{v: make(map[string]Exam)}

func SeedStudent(s *StudentData) {
	var test Student
	var data = []byte(`{"Exams": [{"ExamID": 10775,"Score": 0.7789680185088161}],"Average": 0.7789680185088161}`)
	json.Unmarshal(data, &test)
	s.v["Angelina.Jones"] = test
	jsonString, _ := json.Marshal(s.v)
	s.json = jsonString
	return
}

func SeedExam(e *ExamData) {
	var exam = Exam{[]float64{0.7789680185088161}}
	e.v["10775"] = exam
	return
}

func init() {
	SeedStudent(s)
	SeedExam(e)
}

func TestGetAllStudents(t *testing.T) {
	req, err := http.NewRequest("GET", "/students", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetAllStudents(s))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned the wrong status code: got %v expected %v", status, http.StatusOK)
	}

	expected := `{"Angelina.Jones":{"Exams":[{"ExamID":10775,"Score":0.7789680185088161}],"Average":0.7789680185088161}}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetStudentTestsAndAverage(t *testing.T) {
	req, err := http.NewRequest("GET", "/students/Angelina.Jones", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/students/{studentID}", GetStudentTestsAndAverage(s))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned the wrong status code: got %v expected %v", status, http.StatusOK)
	}

	expected := `{"Exams":[{"ExamID":10775,"Score":0.7789680185088161}],"Average":0.7789680185088161}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetAllExams(t *testing.T) {
	req, err := http.NewRequest("GET", "/exams", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetAllExams(e))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned the wrong status code: got %v expected %v", status, http.StatusOK)
	}

	expected := `["10775"]`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetExamResults(t *testing.T) {
	req, err := http.NewRequest("GET", "/exams/10775", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/exams/{examID}", GetExamResults(e))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned the wrong status code: got %v expected %v", status, http.StatusOK)
	}

	expected := `[0.7789680185088161]`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetStudent(t *testing.T) {
	student, ok := GetStudent(s, "Angelina.Jones")

	if student.Average != 0.7789680185088161 {
		t.Errorf("map lookup returned wrong result: got %v want %v", student.Average, 0.7789680185088161)
	}

	if ok != true {
		t.Errorf("map lookup returned result: got %v want %v", false, ok)
	}
}

func TestSetStudent(t *testing.T) {
	test := Test{12345, 0.75}
	student := Student{[]Test{test}, 0.75}
	SetStudent(s, "Braden.Altstatt", student)

	info, ok := s.v["Braden.Altstatt"]

	if info.Average != 0.75 {
		t.Errorf("map lookup returned wrong result: got %v want %v", info.Average, 0.75)
	}

	if ok != true {
		t.Errorf("map lookup returned result: got %v want %v", false, ok)
	}

	delete(s.v, "Braden.Altstatt")
}

func TestAddStudentTest(t *testing.T) {
	test1 := Test{12345, 0.75}
	student := Student{[]Test{test1}, 0.75}
	s.v["Braden.Altstatt"] = student
	test2 := Test{98765, 0.25}
	AddStudentTest(s, "Braden.Altstatt", test2)

	info, ok := s.v["Braden.Altstatt"]

	found := false
	for _, exam := range info.Exams {
		if exam.ExamID == 98765 {
			found = true
		}
	}

	if found != true {
		t.Errorf("map lookup expected to find a new exam: got %v want %v", found, true)
	}

	if ok != true {
		t.Errorf("map lookup returned result: got %v want %v", false, ok)
	}
	delete(s.v, "Braden.Altstatt")
}

func TestAverageAllStudentsExams(t *testing.T) {
	test1 := Test{12345, 0.75}
	student := Student{[]Test{test1}, 0.75}
	test2 := Test{98765, 0.25}
	newExams := append(student.Exams, test2)
	student.Exams = newExams
	s.v["Braden.Altstatt"] = student
	AverageAllStudentsExams(s)

	info, ok := s.v["Braden.Altstatt"]

	if info.Average != 0.50 {
		t.Errorf("map lookup expected to find a new exam: got %v want %v", info.Average, 0.50)
	}

	if ok != true {
		t.Errorf("map lookup returned result: got %v want %v", false, ok)
	}
	delete(s.v, "Braden.Altstatt")
}

func TestUpdateStudentDataJSON(t *testing.T) {
	test1 := Test{12345, 0.75}
	student := Student{[]Test{test1}, 0.75}
	test2 := Test{98765, 0.25}
	newExams := append(student.Exams, test2)
	student.Exams = newExams
	s.v["Braden.Altstatt"] = student
	UpdateStudentDataJSON(s)

	json := s.json

	if len(json) != 208 {
		t.Errorf("The new json string has the wrong length: got %v want %v", len(json), 208)
	}
}

func TestGetExam(t *testing.T) {
	exam, ok := GetExam(e, "10775")

	if exam.Scores[0] != 0.7789680185088161 {
		t.Errorf("map lookup returned wrong result: got %v want %v", exam.Scores[0], 0.7789680185088161)
	}

	if ok != true {
		t.Errorf("map lookup returned result: got %v want %v", false, ok)
	}
}

func TestSetExam(t *testing.T) {
	exam := Exam{[]float64{0.123456789}}
	SetExam(e, "12345", exam)

	info, ok := e.v["12345"]

	if info.Scores[0] != 0.123456789 {
		t.Errorf("map lookup returned wrong result: got %v want %v", info.Scores[0], 0.123456789)
	}

	if ok != true {
		t.Errorf("map lookup returned result: got %v want %v", false, ok)
	}
	delete(e.v, "12345")
}

func TestAddExamScore(t *testing.T) {
	AddExamScore(e, "10775", 0.123456789)

	info, ok := e.v["10775"]

	found := false
	for _, score := range info.Scores {
		if score == 0.123456789 {
			found = true
		}
	}
	if found != true {
		t.Errorf("Did not successfully add a new score to exam: got %v want %v", found, true)
	}

	if ok != true {
		t.Errorf("map lookup returned result: got %v want %v", false, ok)
	}

}
