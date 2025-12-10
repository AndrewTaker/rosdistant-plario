package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"palario/pkg/llm"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"golang.org/x/net/html"
)

var (
	subjectID, courseID, moduleID int

	plarioToken string

	info bool

	totalCorrent, totalWrong int
)

func main() {
	f, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	handler := slog.NewJSONHandler(mw, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	flag.StringVar(&plarioToken, "ptoken", "", "plario access token")
	flag.BoolVar(&info, "info", false, "print out availabe subjects, courses and modules and exit")
	flag.IntVar(&subjectID, "subject", 0, "subject_id")
	flag.IntVar(&courseID, "course", 0, "course_id")
	flag.IntVar(&moduleID, "module", 0, "module_id")
	flag.Parse()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	client := &http.Client{}
	plario := NewPlario(plarioToken)

	if plarioToken == "" {
		log.Fatal("gotta provide valid plario access token")
	}

	if info {
		subjects, err := plario.GetAvailable(client)
		if err != nil {
			logger.Error("e", "e", err)
		}

		courses := make(map[Course][]Module)
		for _, s := range subjects {
			for _, c := range s.Courses {
				plario.CourseID = c.ID
				modules, err := plario.GetModules(client)
				if err != nil {
					log.Fatal(err)
				}
				courses[c] = modules
			}
		}
		printSubjectsTable(subjects, courses)
		os.Exit(0)
	}
	plario.SubjectID = subjectID
	plario.CourseID = courseID
	plario.ModuleID = moduleID

	groq := llm.NewGroq(
		os.Getenv("GROQ_TOKEN"),
		llm.ModelOpenAIGptOss120B,
		"You are solving a test on a subject of mathematical analysis in russian. You will receive question and possible answers, it is in latex format. Only return id of correct answer, never return reasoning or any text data.",
	)

	// subjects, err := plario.GetAvailable(client)
	// if err != nil {
	// 	logger.Error("", err)
	// }

	// module belongs to a course
	// 	plario subject 								-> Высшая математика
	//	rosdistant course 							-> Математический анализ
	//	learning module (thematic set as in a book) -> Пределы дробнорациональных/иррациональных выражений
	//	------------------------------------------------------------------------------------------------------------
	//	| subject				|	course					|	module												|
	// 	| Высшая математика		|	Математический анализ	|	Пределы дробнорациональных/иррациональных выражений	|
	//	------------------------------------------------------------------------------------------------------------
	// heirerchy is
	//	-> subject
	//		-> []course
	//			-> []module
	modules, err := plario.GetModules(client)
	if err != nil {
		log.Fatal(err)
	}
	if len(modules) == 0 {
		log.Fatal("no modules")
	}

	for {
		randomSleep := RandInRange(r, 5, 10)

		question, err := plario.GetQuestion(client)
		if err != nil {
			logger.Error(err.Error())
			break
		}

		withMeta := logger.WithGroup("meta").With(
			slog.Int("course_id", plario.CourseID),
			slog.Int("module_id", plario.ModuleID),
			slog.Int("question_id", question.Exercise.ActivityID),
		)

		attempt, err := plario.GetAttempt(client, question.Exercise.ActivityID)
		if err != nil {
			withMeta.Error(err.Error())
			break
		}
		plario.Attempt = attempt

		if len(question.Exercise.PossibleAnswers) == 0 {
			withMeta.Info("no answers in response, probably a theory, submitting")
			err := plario.CompleteLesson(client, question.Exercise.ActivityID)
			if err != nil {
				withMeta.Error(err.Error())
				break
			}
			time.Sleep(time.Duration(randomSleep) * time.Second)
			continue
		}

		withMeta.Info("sending question to groq")
		groqResponse, err := groq.SendGroqRequest(client, question.Exercise.ToString())
		if err != nil {
			withMeta.Error(err.Error())
			break
		}

		if len(groqResponse.Choices) == 0 {
			withMeta.Error("groq returned bad response", err)
			break
		}

		answer, err := strconv.Atoi(groqResponse.Choices[0].Message.Content)
		if err != nil {
			withMeta.Error(err.Error())
			break
		}

		response, err := plario.PostAnswer(client, question.Exercise.ActivityID, []int{answer}, false)
		if err != nil {
			withMeta.Error(err.Error())
			break
		}

		rightAnswer := response.RightAnswerIDs[0]
		withMeta.Info(fmt.Sprintf("groq -> %d, righ -> %d", answer, rightAnswer))

		if answer != rightAnswer {
			withMeta.Warn(fmt.Sprintf("wrong answer, submited %d but right one is %d, will submite again and go further", answer, rightAnswer))
			totalWrong++
			_, err = plario.PostAnswer(client, question.Exercise.ActivityID, []int{rightAnswer}, true)
			if err != nil {
				withMeta.Error(err.Error())
				break
			}
		} else {
			withMeta.Info("first try hit")
			totalCorrent++
		}
		time.Sleep(time.Duration(randomSleep) * time.Second)
	}

	logger.Info("Total correct", "count", totalCorrent)
	logger.Info("Total wrong", "count", totalWrong)
}

func RandInRange(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}

func printSubjectsTable(objs []Subject, courses map[Course][]Module) {
	headerFmt := color.New(color.FgWhite, color.Underline, color.Bold, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	t := table.New("s_id", "s_name", "c_id", "c_name", "m_id", "m_name", "m_mastery")
	t.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, o := range objs {
		for c, modules := range courses {
			for _, m := range modules {
				t.AddRow(o.ID, o.Name, c.ID, c.Name, m.ID, m.Name, fmt.Sprintf("%.2f (%.2f%%)", m.Mastery, m.Mastery*100))
			}
		}
	}
	t.Print()
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
