package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"pkg/llm"
	pl "pkg/plario"

	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

var (
	subjectID, courseID, moduleID int
	plarioToken, groqToken        string

	infoMode, browserMode bool

	masteryCap   float64
	isMasteryCap bool

	totalCorrent, totalWrong int

	logLevel string
)

func main() {
	flag.StringVar(&plarioToken, "ptoken", "", "plario access token")
	flag.StringVar(&groqToken, "gtoken", "", "groq api token")
	flag.StringVar(&logLevel, "loglevel", "info", "provide to change log level")
	flag.IntVar(&subjectID, "subject", 0, "subject_id")
	flag.IntVar(&courseID, "course", 0, "course_id")
	flag.IntVar(&moduleID, "module", 0, "module_id")
	flag.Float64Var(&masteryCap, "till_mastery", 0.0, "provide if you want to stop program execution at certain mastery level")
	flag.BoolVar(&infoMode, "info", false, "print out availabe subjects, courses and modules and exit")
	flag.Parse()

	if plarioToken == "" || groqToken == "" {
		flag.Usage()
		os.Exit(1)
	}

	if masteryCap != 0.0 {
		isMasteryCap = true
	}

	logger := InitLogger(logLevel)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	client := &http.Client{}
	plario := pl.NewPlario(plarioToken, logger)

	if infoMode {
		err := QuizMode(plario, client)
		if err != nil {
			logger.Error("GetInfo", "message", err.Error())
		}
		return
	}

	plario.SubjectID = subjectID
	plario.CourseID = courseID
	plario.ModuleID = moduleID

	groq := llm.NewGroq(
		os.Getenv("GROQ_TOKEN"),
		llm.ModelOpenAIGptOss120B,
		"You are solving a test on a subject of mathematical analysis in russian. You will receive question and possible answers, it is in latex format. Only return id of correct answer, never return reasoning or any text data.",
		logger,
	)

	for {
		randomSleep := RandInRange(r, 5, 10)

		question, err := plario.GetQuestion(client)
		if err != nil {
			logger.Error("plario.GetQuestion", "message", err)
			break
		}

		attempt, err := plario.GetAttempt(client, question.Exercise.ActivityID)
		if err != nil {
			logger.Error(err.Error())
			break
		}
		plario.Attempt = attempt

		withMeta := logger.WithGroup("meta").With(
			slog.Int("course_id", plario.CourseID),
			slog.Int("module_id", plario.ModuleID),
			slog.Int("question_id", question.Exercise.ActivityID),
			slog.Int("attempt_id", plario.Attempt),
		)

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
			withMeta.Error("groq returned bad response", "message", err)
			break
		}

		answer, err := strconv.Atoi(groqResponse.Choices[0].Message.Content)
		if err != nil {
			withMeta.Error("could not convert atoi", "message", err.Error())
			break
		}

		response, err := plario.PostAnswer(client, question.Exercise.ActivityID, []int{answer}, false)
		if err != nil {
			withMeta.Error("p.PostAnswer", "message", err.Error())
			break
		}

		rightAnswer := response.RightAnswerIDs[0]
		withMeta.Info(fmt.Sprintf("groq -> %d, right -> %d", answer, rightAnswer))

		if answer != rightAnswer {
			withMeta.Warn(fmt.Sprintf("wrong answer, submited %d but right one is %d, will submit again and go further", answer, rightAnswer))
			totalWrong++
			_, err = plario.PostAnswer(client, question.Exercise.ActivityID, []int{rightAnswer}, true)
			if err != nil {
				withMeta.Error("p.PostAnswer", "message", err.Error())
				break
			}
		} else {
			withMeta.Info("first try hit")
			totalCorrent++
		}

		ms, err := plario.GetModules(client)
		if err != nil {
			withMeta.Error("p.GetModules", "message", err.Error())
		}

		var currentMastery float64
		for _, m := range ms {
			if m.ID == plario.ModuleID {
				currentMastery = m.Mastery
				logger.Info("mastery", "value", slog.Float64Value(m.Mastery))
				break
			}
		}

		if isMasteryCap && currentMastery >= masteryCap {
			logger.Info("mastery", "hit mastery cap", slog.Float64Value(currentMastery))
			break
		}
		time.Sleep(time.Duration(randomSleep) * time.Second)
	}

	logger.Info("Total correct", "count", totalCorrent)
	logger.Info("Total wrong", "count", totalWrong)
}

func RandInRange(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}

func printSubjectsTable(objs []pl.Subject, courses map[pl.Course][]pl.Module) {
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
