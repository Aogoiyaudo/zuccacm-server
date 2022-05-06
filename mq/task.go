package mq

import "zuccacm-server/db"

func SubmissionTask(username []string, count int, group []string, groupCount int) (t *Task) {
	t = newTask()
	t.mustSet("submission", "task_type")
	if len(username) > 0 {
		t.mustSet(username, "username")
		t.mustSet(count, "count")
	}
	if len(group) > 0 {
		t.mustSet(group, "group")
		t.mustSet(groupCount, "group_count")
	}
	return
}

func ContestTask(ojId, id int, cid string) {
	t := newTask()
	t.mustSet("contest", "task_type")
	t.mustSet(id, "id")
	t.mustSet(cid, "cid")
	if ojId == db.GetOJByName("codeforces").OjId {
		t.mustSet("5H0hEjEiuF", "group")
	}
	ExecTask(Topic(ojId), t)
}

func RatingTask(username []string) (t *Task) {
	t = newTask()
	t.mustSet("rating", "task_type")
	t.mustSet(username, "username")
	return
}
