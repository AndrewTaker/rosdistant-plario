package main

import "encoding/json"

type Exercise struct {
	ActivityID      int              `json:"activityId"`
	Content         string           `json:"content"`
	PossibleAnswers []PossibleAnswer `json:"possibleAnswers"`
}

func (e *Exercise) ToString() string {
	type Answer struct {
		ID     int    `json:"id"`
		Option string `json:"answer"`
	}
	type Quiz struct {
		Question string   `json:"question"`
		Answers  []Answer `json:"answers"`
	}

	var quiz Quiz
	quiz.Question = StripHTMLKeepLatex(e.Content)
	for _, i := range e.PossibleAnswers {
		quiz.Answers = append(quiz.Answers, Answer{ID: i.AnswerID, Option: StripHTMLKeepLatex(i.Text)})
	}
	s, _ := json.Marshal(quiz)
	return string(s)
}

// func (e *Exercise) ToString() string {
// 	s := fmt.Sprintf("Question: %s. Possible answers: ", StripHTMLKeepLatex(e.Content))
// 	if len(e.PossibleAnswers) > 0 {
// 		for i, a := range e.PossibleAnswers {
// 			if i == 0 {
// 				s += "["
// 			}
// 			text := StripHTMLKeepLatex(a.Text)
// 			s += fmt.Sprintf("text: %s id: %d,", text, a.AnswerID)
// 		}
// 		s += "]"
// 	} else {
// 		s += "[]"
// 	}
//
// 	return s
// }

type PossibleAnswer struct {
	AnswerID  int    `json:"answerId"`
	IsCorrect bool   `json:"isCorrect"`
	Text      string `json:"text"`
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
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Mastery float32 `json:"mastery"`
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
