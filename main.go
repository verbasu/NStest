package main

import (
	"fmt"
	"time"

	"github.com/beevik/guid"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Ttype represents a meaninglessness of our life
type Ttype struct {
	id         *guid.Guid
	cT         string // время создания
	fT         string // время выполнения
	payLoad    []byte
	taskRESULT bool
}

func main() {
	taskCreaturer := func(a chan Ttype) {
		go func() {
			for {
				tn := time.Now()
				ft := []byte("good task")
				if tn.Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
					ft = []byte("bad task")
				}
				time.Sleep(time.Millisecond * 150)
				a <- Ttype{payLoad: ft, cT: tn.Format(time.RFC3339Nano), id: guid.New()} // передаем таск на выполнение
			}
		}()
	}

	superChan := make(chan Ttype, 10)

	go taskCreaturer(superChan)

	taskWorker := func(a Ttype) Ttype {
		tt, _ := time.Parse(time.RFC3339, a.cT)
		g, _ := guid.ParseString(a.id.String())
		if guid.IsGuid(g.String()) && tt.After(time.Now().Add(-20*time.Second)) {
			a.taskRESULT = true
		} else {
			a.taskRESULT = false
		}
		a.fT = time.Now().Format(time.RFC3339Nano)

		return a
	}

	doneTasks := make(chan Ttype)
	undoneTasks := make(chan error)

	taskSorter := func(t Ttype) {
		if string(t.payLoad[:4]) == "good" && t.taskRESULT {
			doneTasks <- t
		} else {
			undoneTasks <- fmt.Errorf("Task id %s time %s, error %s", t.id, t.cT, t.payLoad)
		}
	}

	go func() {
		// получение тасков
		for t := range superChan {
			t = taskWorker(t)
			go taskSorter(t)
		}
		close(superChan)
	}()

	result := map[string]Ttype{}
	go func() {
		for r := range doneTasks {
			result[r.id.String()] = r
		}
		close(doneTasks)
	}()

	err := []error{}
	go func() {
		for r := range undoneTasks {
			err = append(err, r)
		}
		close(undoneTasks)
	}()

	time.Sleep(time.Second * 3)

	println("Errors:")
	for r := range err {
		println(err[r].Error())
	}

	println("Done tasks:")
	for r := range result {
		println(result[r].id.String(), result[r].cT, result[r].fT)
	}
}
