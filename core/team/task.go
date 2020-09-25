package team

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
	"time"
)

func _sub_key(m *ice.Message, zone string) string {
	return kit.Keys(m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, kit.Hashs(zone))
}

func _task_list(m *ice.Message, zone string, id string) {
	m.Option(mdb.FIELDS, "begin_time,zone,id,status,level,type,name,text")
	m.Cmdy(mdb.SELECT, m.Prefix(TASK), kit.Keys(m.Optionv(ice.MSG_DOMAIN)), mdb.ZONE, zone, id)
}
func _task_create(m *ice.Message, zone string) {
	m.Conf(m.Prefix(TASK), kit.Keys(m.Optionv(ice.MSG_DOMAIN), kit.MDB_META, kit.MDB_SHORT), kit.MDB_ZONE)
	m.Cmdy(mdb.INSERT, m.Prefix(TASK), kit.Keys(m.Optionv(ice.MSG_DOMAIN)), mdb.HASH, kit.MDB_ZONE, zone)
}
func _task_insert(m *ice.Message, zone string, arg ...string) {
	if msg := m.Cmd(mdb.SELECT, m.Prefix(TASK), kit.Keys(m.Optionv(ice.MSG_DOMAIN)), mdb.HASH, kit.MDB_ZONE, zone); len(msg.Appendv(kit.MDB_HASH)) == 0 {
		m.Debug("what %v", msg.Formats("meta"))
		_task_create(m, zone)
	}

	m.Cmdy(mdb.INSERT, m.Prefix(TASK), _sub_key(m, zone), mdb.LIST,
		BEGIN_TIME, m.Time(), CLOSE_TIME, m.Time("30m"), STATUS, StatusPrepare, LEVEL, 3, SCORE, 3, arg,
	)
}
func _task_modify(m *ice.Message, zone, id, field, value string, arg ...string) {
	if field == STATUS {
		switch value {
		case StatusProcess:
			arg = append(arg, BEGIN_TIME, m.Time())
		case StatusCancel, StatusFinish:
			arg = append(arg, CLOSE_TIME, m.Time())
		}
	}
	m.Cmdy(mdb.MODIFY, m.Prefix(TASK), _sub_key(m, zone), mdb.LIST, kit.MDB_ID, id, field, value, arg)
}
func _task_delete(m *ice.Message, zone, id string) {
	m.Cmdy(mdb.MODIFY, m.Prefix(TASK), _sub_key(m, zone), mdb.LIST, kit.MDB_ID, id, STATUS, StatusCancel)
}
func _task_export(m *ice.Message, file string) {
	m.Option(mdb.FIELDS, "zone,id,time,type,name,text,level,status,score,begin_time,close_time,extra")
	m.Cmdy(mdb.EXPORT, m.Prefix(TASK), kit.Keys(m.Optionv(ice.MSG_DOMAIN)), mdb.ZONE)
}
func _task_import(m *ice.Message, file string) {
	m.Option(mdb.FIELDS, kit.MDB_ZONE)
	m.Cmdy(mdb.IMPORT, m.Prefix(TASK), kit.Keys(m.Optionv(ice.MSG_DOMAIN)), mdb.ZONE, file)
}
func _task_input(m *ice.Message, field, value string) {
	switch field {
	case kit.MDB_ZONE:
		m.Cmdy(mdb.INPUTS, m.Prefix(TASK), kit.Keys(m.Optionv(ice.MSG_DOMAIN)), mdb.HASH, field, value)
	case kit.MDB_NAME, kit.MDB_TEXT:
		m.Cmdy(mdb.INPUTS, m.Prefix(TASK), _sub_key(m, m.Option(kit.MDB_ZONE)), mdb.LIST, field, value)
	}
}

func _task_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			if value[kit.MDB_NAME] == name || strings.Contains(kit.Format(value[kit.MDB_TEXT]), name) {
				m.Push("pod", m.Option(ice.MSG_USERPOD))
				m.Push("ctx", m.Prefix())
				m.Push("cmd", TASK)
				m.Push("time", value[kit.MDB_TIME])
				m.Push("size", 1)
				m.Push("type", TASK)
				m.Push("name", value[kit.MDB_NAME])
				m.Push("text", kit.Format("%s:%d", kit.Value(val, "meta.zone"), kit.Int(value[kit.MDB_ID])))

			}
		})
	})
}
func _task_render(m *ice.Message, kind, name, text string, arg ...string) {
	ls := strings.Split(text, ":")
	m.Richs(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), ls[0], func(key string, val map[string]interface{}) {
		m.Grows(TASK, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), "id", ls[1], func(index int, value map[string]interface{}) {
			m.Push("detail", value)
		})
	})
}
func _task_action(m *ice.Message, status interface{}, action ...string) []string {
	switch status {
	case StatusPrepare:
		action = append(action, "开始")
	case StatusProcess:
		action = append(action, "完成")
	case StatusCancel, StatusFinish:
	}
	return action
}
func _task_scope(m *ice.Message, arg ...string) (time.Time, time.Time) {
	begin_time := time.Now()
	if len(arg) > 1 {
		begin_time, _ = time.ParseInLocation(ice.MOD_TIME, arg[1], time.Local)
	}
	end_time := begin_time

	switch begin_time = begin_time.Add(-time.Duration(begin_time.UnixNano()) % (24 * time.Hour)); arg[0] {
	case ScaleDay:
		end_time = begin_time.AddDate(0, 0, 1)
	case ScaleWeek:
		begin_time = begin_time.AddDate(0, 0, -int(begin_time.Weekday()))
		end_time = begin_time.AddDate(0, 0, 7)
	case ScaleMonth:
		begin_time = begin_time.AddDate(0, 0, -begin_time.Day()+1)
		end_time = begin_time.AddDate(0, 1, 0)
	case ScaleYear:
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1)
		end_time = begin_time.AddDate(1, 0, 0)
	case ScaleLong:
		begin_time = begin_time.AddDate(0, 0, -begin_time.YearDay()+1)
		begin_time = begin_time.AddDate(-begin_time.Year()%5, 0, 0)
		end_time = begin_time.AddDate(5, 0, 0)
	}

	begin_time = begin_time.Add(-8 * time.Hour)
	end_time = end_time.Add(-8 * time.Hour)
	return begin_time, end_time
}

const (
	LEVEL  = "level"
	STATUS = "status"
	SCORE  = "score"

	BEGIN_TIME = "begin_time"
	CLOSE_TIME = "close_time"

	EXPORT = "usr/export/web.team.task/"
)
const (
	ScaleDay   = "day"
	ScaleWeek  = "week"
	ScaleMonth = "month"
	ScaleYear  = "year"
	ScaleLong  = "long"
)
const (
	StatusPrepare = "prepare"
	StatusProcess = "process"
	StatusCancel  = "cancel"
	StatusFinish  = "finish"
)
const TASK = "task"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TASK: {Name: "task", Help: "task", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
		},
		Commands: map[string]*ice.Command{
			TASK: {Name: "task zone id auto 添加 导出 导入", Help: "任务", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert zone@key type=once,step,week name@key text@key begin_time@date close_time@date", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_task_insert(m, arg[1], arg[2:]...)
				}},
				mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1])
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					_task_delete(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					_task_export(m, kit.Select(path.Join(EXPORT, m.Option(ice.MSG_DOMAIN), "list.csv"), arg, 0))
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_task_import(m, kit.Select(path.Join(EXPORT, m.Option(ice.MSG_DOMAIN), "list.csv"), arg, 0))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					_task_input(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				}},

				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_task_search(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},
				mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					_task_render(m, arg[0], arg[1], arg[2], arg[3:]...)
				}},

				gdb.BEGIN: {Name: "begin", Help: "开始", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, StatusProcess)
				}},
				gdb.END: {Name: "end", Help: "完成", Hand: func(m *ice.Message, arg ...string) {
					_task_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), STATUS, StatusFinish)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(mdb.SELECT, m.Prefix(TASK), m.Option(ice.MSG_DOMAIN), mdb.ZONE, arg)
				if len(arg) > 0 {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushRender("action", "button", "", _task_action(m, value[STATUS])...)
					})
				}
			}},
		},
	}, nil)
}
