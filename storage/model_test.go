package storage

import (
	"fmt"
	"github.com/fatih/structs"
	"testing"
)

func TestFindOne(t *testing.T) {
	task := new(Task)

	Db.Where("tid = ?", "f3729718").Find(&task)

	t.Log(task)
}

func TestGetTasks(t *testing.T) {
	q := TaskQuery{}

	q.Cid = 1
	q.Count = 1
	q.Index = 2
	q.Order = "tid desc"

	tasks, err := GetTasks(&q)
	if err != nil {
		t.Fatal(err)
	}

	for _, task := range tasks {
		t.Log(task)
	}
}

func TestGetLogs(t *testing.T) {
	q := LogQuery{}
	q.Tid = 3

	logs, err := GetLogs(&q)

	if err != nil {
		t.Fatal(err)
	}

	for _, task := range logs {
		t.Log(task)
	}
}

func TestGetFields(t *testing.T) {
	var query TaskQuery

	s := structs.New(&query)
	for _, key := range s.Names() {
		tmp := s.Field(key)
		fields := tmp.Fields()
		for _, f := range fields {
			t.Log(f.Tag("json"))
			kind := fmt.Sprintf("%v", f.Kind())
			t.Log(kind)
			t.Log(f.Value())
		}
	}

}
