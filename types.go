package main

type Exercise struct {
	ActivityID      int              `json:"activityId"`
	Content         string           `json:"content"`
	PossibleAnswers []PossibleAnswer `json:"possibleAnswers"`
}

type PossibleAnswer struct {
	AnswerID  int  `json:"answerId"`
	IsCorrect bool `json:"isCorrect"`
}

type PlarioQuestionResponse struct {
	ActivityStatus string   `json:"activityStatus"`
	Exercise       Exercise `json:"exercise"`
}

// --data-raw '{"activityId":856,"attemptId":9188881,"answerIds":[250],"moduleId":12,"teacherCourseId":2274}'
type PlarionAnswerRequest struct {
	ActivityID      int   `json:"activityId"`
	AttemptID       int   `json:"attemptId"`
	AnswerIDs       []int `json:"answerIds"`
	ModuleID        int   `json:"moduleId"`
	TeacherCourseID int   `json:"teacherCourseId"`
}

type PlarioAnswerResponse struct {
	RightAnswerIDs []int `json:"rightAnswerIds"`
}

type Module struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Course struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Subject struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Courses []Course `json:"courses"`
}
