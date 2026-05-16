package main

import (
	"database/sql"
	"time"

	_ "gorm.io/gorm"
)

type DateRange struct {
	From *time.Time
	To   *time.Time
}

func (dr DateRange) where(col string, args *[]interface{}) string {
	clause := ""
	if dr.From != nil {
		*args = append(*args, *dr.From)
		clause += " AND " + col + " >= $" + itoa(len(*args))
	}
	if dr.To != nil {
		*args = append(*args, *dr.To)
		clause += " AND " + col + " <= $" + itoa(len(*args))
	}
	return clause
}

func itoa(n int) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	return string(buf[i:])
}

type UserKPI struct {
	UserID         string     `json:"user_id"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Email          string     `json:"email"`
	UserType       string     `json:"user_type"`
	GroupID        *string    `json:"group_id"`
	GroupName      *string    `json:"group_name"`
	TasksAssigned  int        `json:"tasks_assigned"`
	TasksCompleted int        `json:"tasks_completed"`
	TasksOnTime    int        `json:"tasks_on_time"`
	OnTimePercent  float64    `json:"on_time_percent"`
	AvgDifficulty  float64    `json:"avg_difficulty"`
	AvgQuality     float64    `json:"avg_quality"`
	CommentsCount  int        `json:"comments_count"`
	MessagesCount  int        `json:"messages_count"`
	ActivityScore  int        `json:"activity_score"`
	ProjectsActive int        `json:"projects_active"`
	LastActive     *time.Time `json:"last_active"`
	ActiveDays30   int        `json:"active_days_30"`
	RegularityPct  float64    `json:"regularity_pct"`
}

type UserKPIFilter struct {
	GroupID  string
	UserType string
}

func (r *ReportRepository) UserKPIs(projectFilter string, dr DateRange) ([]UserKPI, error) {
	return r.UserKPIsFiltered(projectFilter, dr, UserKPIFilter{})
}

func (r *ReportRepository) UserKPIsFiltered(projectFilter string, dr DateRange, f UserKPIFilter) ([]UserKPI, error) {
	args := []interface{}{}
	pTask, pComment, pPart, pMsg, pAct := "", "", "", "", ""
	if projectFilter != "" {
		args = append(args, projectFilter)
		idx := itoa(len(args))
		pTask = " AND pt.project_id = $" + idx
		pComment = " AND pt.project_id = $" + idx
		pPart = " AND pp.project_id = $" + idx
		pMsg = " AND pc.project_id = $" + idx
		pAct = " AND al.project_id = $" + idx
	}
	userFilter := ""
	if f.GroupID != "" {
		args = append(args, f.GroupID)
		userFilter += " AND u.group_id = $" + itoa(len(args))
	}
	if f.UserType != "" {
		args = append(args, f.UserType)
		userFilter += " AND u.user_type = $" + itoa(len(args))
	}

	completionDateClause := dr.where("pt.completion_date", &args)
	commentDateClause := dr.where("tc.publication_date", &args)
	msgDateClause := dr.where("cm.sent_at", &args)
	actDateClause := dr.where("al.occurred_at", &args)

	q := `
		WITH tasks_stats AS (
			SELECT pt.assigned_to AS uid,
				COUNT(*) FILTER (WHERE pt.assigned_to IS NOT NULL) AS assigned,
				COUNT(*) FILTER (WHERE pt.status = 'завершена') AS done,
				COUNT(*) FILTER (WHERE pt.status = 'завершена' AND pt.due_date IS NOT NULL AND pt.completion_date <= pt.due_date) AS on_time,
				COUNT(*) FILTER (WHERE pt.status = 'завершена' AND pt.due_date IS NOT NULL) AS done_with_due,
				COALESCE(AVG(pt.difficulty) FILTER (WHERE pt.status = 'завершена'), 0) AS avg_diff,
				COALESCE(AVG(pt.quality_rating) FILTER (WHERE pt.quality_rating IS NOT NULL), 0) AS avg_qual
			FROM project_tasks pt
			WHERE pt.assigned_to IS NOT NULL` + pTask + completionDateClause + `
			GROUP BY pt.assigned_to
		),
		comment_stats AS (
			SELECT tc.user_id AS uid, COUNT(*) AS c
			FROM task_comments tc
			JOIN project_tasks pt ON pt.id = tc.task_id
			WHERE 1=1` + pComment + commentDateClause + `
			GROUP BY tc.user_id
		),
		msg_stats AS (
			SELECT cm.user_id AS uid, COUNT(*) AS c
			FROM chat_messages cm
			JOIN project_chats pc ON pc.id = cm.chat_id
			WHERE 1=1` + pMsg + msgDateClause + `
			GROUP BY cm.user_id
		),
		activity_stats AS (
			SELECT al.user_id AS uid, COUNT(*) AS c
			FROM activity_log al
			WHERE 1=1` + pAct + actDateClause + `
			GROUP BY al.user_id
		),
		participation_stats AS (
			SELECT pp.user_id AS uid, COUNT(DISTINCT pp.project_id) FILTER (WHERE p.status = 'активен') AS act
			FROM project_participations pp
			JOIN projects p ON p.id = pp.project_id
			WHERE 1=1` + pPart + `
			GROUP BY pp.user_id
		),
		engagement_stats AS (
			SELECT al.user_id AS uid,
				MAX(al.occurred_at) AS last_active,
				COUNT(DISTINCT al.occurred_at::date) FILTER (
					WHERE al.occurred_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
				) AS active_days_30
			FROM activity_log al
			WHERE 1=1` + pAct + `
			GROUP BY al.user_id
		)
		SELECT u.id, u.first_name, u.last_name, u.email,
			COALESCE(u.user_type, 'student'), u.group_id, g.name AS group_name,
			COALESCE(ts.assigned, 0),
			COALESCE(ts.done, 0),
			COALESCE(ts.on_time, 0),
			CASE WHEN COALESCE(ts.done_with_due,0) > 0 THEN ts.on_time::float * 100 / ts.done_with_due ELSE 0 END,
			COALESCE(ts.avg_diff, 0),
			COALESCE(ts.avg_qual, 0),
			COALESCE(cs.c, 0),
			COALESCE(ms.c, 0),
			COALESCE(cs.c, 0) + COALESCE(ms.c, 0) + COALESCE(acs.c, 0),
			COALESCE(ps.act, 0),
			es.last_active,
			COALESCE(es.active_days_30, 0),
			COALESCE(es.active_days_30, 0) * 100.0 / 30
		FROM users u
		LEFT JOIN groups g ON g.id = u.group_id
		LEFT JOIN tasks_stats ts ON ts.uid = u.id
		LEFT JOIN comment_stats cs ON cs.uid = u.id
		LEFT JOIN msg_stats ms ON ms.uid = u.id
		LEFT JOIN activity_stats acs ON acs.uid = u.id
		LEFT JOIN participation_stats ps ON ps.uid = u.id
		LEFT JOIN engagement_stats es ON es.uid = u.id
		WHERE u.status = true` + userFilter + `
		ORDER BY (COALESCE(ts.done, 0) * 5 + COALESCE(ts.on_time, 0) * 3 + COALESCE(cs.c, 0) + COALESCE(ms.c, 0)) DESC,
			u.last_name, u.first_name`

	rows, err := r.db.Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []UserKPI{}
	for rows.Next() {
		var k UserKPI
		var lastActive sql.NullTime
		if err := rows.Scan(&k.UserID, &k.FirstName, &k.LastName, &k.Email,
			&k.UserType, &k.GroupID, &k.GroupName,
			&k.TasksAssigned, &k.TasksCompleted, &k.TasksOnTime, &k.OnTimePercent,
			&k.AvgDifficulty, &k.AvgQuality, &k.CommentsCount, &k.MessagesCount, &k.ActivityScore,
			&k.ProjectsActive, &lastActive, &k.ActiveDays30, &k.RegularityPct); err == nil {
			if lastActive.Valid {
				v := lastActive.Time
				k.LastActive = &v
			}
			out = append(out, k)
		}
	}
	return out, nil
}

type GroupKPI struct {
	GroupID         *string `json:"group_id"`
	GroupName       string  `json:"group_name"`
	UsersCount      int     `json:"users_count"`
	ActiveUsers     int     `json:"active_users"`
	TasksCompleted  int     `json:"tasks_completed"`
	TasksOnTime     int     `json:"tasks_on_time"`
	OnTimePercent   float64 `json:"on_time_percent"`
	AvgQuality      float64 `json:"avg_quality"`
	AvgDifficulty   float64 `json:"avg_difficulty"`
	CommentsCount   int     `json:"comments_count"`
	MessagesCount   int     `json:"messages_count"`
}

func (r *ReportRepository) GroupKPIs(dr DateRange) ([]GroupKPI, error) {
	args := []interface{}{}
	completionClause := dr.where("pt.completion_date", &args)
	commentClause := dr.where("tc.publication_date", &args)
	msgClause := dr.where("cm.sent_at", &args)

	q := `
		WITH tasks_by_user AS (
			SELECT pt.assigned_to AS uid,
				COUNT(*) FILTER (WHERE pt.status = 'завершена') AS done,
				COUNT(*) FILTER (WHERE pt.status = 'завершена' AND pt.due_date IS NOT NULL AND pt.completion_date <= pt.due_date) AS on_time,
				COUNT(*) FILTER (WHERE pt.status = 'завершена' AND pt.due_date IS NOT NULL) AS done_with_due,
				COALESCE(AVG(pt.quality_rating) FILTER (WHERE pt.quality_rating IS NOT NULL), 0) AS avg_qual,
				COALESCE(AVG(pt.difficulty) FILTER (WHERE pt.status = 'завершена'), 0) AS avg_diff
			FROM project_tasks pt
			WHERE pt.assigned_to IS NOT NULL` + completionClause + `
			GROUP BY pt.assigned_to
		),
		comments_by_user AS (
			SELECT tc.user_id AS uid, COUNT(*) AS c
			FROM task_comments tc WHERE 1=1` + commentClause + ` GROUP BY tc.user_id
		),
		msgs_by_user AS (
			SELECT cm.user_id AS uid, COUNT(*) AS c
			FROM chat_messages cm WHERE 1=1` + msgClause + ` GROUP BY cm.user_id
		)
		SELECT u.group_id, COALESCE(g.name, 'Без группы') AS group_name,
			COUNT(DISTINCT u.id) AS users_count,
			COUNT(DISTINCT u.id) FILTER (
				WHERE COALESCE(tbu.done, 0) > 0 OR COALESCE(cbu.c, 0) > 0 OR COALESCE(mbu.c, 0) > 0
			) AS active_users,
			COALESCE(SUM(tbu.done), 0) AS tasks_done,
			COALESCE(SUM(tbu.on_time), 0) AS tasks_on_time,
			CASE WHEN COALESCE(SUM(tbu.done_with_due), 0) > 0
				THEN COALESCE(SUM(tbu.on_time), 0)::float * 100 / SUM(tbu.done_with_due) ELSE 0 END AS on_time_pct,
			COALESCE(AVG(NULLIF(tbu.avg_qual, 0)), 0) AS avg_quality,
			COALESCE(AVG(NULLIF(tbu.avg_diff, 0)), 0) AS avg_difficulty,
			COALESCE(SUM(cbu.c), 0) AS comments,
			COALESCE(SUM(mbu.c), 0) AS messages
		FROM users u
		LEFT JOIN groups g ON g.id = u.group_id
		LEFT JOIN tasks_by_user tbu ON tbu.uid = u.id
		LEFT JOIN comments_by_user cbu ON cbu.uid = u.id
		LEFT JOIN msgs_by_user mbu ON mbu.uid = u.id
		WHERE u.status = true
		GROUP BY u.group_id, g.name
		ORDER BY g.name NULLS LAST`

	rows, err := r.db.Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GroupKPI{}
	for rows.Next() {
		var k GroupKPI
		if err := rows.Scan(&k.GroupID, &k.GroupName, &k.UsersCount, &k.ActiveUsers,
			&k.TasksCompleted, &k.TasksOnTime, &k.OnTimePercent,
			&k.AvgQuality, &k.AvgDifficulty, &k.CommentsCount, &k.MessagesCount); err == nil {
			out = append(out, k)
		}
	}
	return out, nil
}

type UserTypeKPI struct {
	UserType       string  `json:"user_type"`
	UsersCount     int     `json:"users_count"`
	ActiveUsers    int     `json:"active_users"`
	TasksCompleted int     `json:"tasks_completed"`
	TasksOnTime    int     `json:"tasks_on_time"`
	OnTimePercent  float64 `json:"on_time_percent"`
	AvgQuality     float64 `json:"avg_quality"`
}

func (r *ReportRepository) UserTypeKPIs(dr DateRange) ([]UserTypeKPI, error) {
	args := []interface{}{}
	completionClause := dr.where("pt.completion_date", &args)
	commentClause := dr.where("tc.publication_date", &args)
	msgClause := dr.where("cm.sent_at", &args)

	q := `
		WITH tasks_by_user AS (
			SELECT pt.assigned_to AS uid,
				COUNT(*) FILTER (WHERE pt.status = 'завершена') AS done,
				COUNT(*) FILTER (WHERE pt.status = 'завершена' AND pt.due_date IS NOT NULL AND pt.completion_date <= pt.due_date) AS on_time,
				COUNT(*) FILTER (WHERE pt.status = 'завершена' AND pt.due_date IS NOT NULL) AS done_with_due,
				COALESCE(AVG(pt.quality_rating) FILTER (WHERE pt.quality_rating IS NOT NULL), 0) AS avg_qual
			FROM project_tasks pt
			WHERE pt.assigned_to IS NOT NULL` + completionClause + `
			GROUP BY pt.assigned_to
		),
		comments_by_user AS (
			SELECT tc.user_id AS uid, COUNT(*) AS c FROM task_comments tc
			WHERE 1=1` + commentClause + ` GROUP BY tc.user_id
		),
		msgs_by_user AS (
			SELECT cm.user_id AS uid, COUNT(*) AS c FROM chat_messages cm
			WHERE 1=1` + msgClause + ` GROUP BY cm.user_id
		)
		SELECT COALESCE(u.user_type, 'student') AS ut,
			COUNT(DISTINCT u.id) AS users_count,
			COUNT(DISTINCT u.id) FILTER (
				WHERE COALESCE(tbu.done, 0) > 0 OR COALESCE(cbu.c, 0) > 0 OR COALESCE(mbu.c, 0) > 0
			) AS active_users,
			COALESCE(SUM(tbu.done), 0) AS tasks_done,
			COALESCE(SUM(tbu.on_time), 0) AS tasks_on_time,
			CASE WHEN COALESCE(SUM(tbu.done_with_due), 0) > 0
				THEN COALESCE(SUM(tbu.on_time), 0)::float * 100 / SUM(tbu.done_with_due) ELSE 0 END AS on_time_pct,
			COALESCE(AVG(NULLIF(tbu.avg_qual, 0)), 0) AS avg_quality
		FROM users u
		LEFT JOIN tasks_by_user tbu ON tbu.uid = u.id
		LEFT JOIN comments_by_user cbu ON cbu.uid = u.id
		LEFT JOIN msgs_by_user mbu ON mbu.uid = u.id
		WHERE u.status = true
		GROUP BY ut
		ORDER BY ut`

	rows, err := r.db.Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []UserTypeKPI{}
	for rows.Next() {
		var k UserTypeKPI
		if err := rows.Scan(&k.UserType, &k.UsersCount, &k.ActiveUsers,
			&k.TasksCompleted, &k.TasksOnTime, &k.OnTimePercent, &k.AvgQuality); err == nil {
			out = append(out, k)
		}
	}
	return out, nil
}

type ProjectKPI struct {
	ProjectID         string     `json:"project_id"`
	Title             string     `json:"title"`
	Status            string     `json:"status"`
	OrganizerName     string     `json:"organizer_name"`
	CreationDate      time.Time  `json:"creation_date"`
	PlannedEndDate    *time.Time `json:"planned_end_date"`
	CompletionDate    *time.Time `json:"completion_date"`
	ParticipantsCount int        `json:"participants_count"`
	ActiveParticipants int       `json:"active_participants"`
	TasksTotal        int        `json:"tasks_total"`
	TasksCompleted    int        `json:"tasks_completed"`
	CompletionRate    float64    `json:"completion_rate"`
	OnTimeRate        float64    `json:"on_time_rate"`
	AvgTaskDays       float64    `json:"avg_task_days"`
	AvgQuality        float64    `json:"avg_quality"`
	GoalsTotal        int        `json:"goals_total"`
	GoalsAchieved     int        `json:"goals_achieved"`
	GoalsRate         float64    `json:"goals_rate"`
	OnSchedule        *bool      `json:"on_schedule"`
	DaysToDeadline    *int       `json:"days_to_deadline"`
}

func (r *ReportRepository) ProjectKPIs(dr DateRange) ([]ProjectKPI, error) {
	args := []interface{}{}
	taskDateClause := dr.where("pt.completion_date", &args)
	creationClause := dr.where("p.creation_date", &args)

	q := `
		SELECT
			p.id, p.title, p.status,
			COALESCE(u.first_name || ' ' || u.last_name, 'Неизвестно'),
			p.creation_date, p.planned_end_date, p.completion_date,
			COALESCE((SELECT COUNT(*) FROM project_participations WHERE project_id = p.id), 0),
			COALESCE((
				SELECT COUNT(DISTINCT pp.user_id)
				FROM project_participations pp
				WHERE pp.project_id = p.id AND (
					EXISTS (SELECT 1 FROM project_tasks pt WHERE pt.project_id = p.id AND pt.assigned_to = pp.user_id AND pt.status = 'завершена'` + taskDateClause + `)
					OR EXISTS (SELECT 1 FROM task_comments tc JOIN project_tasks pt ON pt.id = tc.task_id WHERE pt.project_id = p.id AND tc.user_id = pp.user_id)
				)
			), 0) AS active_participants,
			COALESCE((SELECT COUNT(*) FROM project_tasks WHERE project_id = p.id), 0),
			COALESCE((SELECT COUNT(*) FROM project_tasks WHERE project_id = p.id AND status = 'завершена'), 0),
			CASE WHEN (SELECT COUNT(*) FROM project_tasks WHERE project_id = p.id) > 0
				THEN (SELECT COUNT(*) FROM project_tasks WHERE project_id = p.id AND status = 'завершена')::float * 100
					/ (SELECT COUNT(*) FROM project_tasks WHERE project_id = p.id)
				ELSE 0 END AS completion_rate,
			CASE WHEN (SELECT COUNT(*) FROM project_tasks WHERE project_id = p.id AND status = 'завершена' AND due_date IS NOT NULL) > 0
				THEN (SELECT COUNT(*) FROM project_tasks WHERE project_id = p.id AND status = 'завершена' AND due_date IS NOT NULL AND completion_date <= due_date)::float * 100
					/ (SELECT COUNT(*) FROM project_tasks WHERE project_id = p.id AND status = 'завершена' AND due_date IS NOT NULL)
				ELSE 0 END AS on_time_rate,
			COALESCE((
				SELECT AVG(EXTRACT(EPOCH FROM (pt.completion_date::timestamp - pt.creation_date::timestamp)) / 86400.0)
				FROM project_tasks pt WHERE pt.project_id = p.id AND pt.status = 'завершена' AND pt.completion_date IS NOT NULL` + taskDateClause + `
			), 0) AS avg_task_days,
			COALESCE((SELECT AVG(quality_rating) FROM project_tasks WHERE project_id = p.id AND quality_rating IS NOT NULL), 0),
			COALESCE((SELECT COUNT(*) FROM project_goals WHERE project_id = p.id), 0),
			COALESCE((SELECT COUNT(*) FROM project_goals WHERE project_id = p.id AND is_achieved = true), 0),
			CASE WHEN (SELECT COUNT(*) FROM project_goals WHERE project_id = p.id) > 0
				THEN (SELECT COUNT(*) FROM project_goals WHERE project_id = p.id AND is_achieved = true)::float * 100
					/ (SELECT COUNT(*) FROM project_goals WHERE project_id = p.id)
				ELSE 0 END AS goals_rate,
			CASE
				WHEN p.planned_end_date IS NULL THEN NULL
				WHEN p.completion_date IS NOT NULL THEN (p.completion_date <= p.planned_end_date)
				ELSE (CURRENT_DATE <= p.planned_end_date)
			END AS on_schedule,
			CASE
				WHEN p.planned_end_date IS NULL OR p.completion_date IS NOT NULL THEN NULL
				ELSE (p.planned_end_date - CURRENT_DATE)
			END AS days_to_deadline
		FROM projects p
		LEFT JOIN users u ON u.id = p.organizer_id
		WHERE 1=1` + creationClause + `
		ORDER BY p.creation_date DESC`

	rows, err := r.db.Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ProjectKPI{}
	for rows.Next() {
		var k ProjectKPI
		var onSchedule sql.NullBool
		var daysToDeadline sql.NullInt32
		if err := rows.Scan(&k.ProjectID, &k.Title, &k.Status, &k.OrganizerName,
			&k.CreationDate, &k.PlannedEndDate, &k.CompletionDate,
			&k.ParticipantsCount, &k.ActiveParticipants, &k.TasksTotal, &k.TasksCompleted,
			&k.CompletionRate, &k.OnTimeRate, &k.AvgTaskDays, &k.AvgQuality,
			&k.GoalsTotal, &k.GoalsAchieved, &k.GoalsRate,
			&onSchedule, &daysToDeadline); err == nil {
			if onSchedule.Valid {
				v := onSchedule.Bool
				k.OnSchedule = &v
			}
			if daysToDeadline.Valid {
				v := int(daysToDeadline.Int32)
				k.DaysToDeadline = &v
			}
			out = append(out, k)
		}
	}
	return out, nil
}

func (r *ReportRepository) ProjectKPI(projectID string) (ProjectKPI, error) {
	all, err := r.ProjectKPIs(DateRange{})
	if err != nil {
		return ProjectKPI{}, err
	}
	for _, p := range all {
		if p.ProjectID == projectID {
			return p, nil
		}
	}
	return ProjectKPI{}, sql.ErrNoRows
}

type DailyPoint struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

type Timeseries struct {
	TasksCompleted []DailyPoint `json:"tasks_completed"`
	TasksCreated   []DailyPoint `json:"tasks_created"`
	UsersJoined    []DailyPoint `json:"users_joined"`
}

func (r *ReportRepository) TimeseriesLast30() (Timeseries, error) {
	var ts Timeseries
	queries := []struct {
		q   string
		dst *[]DailyPoint
	}{
		{`SELECT to_char(d::date, 'YYYY-MM-DD'),
			COALESCE((SELECT COUNT(*) FROM project_tasks WHERE completion_date = d::date AND status = 'завершена'), 0)
		  FROM generate_series(CURRENT_DATE - INTERVAL '29 days', CURRENT_DATE, INTERVAL '1 day') d`, &ts.TasksCompleted},
		{`SELECT to_char(d::date, 'YYYY-MM-DD'),
			COALESCE((SELECT COUNT(*) FROM project_tasks WHERE creation_date = d::date), 0)
		  FROM generate_series(CURRENT_DATE - INTERVAL '29 days', CURRENT_DATE, INTERVAL '1 day') d`, &ts.TasksCreated},
		{`SELECT to_char(d::date, 'YYYY-MM-DD'),
			COALESCE((SELECT COUNT(*) FROM users WHERE registration_date = d::date), 0)
		  FROM generate_series(CURRENT_DATE - INTERVAL '29 days', CURRENT_DATE, INTERVAL '1 day') d`, &ts.UsersJoined},
	}
	for _, qq := range queries {
		rows, err := r.db.Raw(qq.q).Rows()
		if err != nil {
			return ts, err
		}
		points := []DailyPoint{}
		for rows.Next() {
			var p DailyPoint
			if err := rows.Scan(&p.Date, &p.Value); err == nil {
				points = append(points, p)
			}
		}
		rows.Close()
		*qq.dst = points
	}
	return ts, nil
}

type StatusBreakdown struct {
	New       int `json:"new"`
	InProgress int `json:"in_progress"`
	Done      int `json:"done"`
}

func (r *ReportRepository) ProjectStatusBreakdown(projectID string) (StatusBreakdown, error) {
	var b StatusBreakdown
	row := r.db.Raw(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'новая') AS new_count,
			COUNT(*) FILTER (WHERE status = 'в работе') AS in_progress,
			COUNT(*) FILTER (WHERE status = 'завершена') AS done
		FROM project_tasks WHERE project_id = ?`, projectID).Row()
	err := row.Scan(&b.New, &b.InProgress, &b.Done)
	return b, err
}

type ProjectStatusBreakdown struct {
	Active     int `json:"active"`
	Completed  int `json:"completed"`
	Archived   int `json:"archived"`
}

func (r *ReportRepository) GlobalProjectStatusBreakdown() (ProjectStatusBreakdown, error) {
	var b ProjectStatusBreakdown
	row := r.db.Raw(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'активен') AS active,
			COUNT(*) FILTER (WHERE status = 'завершён') AS completed,
			COUNT(*) FILTER (WHERE status = 'архивирован') AS archived
		FROM projects`).Row()
	err := row.Scan(&b.Active, &b.Completed, &b.Archived)
	return b, err
}
