package task

import (
	"log"
	"myiotqq-plugin/strategy"
	"time"
)

type TaskFunc func()

type Task struct {
	name string
	time time.Duration
	f    TaskFunc
}

type TaskTable map[string]Task

var taskTable TaskTable = map[string]Task{
	"news":  {"news", time.Hour * 2, pushNews},
	"rezan": {"rezan", time.Hour * 24, resetzan},
}

func Start() {
	for _, v := range taskTable {
		go periodlyCall(v)
	}
}

func pushNews() {

}

func resetzan() {
	l := len(strategy.Zanok)
	for m := 0; m < l; m++ {
		i := 0
		strategy.Zanok = append(strategy.Zanok[:i], strategy.Zanok[i+1:]...)
	}
}

func periodlyCall(t Task) {
	for range time.Tick(t.time) {
		log.Println("定时任务：" + t.name)
		t.f()
	}
}
