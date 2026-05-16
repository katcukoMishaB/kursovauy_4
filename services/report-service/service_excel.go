package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func userTypeRu(t string) string {
	switch t {
	case "teacher":
		return "Преподаватель"
	case "staff":
		return "Сотрудник"
	case "student":
		return "Студент"
	}
	return "Студент"
}

type excelStyles struct {
	titleBar    int
	sectionBar  int
	tableHeader int
	cellBorder  int
	cellAlt     int
	cellNumeric int
}

func buildStyles(f *excelize.File) excelStyles {
	titleBar, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4F46E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
	})
	sectionBar, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"6366F1"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
	})
	tableHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "0F172A"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E0E7FF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "C7D2FE", Style: 1},
			{Type: "top", Color: "C7D2FE", Style: 1},
			{Type: "right", Color: "C7D2FE", Style: 1},
			{Type: "bottom", Color: "818CF8", Style: 2},
		},
	})
	cellBorder, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "0F172A"},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1, WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
		},
	})
	cellAlt, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "0F172A"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F8FAFC"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1, WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
		},
	})
	cellNumeric, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "0F172A"},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
		},
	})
	return excelStyles{titleBar, sectionBar, tableHeader, cellBorder, cellAlt, cellNumeric}
}

func colLetter(i int) string {
	if i < 26 {
		return string(rune('A' + i))
	}
	return string(rune('A'+i/26-1)) + string(rune('A'+i%26))
}

func writeSection(f *excelize.File, sheet, name string, startRow int, headers []string, widths []float64, rows [][]interface{}, numericFrom int, lastCol string, st excelStyles) int {
	if name != "" {
		cell := fmt.Sprintf("A%d", startRow)
		f.SetCellValue(sheet, cell, name)
		f.MergeCell(sheet, cell, lastCol+fmt.Sprintf("%d", startRow))
		f.SetCellStyle(sheet, cell, lastCol+fmt.Sprintf("%d", startRow), st.sectionBar)
		f.SetRowHeight(sheet, startRow, 22)
		startRow++
	}
	for i, h := range headers {
		cell := fmt.Sprintf("%s%d", colLetter(i), startRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, st.tableHeader)
	}
	f.SetRowHeight(sheet, startRow, 26)
	for i, w := range widths {
		if w > 0 {
			letter := colLetter(i)
			cur, _ := f.GetColWidth(sheet, letter)
			if w > cur {
				f.SetColWidth(sheet, letter, letter, w)
			}
		}
	}
	for ri, row := range rows {
		r := startRow + 1 + ri
		for ci, v := range row {
			cell := fmt.Sprintf("%s%d", colLetter(ci), r)
			f.SetCellValue(sheet, cell, v)
			var style int
			if numericFrom >= 0 && ci >= numericFrom {
				style = st.cellNumeric
			} else if ri%2 == 1 {
				style = st.cellAlt
			} else {
				style = st.cellBorder
			}
			f.SetCellStyle(sheet, cell, cell, style)
		}
		f.SetRowHeight(sheet, r, 20)
	}
	return startRow + 1 + len(rows)
}

func writeTitleBar(f *excelize.File, sheet string, lastCol string, title, subtitle string, st excelStyles) int {
	f.SetCellValue(sheet, "A1", title)
	f.MergeCell(sheet, "A1", lastCol+"1")
	f.SetCellStyle(sheet, "A1", lastCol+"1", st.titleBar)
	f.SetRowHeight(sheet, 1, 34)
	if subtitle != "" {
		subStyle, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Italic: true, Size: 10, Color: "475569"},
			Alignment: &excelize.Alignment{Horizontal: "left", Indent: 1, Vertical: "center"},
		})
		f.SetCellValue(sheet, "A2", subtitle)
		f.MergeCell(sheet, "A2", lastCol+"2")
		f.SetCellStyle(sheet, "A2", lastCol+"2", subStyle)
		f.SetRowHeight(sheet, 2, 18)
		return 3
	}
	return 2
}

func dashOrEmpty(s string) string { return s }

func (s *ReportService) BuildKPIExcel(from, to string) ([]byte, string, error) {
	return s.BuildKPIExcelFiltered(from, to, "", "")
}

func (s *ReportService) BuildKPIExcelFiltered(from, to, groupID, userType string) ([]byte, string, error) {
	users, err := s.repo.UserKPIsFiltered("", makeRange(from, to), UserKPIFilter{
		GroupID: groupID, UserType: userType,
	})
	if err != nil {
		return nil, "", err
	}
	projects, err := s.repo.ProjectKPIs(makeRange(from, to))
	if err != nil {
		return nil, "", err
	}
	psb, err := s.repo.GlobalProjectStatusBreakdown()
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	defer f.Close()
	st := buildStyles(f)

	sheet := "KPI"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	period := "за всё время"
	if from != "" || to != "" {
		period = "Период: " + from + " — " + to
	}

	doneTasks, totalTasks := 0, 0
	for _, p := range projects {
		doneTasks += p.TasksCompleted
		totalTasks += p.TasksTotal
	}
	rate := 0.0
	if totalTasks > 0 {
		rate = float64(doneTasks) * 100 / float64(totalTasks)
	}

	row := writeTitleBar(f, sheet, "O", "Сводный KPI-отчёт системы", period, st)

	row = writeSection(f, sheet, "Сводка",
		row,
		[]string{"Проектов всего", "Активных", "Завершённых", "В архиве", "Задач всего", "Выполнено", "Готовность %"},
		[]float64{18, 14, 14, 14, 14, 14, 16},
		[][]interface{}{{
			psb.Active + psb.Completed + psb.Archived,
			psb.Active, psb.Completed, psb.Archived,
			totalTasks, doneTasks, fmt.Sprintf("%.1f", rate),
		}},
		0, "O", st)

	pHeaders := []string{
		"Проект", "Организатор", "Статус", "Создан", "Дедлайн", "Завершён",
		"Участн.", "Активных", "Задач", "Выполн.", "% выполн.", "% в срок",
		"Ср.время (д)", "Ср.качество", "Цели", "Достигн.", "% целей", "Сроки",
	}
	pWidths := []float64{32, 24, 14, 12, 12, 12, 10, 10, 9, 10, 11, 11, 13, 13, 8, 11, 10, 12}
	pRows := make([][]interface{}, 0, len(projects))
	for _, p := range projects {
		planned, completed := "", ""
		if p.PlannedEndDate != nil {
			planned = p.PlannedEndDate.Format("2006-01-02")
		}
		if p.CompletionDate != nil {
			completed = p.CompletionDate.Format("2006-01-02")
		}
		onSched := ""
		if p.OnSchedule != nil {
			if *p.OnSchedule {
				onSched = "в срок"
			} else {
				onSched = "просрочен"
			}
		}
		pRows = append(pRows, []interface{}{
			p.Title, p.OrganizerName, p.Status,
			p.CreationDate.Format("2006-01-02"), planned, completed,
			p.ParticipantsCount, p.ActiveParticipants, p.TasksTotal, p.TasksCompleted,
			fmt.Sprintf("%.1f", p.CompletionRate), fmt.Sprintf("%.1f", p.OnTimeRate),
			fmt.Sprintf("%.1f", p.AvgTaskDays), fmt.Sprintf("%.2f", p.AvgQuality),
			p.GoalsTotal, p.GoalsAchieved, fmt.Sprintf("%.1f", p.GoalsRate), onSched,
		})
	}
	row = writeSection(f, sheet, "KPI проектов", row, pHeaders, pWidths, pRows, 3, "R", st)

	uHeaders := []string{
		"Email", "Имя", "Тип", "Группа", "Назначено", "Выполнено", "В срок", "% в срок",
		"Ср.сложн.", "Ср.качество", "Комм.", "Сообщ.", "Вклад, событий",
		"Акт.проектов", "Акт.дней, д/мес", "Регулярность, %", "Последний раз",
	}
	uWidths := []float64{28, 26, 14, 16, 11, 11, 10, 11, 11, 12, 9, 10, 13, 13, 15, 14, 19}
	uRows := make([][]interface{}, 0, len(users))
	for _, u := range users {
		la := ""
		if u.LastActive != nil {
			la = u.LastActive.Format("2006-01-02 15:04")
		}
		groupName := ""
		if u.GroupName != nil {
			groupName = *u.GroupName
		}
		uRows = append(uRows, []interface{}{
			u.Email, u.FirstName + " " + u.LastName,
			userTypeRu(u.UserType), groupName,
			u.TasksAssigned, u.TasksCompleted, u.TasksOnTime, fmt.Sprintf("%.1f", u.OnTimePercent),
			fmt.Sprintf("%.2f", u.AvgDifficulty), fmt.Sprintf("%.2f", u.AvgQuality),
			u.CommentsCount, u.MessagesCount, u.ActivityScore,
			u.ProjectsActive, u.ActiveDays30, fmt.Sprintf("%.0f", u.RegularityPct), la,
		})
	}
	row = writeSection(f, sheet, "KPI пользователей", row, uHeaders, uWidths, uRows, 2, "Q", st)
	_ = row

	f.SetActiveSheet(0)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	suffix := "all"
	if from != "" || to != "" {
		suffix = from + "_" + to
	}
	return buf.Bytes(), fmt.Sprintf("kpi_report_%s.xlsx", suffix), nil
}

func (s *ReportService) BuildProjectKPIExcel(projectID, userID string, isAdmin bool, from, to string) ([]byte, string, error) {
	kpi, err := s.ProjectKPI(projectID, userID, isAdmin)
	if err != nil {
		return nil, "", err
	}
	users, err := s.repo.UserKPIs(projectID, makeRange(from, to))
	if err != nil {
		return nil, "", err
	}
	users = filterActive(users)

	f := excelize.NewFile()
	defer f.Close()
	st := buildStyles(f)
	period := "за всё время"
	if from != "" || to != "" {
		period = "Период: " + from + " — " + to
	}

	sheet := "Отчёт"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	row := writeTitleBar(f, sheet, "N", "Отчёт по проекту: "+kpi.Title, period, st)

	sumHeaders := []string{
		"Статус", "Создан", "Дедлайн", "Завершён",
		"Участн.", "Активных", "Задач", "Выполн.", "% готов.", "% в срок",
		"Ср.время (д)", "Ср.качество", "Цели", "Достигн.",
	}
	sumWidths := []float64{14, 12, 12, 12, 10, 10, 9, 10, 11, 11, 13, 13, 8, 11}
	planned, completed := "", ""
	if kpi.PlannedEndDate != nil {
		planned = kpi.PlannedEndDate.Format("2006-01-02")
	}
	if kpi.CompletionDate != nil {
		completed = kpi.CompletionDate.Format("2006-01-02")
	}
	sumRows := [][]interface{}{{
		kpi.Status, kpi.CreationDate.Format("2006-01-02"), planned, completed,
		kpi.ParticipantsCount, kpi.ActiveParticipants, kpi.TasksTotal, kpi.TasksCompleted,
		fmt.Sprintf("%.1f", kpi.CompletionRate), fmt.Sprintf("%.1f", kpi.OnTimeRate),
		fmt.Sprintf("%.1f", kpi.AvgTaskDays), fmt.Sprintf("%.2f", kpi.AvgQuality),
		kpi.GoalsTotal, kpi.GoalsAchieved,
	}}
	row = writeSection(f, sheet, "Сводка по проекту", row, sumHeaders, sumWidths, sumRows, 4, "N", st)

	teamHeaders := []string{
		"Email", "Имя", "Назначено", "Выполнено", "В срок", "% в срок",
		"Ср.сложн.", "Ср.качество", "Комм.", "Сообщ.", "Вклад, событий",
		"Акт.дней, д/мес", "Регулярность, %", "Последний раз",
	}
	teamWidths := []float64{28, 26, 11, 11, 10, 11, 11, 12, 9, 10, 13, 14, 14, 19}
	teamRows := make([][]interface{}, 0, len(users))
	for _, u := range users {
		la := ""
		if u.LastActive != nil {
			la = u.LastActive.Format("2006-01-02 15:04")
		}
		teamRows = append(teamRows, []interface{}{
			u.Email, u.FirstName + " " + u.LastName,
			u.TasksAssigned, u.TasksCompleted, u.TasksOnTime, fmt.Sprintf("%.1f", u.OnTimePercent),
			fmt.Sprintf("%.2f", u.AvgDifficulty), fmt.Sprintf("%.2f", u.AvgQuality),
			u.CommentsCount, u.MessagesCount, u.ActivityScore,
			u.ActiveDays30, fmt.Sprintf("%.0f", u.RegularityPct), la,
		})
	}
	row = writeSection(f, sheet, "KPI команды", row, teamHeaders, teamWidths, teamRows, 2, "N", st)
	_ = row

	f.SetActiveSheet(0)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	suffix := "all"
	if from != "" || to != "" {
		suffix = from + "_" + to
	}
	short := projectID
	if len(short) > 8 {
		short = short[:8]
	}
	return buf.Bytes(), fmt.Sprintf("project_kpi_%s_%s.xlsx", short, suffix), nil
}

var _ = dashOrEmpty
