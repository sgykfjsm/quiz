package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Problems []Problem

type Problem struct {
	Question string
	Answer   string
}

func (p Problems) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	for i := range p {
		j := rand.Intn(i + 1)
		p[i], p[j] = p[j], p[i]
	}
}

func NewProblems(problems [][]string) Problems {
	ret := make([]Problem, len(problems))
	for i, problem := range problems {
		ret[i] = Problem{
			Question: strings.TrimSpace(problem[0]),
			Answer:   strings.ToLower(strings.TrimSpace(problem[1])),
		}
	}

	return ret
}

func main() {
	csvPath := flag.String("csv", "problems.csv", `a csv file in the format of 'question, answer'`)
	limit := flag.Int("limit", 30, `the time limit for the quiz in seconds (default 30)`)
	random := flag.Bool("random", false, "shuffle the quiz order (default false)")
	flag.Parse()

	if _, err := os.Stat(*csvPath); os.IsNotExist(err) {
		log.Fatal(err)
	}

	f, err := os.Open(*csvPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	data, err := r.ReadAll()
	problems := NewProblems(data)
	if *random {
		problems.Shuffle()
	}

	var correct int
	answerChan := make(chan string)
problemLoop:
	for i, problem := range problems {
		fmt.Printf("Problem #%d: %s = ", i+1, problem.Question)

		go func() {
			input := bufio.NewReader(os.Stdin)
			answer, err := input.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			answerChan <- strings.TrimSpace(answer)
		}()

		t := time.NewTimer(time.Duration(*limit) * time.Second)
		select {
		case <-t.C:
			fmt.Println()
			fmt.Println("Oops, time is over!")
			break problemLoop
		case answer := <-answerChan:
			t.Stop()
			if strings.ToLower(answer) == problem.Answer {
				correct++
			}
		}
	}

	fmt.Printf("You scored %d out of %d.\n", correct, len(problems))
}
