package cron

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type task struct {
	rules  [][]string
	action func()
}

var tasks = make(map[uint]task)

var mutex sync.Mutex

func init() {
	go initCron()
}

//Add 添加一个任务到自动任务列表
//
//@action：触发的函数，同一时间只会触发一次
//
//@rules：触发点 格式："ss mm hh DD MM YYYY"，可传入多个，若不传则每秒执行，是action执行间隔的最小单元
//
//例：
//
//"00"：每分
//
//"00 30 08"：每天08:30
//
//"05 30 08 14 09 2021"：2021-09-14 08:30:05
func Add(action func(), rules ...string) (id uint, err error) {
	mutex.Lock()
	defer mutex.Unlock()

	tt := task{
		rules:  [][]string{},
		action: action,
	}

	reg, err := regexp.Compile("^([0-5]\\d)(\\s[0-5]\\d(\\s(20|21|22|23|[0-1]\\d)(\\s(0[1-9]|[1-2][0-9]|3[0-1])(\\s(0[1-9]|1[0-2])(\\s\\d{4})?)?)?)?)?$")
	if err != nil {
		return
	}

	for _, rule := range rules {
		if !reg.MatchString(rule) {
			err = errors.New(fmt.Sprintf("\"%s\" rules not match", rule))
			return
		}
		tt.rules = append(tt.rules, strings.Split(rule, " "))
	}

	id = uint(len(tasks))
	tasks[id] = tt
	return
}

//Clear 清除任务
func Clear(ids ...uint) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, id := range ids {
		delete(tasks, id)
	}
}

func initCron() {
	wait()
	for {
		go exec()
		wait()
	}
}

func wait() {
	now := time.Now()
	time.Sleep(time.Duration((now.Unix()+1)*1000*1000*1000 - now.UnixNano()))
}

func exec() {
	ntr := strings.Split(time.Now().Format("05 04 15 02 01 2006"), " ")
	for _, t := range tasks {
		if len(t.rules) == 0 {
			go t.action()
			continue
		}
		for _, rule := range t.rules {
			m := true
			for i, r := range rule {
				if r != ntr[i] {
					m = false
					break
				}
			}
			if m {
				go t.action()
				break
			}
		}
	}
}
