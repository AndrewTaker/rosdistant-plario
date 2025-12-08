package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
)

const MODULE_ID = 12           // name of the section, for example "Пределы дробнорациональных/иррациональных выражений"
const TEACHER_COURSE_ID = 2274 // name of the subject in uni, for example "Математический анализ"
const CULTURE = "ru"

func main() {
	token := os.Getenv("TOKEN")
	// get attempt here with post'https://api.plario.ru/learner/adaptiveLearning/modules/12/activities/774/attempts'
	// get question
	// post answer with attempt
	client := &http.Client{}
	p := NewPlario(MODULE_ID, TEACHER_COURSE_ID, token)

	question, err := p.GetQuestion(client)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(question)

	attempt, err := p.GetAttempt(client, question.Exercise.ActivityID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(attempt)
	p.Attempt = attempt

	randomAnswer := rand.Intn(len(question.Exercise.PossibleAnswers))

	response, err := p.PostAnswer(client, question.Exercise.ActivityID, []int{question.Exercise.PossibleAnswers[randomAnswer].AnswerID})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(response)
}
