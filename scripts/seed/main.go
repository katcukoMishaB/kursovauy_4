
package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func connect() (*sql.DB, error) {
	host := envOr("DB_HOST", "localhost")
	port := envOr("DB_PORT", "5432")
	user := envOr("DB_USER", "crowdsourcing")
	pass := envOr("DB_PASSWORD", "crowdsourcing_pass")
	dbname := envOr("DB_NAME", "crowdsourcing_db")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)
	return sql.Open("postgres", dsn)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

const password = "password123"

type seedUser struct {
	ID        string
	FirstName string
	Email     string
	Skills    []string
	IsOrg     bool
	IsAdmin   bool
}

func main() {
	db, err := connect()
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v\nПроверь переменные DB_HOST/DB_USER/DB_PASSWORD/DB_NAME.", err)
	}
	if err := waitForDB(db, 30); err != nil {
		log.Fatalf("Postgres не отвечает: %v", err)
	}
	defer db.Close()
	log.Println("Подключение к БД установлено")

	if _, err := os.Stat("migrations"); err == nil {
		if err := applyMigrations(db); err != nil {
			log.Printf("предупреждение: миграции не применились: %v", err)
		} else {
			log.Println("Миграции применены из папки migrations/")
		}
	}

	if !shouldSeed(db) {
		log.Println("ℹ Данные уже загружены, пропускаем (SEED_FORCE=true для пересоздания).")
		return
	}

	if err := wipe(db); err != nil {
		log.Fatalf("Очистка: %v", err)
	}
	log.Println("Старые динамические данные удалены")

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedStr := string(hashed)

	users := []seedUser{
		{FirstName: "Михаил", Email: "admin@unicrowd.ru", IsAdmin: true},
		{FirstName: "Иван", Email: "ivan@unicrowd.ru", IsOrg: true, Skills: []string{"Python", "Базы данных", "Docker"}},
		{FirstName: "Анна", Email: "anna@unicrowd.ru", IsOrg: true, Skills: []string{"Менеджмент", "Копирайтинг", "SMM"}},
		{FirstName: "Сергей", Email: "sergey@unicrowd.ru", IsOrg: true, Skills: []string{"Машинное обучение", "Python", "Анализ данных"}},
		{FirstName: "Екатерина", Email: "ekaterina@unicrowd.ru", Skills: []string{"Веб-вёрстка", "JavaScript", "UX/UI"}},
		{FirstName: "Алексей", Email: "alexey@unicrowd.ru", Skills: []string{"Python", "Анализ данных"}},
		{FirstName: "Дмитрий", Email: "dmitry@unicrowd.ru", Skills: []string{"Java", "Базы данных"}},
		{FirstName: "Ольга", Email: "olga@unicrowd.ru", Skills: []string{"Дизайн", "Figma", "Иллюстрация"}},
		{FirstName: "Павел", Email: "pavel@unicrowd.ru", Skills: []string{"Видеосъёмка", "Монтаж"}},
		{FirstName: "Мария", Email: "maria@unicrowd.ru", Skills: []string{"Маркетинг", "Соцсети"}},
		{FirstName: "Никита", Email: "nikita@unicrowd.ru", Skills: []string{"Java", "Spring"}},
		{FirstName: "Виктория", Email: "viktoria@unicrowd.ru", Skills: []string{"Социология", "Опросы"}},
	}

	for i := range users {
		id, err := createUser(db, users[i].FirstName, users[i].Email, hashedStr, users[i].IsOrg, users[i].IsAdmin)
		if err != nil {
			log.Fatalf("user %s: %v", users[i].Email, err)
		}
		users[i].ID = id

		for _, sk := range users[i].Skills {
			if _, err := db.Exec(
				`INSERT INTO user_skills (user_id, tag_id)
				 SELECT $1, id FROM tag_catalog WHERE lower(name) = lower($2)
				 ON CONFLICT DO NOTHING`,
				id, sk,
			); err != nil {
				log.Fatalf("skill: %v", err)
			}
		}
	}
	log.Printf("✓ Создано %d пользователей (все «Безруков»). Логин: email, пароль: %s", len(users), password)

	cats := loadCategories(db)
	if err := setInterests(db, users, cats); err != nil {
		log.Fatalf("interests: %v", err)
	}
	if err := seedTagCatalog(db, cats); err != nil {
		log.Fatalf("tag_catalog: %v", err)
	}
	log.Println("✓ Каталог тегов и интересы заполнены")

	if err := seedProjects(db, users, cats); err != nil {
		log.Fatalf("projects: %v", err)
	}
	log.Println("✓ Создано 4 проекта с участниками, целями, задачами, чатами")

	if err := backfillActivity(db); err != nil {
		log.Fatalf("activity_log: %v", err)
	}
	log.Println("✓ Activity log заполнен историей за 30 дней")

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("  Seed завершён. Доступы:")
	fmt.Printf("    Админ        : admin@unicrowd.ru     / %s\n", password)
	fmt.Printf("    Организатор  : ivan@unicrowd.ru      / %s\n", password)
	fmt.Printf("    Организатор  : anna@unicrowd.ru      / %s\n", password)
	fmt.Printf("    Организатор  : sergey@unicrowd.ru    / %s\n", password)
	fmt.Printf("    Участник     : ekaterina@unicrowd.ru / %s\n", password)
	fmt.Println("════════════════════════════════════════════════════════")
}

func applyMigrations(db *sql.DB) error {
	dir := "migrations"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("папка migrations не найдена. Запускай скрипт из корня проекта (`go run ./scripts/seed`)")
	}
	files := []string{}
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sql") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(data)); err != nil {
			s := err.Error()
			if strings.Contains(s, "already exists") || strings.Contains(s, "duplicate") {
				continue
			}
			log.Printf("  миграция %s: %v (продолжаем)", filepath.Base(f), err)
		}
	}
	return nil
}

func wipe(db *sql.DB) error {
	tables := []string{
		"project_invitations",
		"activity_log",
		"chat_messages",
		"project_chats",
		"task_comments",
		"project_task_assignees",
		"project_tasks",
		"project_goals",
		"project_tags",
		"project_category_links",
		"tag_catalog",
		"project_participation_requests",
		"project_participations",
		"projects",
		"organizer_requests",
		"user_interests",
		"user_skills",
		"user_roles",
		"users",
	}
	for _, t := range tables {
		if _, err := db.Exec("TRUNCATE TABLE " + t + " RESTART IDENTITY CASCADE"); err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				continue
			}
			return fmt.Errorf("truncate %s: %w", t, err)
		}
	}
	return nil
}

func createUser(db *sql.DB, firstName, email, hashed string, isOrg, isAdmin bool) (string, error) {
	var id string
	regOffset := -1 * (1 + len(firstName))
	err := db.QueryRow(
		`INSERT INTO users (first_name, last_name, email, password, registration_date, status)
		 VALUES ($1, 'Безруков', $2, $3, CURRENT_DATE + $4 * INTERVAL '1 day', true) RETURNING id`,
		firstName, email, hashed, regOffset,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	_, err = db.Exec(
		`INSERT INTO user_roles (user_id, is_participant, is_organizer, is_admin) VALUES ($1, true, $2, $3)`,
		id, isOrg || isAdmin, isAdmin,
	)
	return id, err
}

func loadCategories(db *sql.DB) map[string]string {
	rows, err := db.Query(`SELECT id, name FROM project_categories`)
	if err != nil {
		log.Fatalf("categories: %v", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, name string
		_ = rows.Scan(&id, &name)
		out[name] = id
	}
	if len(out) == 0 {
		base := []string{"Технологии", "Образование", "Социальные проекты", "Наука", "Искусство"}
		for _, n := range base {
			var id string
			_ = db.QueryRow(`INSERT INTO project_categories (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`, n).Scan(&id)
			out[n] = id
		}
	}
	return out
}

func setInterests(db *sql.DB, users []seedUser, cats map[string]string) error {
	mapping := map[string][]string{
		"Иван":      {"Технологии"},
		"Анна":      {"Социальные проекты", "Образование"},
		"Сергей":    {"Технологии", "Наука"},
		"Екатерина": {"Технологии", "Искусство"},
		"Алексей":   {"Технологии", "Наука"},
		"Дмитрий":   {"Технологии"},
		"Ольга":     {"Искусство", "Образование"},
		"Павел":     {"Искусство"},
		"Мария":     {"Социальные проекты"},
		"Никита":    {"Технологии"},
		"Виктория":  {"Социальные проекты", "Наука"},
	}
	for _, u := range users {
		for _, c := range mapping[u.FirstName] {
			if cid, ok := cats[c]; ok {
				if _, err := db.Exec(`INSERT INTO user_interests (user_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, u.ID, cid); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func seedTagCatalog(db *sql.DB, _ map[string]string) error {
	tags := []string{
		"Программирование", "Python", "Java", "Go", "C++", "JavaScript", "TypeScript",
		"Веб-разработка", "Мобильная разработка", "Frontend", "Backend", "Базы данных",
		"DevOps", "Docker", "Kubernetes", "Тестирование", "API", "Spring", "Flask",
		"Машинное обучение", "Анализ данных", "Big Data", "Нейросети", "Компьютерное зрение",
		"NLP", "Статистика",
		"Дизайн", "UX/UI", "Figma", "Иллюстрация", "Графика", "Фотография", "Видеосъёмка",
		"Монтаж", "Анимация", "3D-моделирование",
		"Менеджмент", "Маркетинг", "SMM", "Соцсети", "Копирайтинг", "Журналистика",
		"Волонтёрство", "Социология", "Опросы", "Психология", "Педагогика",
		"Исследование", "Эксперимент", "Публикации", "Гранты", "Биология", "Физика", "Химия",
		"Математика", "Экология",
		"Стартапы", "Бизнес-планирование", "Финансы", "Бухгалтерия", "Юриспруденция",
		"Учебные материалы", "Тьюторство", "Лекции", "Олимпиады", "Профориентация", "Школьники",
		"Спорт", "Городская среда", "Инклюзия", "Здоровье", "Туризм",
	}
	for _, t := range tags {
		if _, err := db.Exec(
			`INSERT INTO tag_catalog (name) VALUES ($1) ON CONFLICT DO NOTHING`,
			t,
		); err != nil {
			return err
		}
	}
	return nil
}

func userByEmail(users []seedUser, email string) seedUser {
	for _, u := range users {
		if u.Email == email {
			return u
		}
	}
	return seedUser{}
}

type taskSpec struct {
	Title       string
	Desc        string
	Priority    string
	Difficulty  int
	Status      string
	AssigneeIdx int
	DueOffset   int
	OnTime      bool
	Quality     int
}

type projectSpec struct {
	Title         string
	ShortDesc     string
	FullDesc      string
	Goal          string
	ImageURL      string
	Category      string
	Status        string
	CreatedAgo    int
	PlannedInDays int
	OrganizerEm   string
	Tags          []string
	ReqSkills     []string
	Members       []memberSpec
	Goals         []goalSpec
	Tasks         []taskSpec
	ChatMessages  []chatMsg
}

type memberSpec struct{ Email, Role string }
type goalSpec struct {
	Title    string
	Desc     string
	Achieved bool
}
type chatMsg struct {
	Email string
	Text  string
	AgoH  int
}

func seedProjects(db *sql.DB, users []seedUser, cats map[string]string) error {
	specs := []projectSpec{
		{
			Title:         "Сайт студенческой жизни",
			ShortDesc:     "Портал с расписанием мероприятий, новостями факультета и личным кабинетом студента.",
			FullDesc:      "Делаем единый сайт, где студенты могут узнать о мероприятиях, записаться на курсы по выбору, посмотреть расписание и оценки. Пишем фронт на простом HTML/JS, бэкенд — на Python+Flask. Первая версия — только для одного факультета, дальше масштабируем.",
			Goal:          "К концу семестра запустить рабочий портал и собрать обратную связь от 100+ студентов.",
			ImageURL:      "https://images.unsplash.com/photo-1523580494863-6f3031224c94?w=800",
			Category:      "Технологии",
			Status:        "активен",
			CreatedAgo:    20,
			PlannedInDays: 50,
			OrganizerEm:   "ivan@unicrowd.ru",
			Tags:          []string{"Веб-разработка", "Базы данных"},
			ReqSkills:     []string{"Python", "Веб-вёрстка", "Базы данных"},
			Members: []memberSpec{
				{"dmitry@unicrowd.ru", "руководитель"},
				{"ekaterina@unicrowd.ru", "заместитель"},
				{"nikita@unicrowd.ru", "участник"},
			},
			Goals: []goalSpec{
				{Title: "Спроектировать структуру БД", Achieved: true},
				{Title: "Сверстать главную страницу", Achieved: true},
				{Title: "Реализовать страницу расписания", Achieved: false},
				{Title: "Запустить тестирование с фокус-группой"},
			},
			Tasks: []taskSpec{
				{"Создать таблицы пользователей и расписания", "users, schedule, courses.", "высокий", 3, "завершена", 0, -15, true, 5},
				{"Сверстать главную страницу", "Только HTML+Tailwind.", "средний", 2, "завершена", 1, -10, true, 4},
				{"Сверстать страницу расписания", "Сетка дней, ячейки занятий.", "средний", 3, "на проверке", 2, 5, false, 0},
				{"Авторизация по email и паролю", "JWT + bcrypt.", "высокий", 4, "в работе", 0, 7, false, 0},
				{"Личный кабинет студента", "Профиль, оценки, выбранные курсы.", "средний", 3, "новая", -1, 14, false, 0},
				{"Раздел новостей", "CRUD для админа.", "низкий", 2, "новая", -1, 21, false, 0},
				{"Тестирование с группой Б-22", "20 студентов, обратная связь.", "средний", 3, "новая", 1, 30, false, 0},
			},
			ChatMessages: []chatMsg{
				{"ivan@unicrowd.ru", "Команда, всем привет! Сегодня старт.", 24 * 20},
				{"dmitry@unicrowd.ru", "Подключился, возьму бэк.", 24 * 20},
				{"ekaterina@unicrowd.ru", "Я займусь вёрсткой и UX.", 24 * 19},
				{"ivan@unicrowd.ru", "Дима, схема БД — за тобой первой.", 24 * 18},
				{"dmitry@unicrowd.ru", "Сделал, скинул в чат.", 24 * 14},
				{"ekaterina@unicrowd.ru", "Главную сверстала, ушла на расписание.", 24 * 9},
			},
		},
		{
			Title:         "Студенческое волонтёрское движение",
			ShortDesc:     "Платформа поиска волонтёров для университетских и городских мероприятий.",
			FullDesc:      "Помогаем организаторам университетских конференций и городских акций находить волонтёров. Студент регистрируется, указывает интересы, получает приглашения. Ведём учёт часов и выдаём электронные сертификаты.",
			Goal:          "Запустить систему до начала весеннего семестра, привлечь 200+ студентов и 30 партнёрских организаций.",
			ImageURL:      "https://images.unsplash.com/photo-1593113598332-cd288d649433?w=800",
			Category:      "Социальные проекты",
			Status:        "активен",
			CreatedAgo:    14,
			PlannedInDays: 80,
			OrganizerEm:   "anna@unicrowd.ru",
			Tags:          []string{"Волонтёрство", "Городская среда"},
			ReqSkills:     []string{"Менеджмент", "SMM", "Соцсети"},
			Members: []memberSpec{
				{"olga@unicrowd.ru", "заместитель"},
				{"viktoria@unicrowd.ru", "руководитель"},
				{"maria@unicrowd.ru", "участник"},
			},
			Goals: []goalSpec{
				{Title: "Согласовать положение с управлением по молодёжной политике", Achieved: true},
				{Title: "Создать брендинг и лендинг", Achieved: true},
				{Title: "Зарегистрировать ≥200 студентов"},
				{Title: "Подключить 30 партнёрских организаций"},
			},
			Tasks: []taskSpec{
				{"Согласовать положение", "Юридическая часть, бонусы в стипендии.", "высокий", 3, "завершена", 1, -7, true, 5},
				{"Создать логотип и стиль", "Логотип, цвета, гайдлайн.", "средний", 3, "завершена", 1, -5, true, 5},
				{"Сверстать лендинг", "Главная страница и FAQ.", "средний", 3, "в работе", 1, 10, false, 0},
				{"Контент-план соцсетей", "VK, Telegram-канал.", "низкий", 2, "новая", -1, 21, false, 0},
				{"Соглашения с 5 первыми партнёрами", "Подписи руководителей.", "высокий", 4, "новая", 2, 30, false, 0},
				{"Опрос студентов", "Google Forms на 500 человек.", "средний", 3, "новая", 3, 18, false, 0},
			},
			ChatMessages: []chatMsg{
				{"anna@unicrowd.ru", "Команда, привет! Старт.", 24 * 14},
				{"viktoria@unicrowd.ru", "Возьму юридическую часть.", 24 * 14},
				{"olga@unicrowd.ru", "Сделаю бренд и лендинг.", 24 * 13},
				{"anna@unicrowd.ru", "Положение подписано 🎉", 24 * 8},
			},
		},
		{
			Title:         "Помощник для абитуриентов",
			ShortDesc:     "Чат-бот, который отвечает на вопросы о поступлении: направления, документы, общежитие.",
			FullDesc:      "Чат-бот для абитуриентов: отвечает на типовые вопросы по направлениям подготовки, проходным баллам, общежитиям и документам. Используем простой поиск по базе из FAQ-вопросов. Цель — закрыть 70% обращений в приёмную комиссию.",
			Goal:          "70% типовых вопросов закрываются ботом без участия приёмной комиссии.",
			ImageURL:      "https://images.unsplash.com/photo-1546410531-bb4caa6b424d?w=800",
			Category:      "Технологии",
			Status:        "активен",
			CreatedAgo:    10,
			PlannedInDays: 60,
			OrganizerEm:   "sergey@unicrowd.ru",
			Tags:          []string{"Машинное обучение", "Анализ данных"},
			ReqSkills:     []string{"Python", "Машинное обучение"},
			Members: []memberSpec{
				{"alexey@unicrowd.ru", "руководитель"},
				{"maria@unicrowd.ru", "заместитель"},
			},
			Goals: []goalSpec{
				{Title: "Собрать датасет вопросов и ответов", Achieved: true},
				{Title: "Подключить телеграм-бот"},
				{Title: "Покрыть 70% сценариев"},
			},
			Tasks: []taskSpec{
				{"Собрать вопросы из реальных писем", "Email-выгрузка, очистка.", "высокий", 4, "завершена", 1, -7, true, 5},
				{"Реализовать поиск по базе", "Простая косинусная похожесть.", "высокий", 4, "завершена", 1, -3, true, 4},
				{"Telegram-bot обёртка", "aiogram, webhook.", "средний", 3, "в работе", 0, 7, false, 0},
				{"Веб-виджет на сайте приёмной", "iframe + API.", "средний", 3, "новая", -1, 21, false, 0},
				{"Тесты с фокус-группой", "20 абитуриентов.", "высокий", 4, "новая", 1, 35, false, 0},
			},
			ChatMessages: []chatMsg{
				{"sergey@unicrowd.ru", "Поехали. Сначала собираем датасет.", 24 * 10},
				{"alexey@unicrowd.ru", "Есть выгрузка email за 2 года, прочищу.", 24 * 9},
				{"sergey@unicrowd.ru", "Сколько уникальных вопросов?", 24 * 6},
				{"alexey@unicrowd.ru", "1240 после кластеризации.", 24 * 6},
			},
		},
		{
			Title:         "Зелёный кампус",
			ShortDesc:     "Раздельный сбор отходов и просветительская программа для студентов.",
			FullDesc:      "Завершённый проект: установили 12 пунктов раздельного сбора, провели 4 лекции и хакатон по экодизайну. Доля раздельной утилизации выросла с 4% до 38%.",
			Goal:          "Не менее 30% от общего объёма отходов отделять для переработки. Провести 3 образовательных мероприятия.",
			ImageURL:      "https://images.unsplash.com/photo-1542601906990-b4d3fb778b09?w=800",
			Category:      "Социальные проекты",
			Status:        "завершён",
			CreatedAgo:    160,
			PlannedInDays: -50,
			OrganizerEm:   "ivan@unicrowd.ru",
			Tags:          []string{"Экология", "Городская среда"},
			ReqSkills:     []string{"Социология"},
			Members: []memberSpec{
				{"viktoria@unicrowd.ru", "руководитель"},
				{"olga@unicrowd.ru", "заместитель"},
				{"alexey@unicrowd.ru", "участник"},
			},
			Goals: []goalSpec{
				{Title: "12 пунктов раздельного сбора", Achieved: true},
				{Title: "Курс лекций по экодизайну", Achieved: true},
				{Title: "Доля раздельной утилизации ≥30%", Achieved: true},
			},
			Tasks: []taskSpec{
				{"Аудит текущей системы вывоза", "Опрос АХЧ.", "высокий", 3, "завершена", 1, -130, true, 5},
				{"Закупка контейнеров", "12 точек, 3 фракции.", "высокий", 4, "завершена", 0, -110, true, 5},
				{"Договор с переработчиком", "Юр. сопровождение.", "высокий", 4, "завершена", 1, -90, true, 4},
				{"Лекция по раздельному сбору", "Поток ИКТ, 200 студентов.", "средний", 2, "завершена", 2, -75, true, 5},
				{"Хакатон 'Эко-дизайн'", "48 часов, 6 команд.", "средний", 3, "завершена", 3, -60, true, 5},
				{"Финальный отчёт ректору", "Метрики, итоги.", "высокий", 3, "завершена", 0, -55, true, 5},
			},
			ChatMessages: []chatMsg{
				{"ivan@unicrowd.ru", "Стартуем эко-проект.", 24 * 160},
				{"viktoria@unicrowd.ru", "Аудит на следующей неделе.", 24 * 158},
				{"ivan@unicrowd.ru", "Контейнеры приехали.", 24 * 110},
				{"viktoria@unicrowd.ru", "Лекция: было 200+ человек.", 24 * 75},
				{"ivan@unicrowd.ru", "Отчёт ректору ушёл, проект закрываем 🎉", 24 * 55},
			},
		},
	}

	for _, sp := range specs {
		if err := seedProject(db, users, cats, sp); err != nil {
			return fmt.Errorf("project %q: %w", sp.Title, err)
		}
	}
	return nil
}

func seedProject(db *sql.DB, users []seedUser, cats map[string]string, sp projectSpec) error {
	organizer := userByEmail(users, sp.OrganizerEm)
	if organizer.ID == "" {
		return fmt.Errorf("organizer not found: %s", sp.OrganizerEm)
	}

	var pid string
	var planned interface{} = nil
	if sp.PlannedInDays != 0 {
		planned = time.Now().AddDate(0, 0, sp.PlannedInDays).Format("2006-01-02")
	}
	var completed interface{} = nil
	if sp.Status == "завершён" {
		completed = time.Now().AddDate(0, 0, sp.PlannedInDays+5).Format("2006-01-02")
	}
	categoryID, _ := cats[sp.Category]

	var imgArg interface{} = nil
	if sp.ImageURL != "" {
		imgArg = sp.ImageURL
	}
	err := db.QueryRow(
		`INSERT INTO projects (organizer_id, title, short_description, full_description,
			goal_description, status, creation_date, planned_end_date, completion_date, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6,
			CURRENT_DATE - $7 * INTERVAL '1 day', $8, $9, $10) RETURNING id`,
		organizer.ID, sp.Title, sp.ShortDesc, sp.FullDesc, sp.Goal, sp.Status,
		sp.CreatedAgo, planned, completed, imgArg,
	).Scan(&pid)
	if err != nil {
		return err
	}

	if _, err := db.Exec(
		`INSERT INTO project_participations (project_id, user_id, role, join_date)
		 VALUES ($1, $2, 'руководитель', CURRENT_DATE - $3 * INTERVAL '1 day') ON CONFLICT DO NOTHING`,
		pid, organizer.ID, sp.CreatedAgo,
	); err != nil {
		return err
	}

	if categoryID != "" {
		_, _ = db.Exec(
			`INSERT INTO project_category_links (project_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			pid, categoryID,
		)
	}

	for _, t := range sp.Tags {
		if _, err := db.Exec(
			`INSERT INTO project_tags (project_id, tag_id, is_required)
			 SELECT $1, id, FALSE FROM tag_catalog WHERE lower(name) = lower($2)
			 ON CONFLICT DO NOTHING`, pid, t); err != nil {
			return err
		}
	}
	for _, t := range sp.ReqSkills {
		if _, err := db.Exec(
			`INSERT INTO project_tags (project_id, tag_id, is_required)
			 SELECT $1, id, TRUE FROM tag_catalog WHERE lower(name) = lower($2)
			 ON CONFLICT (project_id, tag_id) DO UPDATE SET is_required = TRUE`, pid, t); err != nil {
			return err
		}
	}

	memberOrder := []seedUser{organizer}
	for _, m := range sp.Members {
		mu := userByEmail(users, m.Email)
		if mu.ID == "" {
			continue
		}
		if _, err := db.Exec(
			`INSERT INTO project_participations (project_id, user_id, role, join_date)
			 VALUES ($1, $2, $3, CURRENT_DATE - ($4 - 1) * INTERVAL '1 day') ON CONFLICT DO NOTHING`,
			pid, mu.ID, m.Role, sp.CreatedAgo,
		); err != nil {
			return err
		}
		memberOrder = append(memberOrder, mu)
		_, _ = db.Exec(
			`INSERT INTO activity_log (user_id, project_id, action, occurred_at) VALUES ($1, $2, 'project_joined', CURRENT_TIMESTAMP - ($3 - 1) * INTERVAL '1 day')`,
			mu.ID, pid, sp.CreatedAgo,
		)
	}

	for i, g := range sp.Goals {
		var achievedDate interface{} = nil
		if g.Achieved {
			achievedDate = time.Now().AddDate(0, 0, -(sp.CreatedAgo / 2)).Format("2006-01-02")
		}
		if _, err := db.Exec(
			`INSERT INTO project_goals (project_id, title, description, is_achieved, achieved_date, creation_date)
			 VALUES ($1, $2, $3, $4, $5, CURRENT_DATE - $6 * INTERVAL '1 day')`,
			pid, g.Title, nullableStr(g.Desc), g.Achieved, achievedDate, sp.CreatedAgo-i,
		); err != nil {
			return err
		}
	}

	for _, t := range sp.Tasks {
		if err := seedTask(db, pid, memberOrder, t); err != nil {
			return err
		}
	}

	var chatID string
	err = db.QueryRow(
		`INSERT INTO project_chats (project_id, name, creation_date)
		 VALUES ($1, 'Общий чат', CURRENT_DATE - $2 * INTERVAL '1 day') RETURNING id`,
		pid, sp.CreatedAgo,
	).Scan(&chatID)
	if err != nil {
		return err
	}
	for _, m := range sp.ChatMessages {
		mu := userByEmail(users, m.Email)
		if mu.ID == "" {
			continue
		}
		if _, err := db.Exec(
			`INSERT INTO chat_messages (chat_id, user_id, message_text, sent_at)
			 VALUES ($1, $2, $3, CURRENT_TIMESTAMP - $4 * INTERVAL '1 hour')`,
			chatID, mu.ID, m.Text, m.AgoH,
		); err != nil {
			return err
		}
		_, _ = db.Exec(
			`INSERT INTO activity_log (user_id, project_id, action, occurred_at) VALUES ($1, $2, 'message_sent', CURRENT_TIMESTAMP - $3 * INTERVAL '1 hour')`,
			mu.ID, pid, m.AgoH,
		)
	}

	if sp.Status == "активен" {
		applicant := pickNonMember(users, append(memberOrder, organizer))
		if applicant.ID != "" {
			_, _ = db.Exec(
				`INSERT INTO project_participation_requests (project_id, user_id, comment, resume_url, submission_date, status)
				 VALUES ($1, $2, $3, '', CURRENT_DATE - 1 * INTERVAL '1 day', 'в рассмотрении')`,
				pid, applicant.ID,
				"Хочу присоединиться, у меня есть релевантный опыт по теме проекта.",
			)
		}
	}
	return nil
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func seedTask(db *sql.DB, projectID string, members []seedUser, t taskSpec) error {
	if t.AssigneeIdx >= len(members) {
		t.AssigneeIdx = 0
	}
	var assigneeArg interface{} = nil
	var assignee seedUser
	if t.AssigneeIdx >= 0 {
		assignee = members[t.AssigneeIdx]
		assigneeArg = assignee.ID
	}

	var dueArg interface{} = nil
	if t.DueOffset != 0 {
		dueArg = time.Now().AddDate(0, 0, t.DueOffset).Format("2006-01-02")
	}

	var completionArg interface{} = nil
	if t.Status == "завершена" && t.DueOffset != 0 {
		due := time.Now().AddDate(0, 0, t.DueOffset)
		var doneAt time.Time
		if t.OnTime {
			doneAt = due.AddDate(0, 0, -1)
		} else {
			doneAt = due.AddDate(0, 0, 3)
		}
		completionArg = doneAt.Format("2006-01-02")
	} else if t.Status == "завершена" {
		completionArg = time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	}

	var qualityArg interface{} = nil
	if t.Quality > 0 {
		qualityArg = t.Quality
	}

	creationOffset := -25
	if t.DueOffset < 0 {
		creationOffset = t.DueOffset - 5
	}

	var taskID string
	if err := db.QueryRow(
		`INSERT INTO project_tasks (project_id, assigned_to, title, description, status,
			priority, difficulty, quality_rating, creation_date, due_date, completion_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
			CURRENT_DATE + $9 * INTERVAL '1 day', $10, $11) RETURNING id`,
		projectID, assigneeArg, t.Title, t.Desc, t.Status,
		t.Priority, t.Difficulty, qualityArg, creationOffset, dueArg, completionArg,
	).Scan(&taskID); err != nil {
		return err
	}

	_, _ = db.Exec(
		`INSERT INTO activity_log (user_id, project_id, task_id, action, occurred_at) VALUES ($1, $2, $3, 'task_created', CURRENT_TIMESTAMP + $4 * INTERVAL '1 day')`,
		members[0].ID, projectID, taskID, creationOffset,
	)
	if t.Status == "завершена" && completionArg != nil && assignee.ID != "" {
		_, _ = db.Exec(
			`INSERT INTO activity_log (user_id, project_id, task_id, action, occurred_at) VALUES ($1, $2, $3, 'task_completed', $4::date::timestamp)`,
			assignee.ID, projectID, taskID, completionArg,
		)
	}

	if t.Status != "новая" && assignee.ID != "" {
		_, _ = db.Exec(
			`INSERT INTO task_comments (task_id, user_id, content, publication_date)
			 VALUES ($1, $2, 'Принимаю в работу.', CURRENT_DATE + ($3 + 1) * INTERVAL '1 day')`,
			taskID, assignee.ID, creationOffset,
		)
		_, _ = db.Exec(
			`INSERT INTO activity_log (user_id, project_id, task_id, action, occurred_at) VALUES ($1, $2, $3, 'comment_added', CURRENT_TIMESTAMP + ($4 + 1) * INTERVAL '1 day')`,
			assignee.ID, projectID, taskID, creationOffset,
		)
	}
	return nil
}

func pickNonMember(users []seedUser, members []seedUser) seedUser {
	in := func(id string) bool {
		for _, m := range members {
			if m.ID == id {
				return true
			}
		}
		return false
	}
	for _, u := range users {
		if u.IsAdmin {
			continue
		}
		if !in(u.ID) {
			return u
		}
	}
	return seedUser{}
}

func backfillActivity(db *sql.DB) error {
	rows, err := db.Query(`SELECT id FROM users LIMIT 100`)
	if err != nil {
		return err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		_ = rows.Scan(&id)
		ids = append(ids, id)
	}
	for d := 0; d < 30; d++ {
		for k := 0; k < 3; k++ {
			if len(ids) == 0 {
				return nil
			}
			uid := ids[(d+k)%len(ids)]
			_, _ = db.Exec(
				`INSERT INTO activity_log (user_id, action, occurred_at)
				 VALUES ($1, 'login', CURRENT_TIMESTAMP - $2 * INTERVAL '1 day' + $3 * INTERVAL '1 hour')`,
				uid, d, k*4,
			)
		}
	}
	return nil
}

func waitForDB(db *sql.DB, attempts int) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := db.Ping(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(2 * time.Second)
	}
	return lastErr
}

func shouldSeed(db *sql.DB) bool {
	if strings.EqualFold(os.Getenv("SEED_FORCE"), "true") {
		return true
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return true
	}
	return n == 0
}
