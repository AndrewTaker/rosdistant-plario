package plario

import (
	"bytes"
	"encoding/json"

	"golang.org/x/net/html"
)

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
	ActivityID      int   `json:"activityId,omitempty"`
	AttemptID       int   `json:"attemptId,omitempty"`
	AnswerIDs       []int `json:"answerIds,omitempty"`
	ModuleID        int   `json:"moduleId,omitempty"`
	TeacherCourseID int   `json:"teacherCourseId,omitempty"`
}

type PlarioAnswerResponse struct {
	RightAnswerIDs []int `json:"rightAnswerIds"`
}

type Module struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Mastery float64 `json:"mastery"`
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

func StripHTMLKeepLatex(s string) string {
	var b bytes.Buffer
	z := html.NewTokenizer(bytes.NewBufferString(s))

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return b.String()
		case html.TextToken:
			b.Write(z.Text())
		}
	}
}
