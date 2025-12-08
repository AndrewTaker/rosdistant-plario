package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"palario/pkg/database"
	"time"
)

const MODULE_ID = 12           // name of the section, for example "Пределы дробнорациональных/иррациональных выражений"
const TEACHER_COURSE_ID = 2274 // name of the subject in uni, for example "Математический анализ"
const CULTURE = "ru"

func main() {
	db, err := database.New(context.Background(), "plario.db")
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("TOKEN")
	client := &http.Client{}
	p := NewPlario(MODULE_ID, TEACHER_COURSE_ID, token)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	available, err := p.GetAvailable(client)
	if err != nil {
		log.Fatal("GetAvailable err", err)
	}

	for _, s := range available {
		err := db.CreateSubject(s.ID, s.Name)
		if err != nil {
			log.Fatal("db err", err)
		}
		for _, c := range s.Courses {
			err = db.CreateCourse(c.ID, c.Name, s.ID)
			if err != nil {
				log.Fatal("db err", err)
			}
		}
	}

	modules, err := p.GetModules(client)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(modules)

	for _, m := range modules {
		err := db.CreateModule(m.ID, m.Name, TEACHER_COURSE_ID)
		if err != nil {
			log.Fatal(err)
		}
	}

	subjectID := available[0].ID

	for {
		randomSleep := RandInRange(r, 5, 10)

		question, err := p.GetQuestion(client)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("question:", question.Exercise.ActivityID)

		attempt, err := p.GetAttempt(client, question.Exercise.ActivityID)
		if err != nil {
			log.Fatal(err)
		}
		p.Attempt = attempt

		if len(question.Exercise.PossibleAnswers) == 0 {
			log.Println("no answers in response, probably a theory, subbmiting")
			err := p.CompleteLesson(client, question.Exercise.ActivityID)
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Duration(randomSleep) * time.Second)
			continue
		}

		log.Println("trying to get answer from db")
		answer, err := db.GetAnswer(question.Exercise.ActivityID, subjectID, TEACHER_COURSE_ID, MODULE_ID)
		if err != nil {
			log.Fatal(err)
		}

		var passedAnswer int
		if answer != 0 {
			log.Println("there is an answer in database ->", answer)
			passedAnswer = answer
		} else {
			log.Println("no answer in database, creating one for question", question.Exercise.ActivityID)
			index := rand.Intn(len(question.Exercise.PossibleAnswers))
			passedAnswer = question.Exercise.PossibleAnswers[index].AnswerID
		}

		response, err := p.PostAnswer(client, question.Exercise.ActivityID, []int{passedAnswer}, false)
		if err != nil {
			log.Fatal(err)
		}

		responseAnswer := response.RightAnswerIDs

		if passedAnswer != responseAnswer[0] {
			log.Printf("wrong answer, submited %d but right one is %d", passedAnswer, responseAnswer[0])
			log.Println("submitting again with", responseAnswer[0])

			err := db.CreateQuestion(question.Exercise.ActivityID, question.Exercise.Content, responseAnswer[0], subjectID, TEACHER_COURSE_ID, MODULE_ID)
			if err != nil {
				log.Fatal(err)
			}
			_, _ = p.PostAnswer(client, question.Exercise.ActivityID, []int{responseAnswer[0]}, true)
		}
		time.Sleep(time.Duration(randomSleep) * time.Second)
		log.Println("-----------------------------------")
	}
}

func RandInRange(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}
