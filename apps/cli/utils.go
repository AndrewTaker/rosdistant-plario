package main

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	pl "pkg/plario"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// get subjects, courses, modules and mastery levels
func QuizMode(plario *pl.Plario, client *http.Client) error {
	subjects, err := plario.GetAvailable(client)
	if err != nil {
		return err
	}

	courses := make(map[pl.Course][]pl.Module)
	for _, s := range subjects {
		for _, c := range s.Courses {
			plario.CourseID = c.ID
			modules, err := plario.GetModules(client)
			if err != nil {
				return err
			}
			courses[c] = modules
		}
	}
	printSubjectsTable(subjects, courses)
	return nil
}

func InitLogger(level string) *slog.Logger {
	var l slog.Level

	switch level {
	case "info":
		l = slog.LevelInfo
	case "debug":
		l = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(log.Writer(), &slog.HandlerOptions{Level: l})
	return slog.New(handler)
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
